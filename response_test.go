package gremcos

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supplyon/gremcos/interfaces"
	mock_interfaces "github.com/supplyon/gremcos/test/mocks/interfaces"
)

// Dummy responses for mocking
var dummySuccessfulResponse = []byte(`{"result":{"data":[{"id": 2,"label": "person","type": "vertex","properties": [
  {"id": 2, "value": "vadas", "label": "name"},
  {"id": 3, "value": 27, "label": "age"}]}
  ], "meta":{}},
 "requestId":"1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
 "status":{"code":200,"attributes":{},"message":""}}`)

var dummyNeedAuthenticationResponse = []byte(`{"result":{},
 "requestId":"1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
 "status":{"code":407,"attributes":{},"message":""}}`)

var dummySuccessfulResponseMarshalled = interfaces.Response{
	RequestID: "1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
	Status:    interfaces.Status{Code: 200, Attributes: map[string]interface{}{}},
	Result: interfaces.Result{Data: []byte(`[{"id": 2,"label": "person","type": "vertex","properties": [
	  {"id": 2, "value": "vadas", "label": "name"},
	  {"id": 3, "value": 27, "label": "age"}]}
	]`), Meta: map[string]interface{}{}},
}

var dummyNeedAuthenticationResponseMarshalled = interfaces.Response{
	RequestID: "1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
	Status:    interfaces.Status{Code: 407},
	Result:    interfaces.Result{Data: []byte("")},
}

var dummyPartialResponse1Marshalled = interfaces.Response{
	RequestID: "1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
	Status:    interfaces.Status{Code: 206}, // Code 206 indicates that the response is not the terminating response in a sequence of responses
	Result:    interfaces.Result{Data: []byte("testPartialData1")},
}

var dummyPartialResponse2Marshalled = interfaces.Response{
	RequestID: "1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
	Status:    interfaces.Status{Code: 200},
	Result:    interfaces.Result{Data: []byte("testPartialData2")},
}

// TestResponseHandling tests the overall response handling mechanism of gremcos
func TestResponseHandling(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	c := newClient(mockedDialer)

	err := c.handleResponse(dummySuccessfulResponse)
	require.NoError(t, err)

	var expected []interfaces.Response
	expected = append(expected, dummySuccessfulResponseMarshalled)

	r, err := c.retrieveResponse(dummySuccessfulResponseMarshalled.RequestID)
	require.NoError(t, err)

	assert.Equal(t, reflect.TypeOf(r), reflect.TypeOf(expected))
}

func TestAuthRequested(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	c := newClient(mockedDialer, SetAuth(StaticCredentialProvider{
		UsernameStatic: "username",
		PasswordStatic: "password",
	}))

	// WHEN
	err := c.handleResponse(dummyNeedAuthenticationResponse)

	// THEN
	require.NoError(t, err)
}

func TestPrepareAuthenRequest(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	c := newClient(mockedDialer)

	req := prepareAuthRequest(dummyNeedAuthenticationResponseMarshalled.RequestID, "test", "test")

	sampleAuthRequest, err := packageRequest(req)
	require.NoError(t, err)

	c.dispatchRequest(sampleAuthRequest)
	authRequest := <-c.requests //Simulate that client send auth challenge to server
	assert.Equal(t, authRequest, sampleAuthRequest, "Expected data type does not match actual.")
}

func TestAuthCompleted(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	c := newClient(mockedDialer)

	err := c.handleResponse(dummySuccessfulResponse) //If authentication is successful the server returns the origin petition
	require.NoError(t, err)

	var expectedSuccessful []interfaces.Response
	expectedSuccessful = append(expectedSuccessful, dummySuccessfulResponseMarshalled)

	response, err := c.retrieveResponse(dummySuccessfulResponseMarshalled.RequestID)
	require.NoError(t, err)

	assert.Equal(t, reflect.TypeOf(expectedSuccessful), reflect.TypeOf(response), "Expected data type does not match actual.")
}

// TestResponseMarshalling tests the ability to marshal a response into a designated response struct for further manipulation
func TestResponseMarshalling(t *testing.T) {
	resp, err := marshalResponse(dummySuccessfulResponse)
	require.NoError(t, err)

	assert.Equal(t, resp.RequestID, dummySuccessfulResponseMarshalled.RequestID)
	assert.Equal(t, dummySuccessfulResponseMarshalled.Status.Code, resp.Status.Code)
	assert.Equal(t, reflect.TypeOf(resp.Result.Data).String(), "json.RawMessage")
}

// TestResponseSortingSingleResponse tests the ability for sortResponse to save a response received from Gremlin Server
func TestResponseSortingSingleResponse(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	c := newClient(mockedDialer)

	c.saveResponse(dummySuccessfulResponseMarshalled, nil)

	var expected []interface{}
	expected = append(expected, dummySuccessfulResponseMarshalled)

	// WHEN
	result, ok := c.results.Load(dummySuccessfulResponseMarshalled.RequestID)

	// THEN
	assert.True(t, ok)
	assert.Equal(t, expected, result.(interface{}))
}

// TestResponseSortingMultipleResponse tests the ability for the sortResponse function to categorize and group responses that are sent in a stream
func TestResponseSortingMultipleResponse(t *testing.T) {

	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	c := newClient(mockedDialer)

	// WHEN
	c.saveResponse(dummyPartialResponse1Marshalled, nil)
	c.saveResponse(dummyPartialResponse2Marshalled, nil)

	var expected []interface{}
	expected = append(expected, dummyPartialResponse1Marshalled)
	expected = append(expected, dummyPartialResponse2Marshalled)

	result, ok := c.results.Load(dummyPartialResponse1Marshalled.RequestID)

	// THEN
	assert.True(t, ok)
	assert.Equal(t, expected, result.([]interface{}))
}

// TestResponseRetrieval tests the ability for a requester to retrieve the response for a specified requestId generated when sending the request
func TestResponseRetrieval(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	c := newClient(mockedDialer)

	c.saveResponse(dummyPartialResponse1Marshalled, nil)
	c.saveResponse(dummyPartialResponse2Marshalled, nil)

	resp, err := c.retrieveResponse(dummyPartialResponse1Marshalled.RequestID)
	require.NoError(t, err)

	var expected []interfaces.Response
	expected = append(expected, dummyPartialResponse1Marshalled)
	expected = append(expected, dummyPartialResponse2Marshalled)

	assert.Equal(t, resp, expected)
}

func TestResponseRetrievalFail(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	c := newClient(mockedDialer)

	resp, err := c.retrieveResponse("nonexistent response")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestResponseDeletion tests the ability for a requester to clean up after retrieving a response after delivery to a client
func TestResponseDeletion(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	c := newClient(mockedDialer)

	c.saveResponse(dummyPartialResponse1Marshalled, nil)
	_, ok := c.results.Load(dummyPartialResponse1Marshalled.RequestID)
	assert.True(t, ok)

	c.saveResponse(dummyPartialResponse2Marshalled, nil)
	_, ok = c.results.Load(dummyPartialResponse1Marshalled.RequestID)
	assert.True(t, ok)

	// WHEN
	c.deleteResponse(dummyPartialResponse1Marshalled.RequestID)

	// THEN
	_, ok = c.results.Load(dummyPartialResponse1Marshalled.RequestID)
	assert.False(t, ok)
}

func TestAsyncResponseRetrieval(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	c := newClient(mockedDialer)

	c.saveResponse(dummyPartialResponse1Marshalled, nil)
	c.saveResponse(dummyPartialResponse2Marshalled, nil)

	responseChannel := make(chan interfaces.AsyncResponse, 10)
	c.retrieveResponseAsync(dummyPartialResponse1Marshalled.RequestID, responseChannel)

	resp := <-responseChannel
	expectedAsync := interfaces.AsyncResponse{Response: dummyPartialResponse1Marshalled}
	assert.Equal(t, expectedAsync, resp)

	resp = <-responseChannel
	expectedAsync = interfaces.AsyncResponse{Response: dummyPartialResponse2Marshalled}
	assert.Equal(t, expectedAsync, resp)
}

func TestEmptyIfNilOrError(t *testing.T) {
	assert.Empty(t, emptyIfNilOrError(nil))
	assert.Equal(t, "failure", emptyIfNilOrError(fmt.Errorf("failure")))
}
