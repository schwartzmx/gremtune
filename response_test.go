package gremtune

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Dummy responses for mocking
var dummySuccessfulResponse = []byte(`{"result":{"data":[{"id": 2,"label": "person","type": "vertex","properties": [
  {"id": 2, "value": "vadas", "label": "name"},
  {"id": 3, "value": 27, "label": "age"}]}
  ], "meta":{}},
 "requestId":"1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
 "status":{"code":200,"attributes":{},"message":""}}`)

//var dummyNeedAuthenticationResponse = []byte(`{"result":{},
// "requestId":"1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
// "status":{"code":407,"attributes":{},"message":""}}`)

var dummySuccessfulResponseMarshalled = Response{
	RequestID: "1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
	Status:    Status{Code: 200},
	Result:    Result{Data: []byte("testData")},
}

//var dummyNeedAuthenticationResponseMarshalled = Response{
//	RequestID: "1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
//	Status:    Status{Code: 407},
//	Result:    Result{Data: []byte("")},
//}

//var dummyPartialResponse1Marshalled = Response{
//	RequestID: "1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
//	Status:    Status{Code: 206}, // Code 206 indicates that the response is not the terminating response in a sequence of responses
//	Result:    Result{Data: []byte("testPartialData1")},
//}

//var dummyPartialResponse2Marshalled = Response{
//	RequestID: "1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
//	Status:    Status{Code: 200},
//	Result:    Result{Data: []byte("testPartialData2")},
//}

// TestResponseHandling tests the overall response handling mechanism of gremtune
func TestResponseHandling(t *testing.T) {
	//c := newClient()
	//
	//err := c.handleResponse(dummySuccessfulResponse)
	//require.NoError(t, err)
	//
	//var expected []Response
	//expected = append(expected, dummySuccessfulResponseMarshalled)
	//
	//r, err := c.retrieveResponse(dummySuccessfulResponseMarshalled.RequestID)
	//require.NoError(t, err)
	//
	//assert.Equal(t, reflect.TypeOf(r), reflect.TypeOf(expected))
}

func TestResponseAuthHandling(t *testing.T) {
	//c := newClient()
	//ws := &websocket{}
	//ws.auth = &auth{username: "test", password: "test"}
	//c.conn = ws
	//err := c.handleResponse(dummyNeedAuthenticationResponse)
	//require.NoError(t, err)
	//
	//req, err := prepareAuthRequest(dummyNeedAuthenticationResponseMarshalled.RequestID, "test", "test")
	//require.NoError(t, err)
	//
	//sampleAuthRequest, err := packageRequest(req)
	//require.NoError(t, err)
	//
	//c.dispatchRequest(sampleAuthRequest)
	//authRequest := <-c.requests //Simulate that client send auth challenge to server
	//assert.Equal(t, authRequest, sampleAuthRequest, "Expected data type does not match actual.")
	//
	//err = c.handleResponse(dummySuccessfulResponse) //If authentication is successful the server returns the origin petition
	//require.NoError(t, err)
	//
	//var expectedSuccessful []Response
	//expectedSuccessful = append(expectedSuccessful, dummySuccessfulResponseMarshalled)
	//
	//r, err := c.retrieveResponse(dummySuccessfulResponseMarshalled.RequestID)
	//require.NoError(t, err)
	//assert.Equal(t, reflect.TypeOf(expectedSuccessful), reflect.TypeOf(r), "Expected data type does not match actual.")
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

	//c := newClient()
	//
	//c.saveResponse(dummySuccessfulResponseMarshalled, nil)
	//
	//var expected []interface{}
	//expected = append(expected, dummySuccessfulResponseMarshalled)
	//
	//result, ok := c.results.Load(dummySuccessfulResponseMarshalled.RequestID)
	//assert.True(t, ok)
	//
	//if reflect.DeepEqual(result.([]interface{}), expected) != true {
	//	t.Fail()
	//}
}

// TestResponseSortingMultipleResponse tests the ability for the sortResponse function to categorize and group responses that are sent in a stream
func TestResponseSortingMultipleResponse(t *testing.T) {

	//c := newClient()
	//
	//c.saveResponse(dummyPartialResponse1Marshalled, nil)
	//c.saveResponse(dummyPartialResponse2Marshalled, nil)
	//
	//var expected []interface{}
	//expected = append(expected, dummyPartialResponse1Marshalled)
	//expected = append(expected, dummyPartialResponse2Marshalled)
	//
	//results, ok := c.results.Load(dummyPartialResponse1Marshalled.RequestID)
	//assert.True(t, ok)
	//if reflect.DeepEqual(results.([]interface{}), expected) != true {
	//	t.Fail()
	//}
}

// TestResponseRetrieval tests the ability for a requester to retrieve the response for a specified requestId generated when sending the request
func TestResponseRetrieval(t *testing.T) {
	//c := newClient()
	//
	//c.saveResponse(dummyPartialResponse1Marshalled, nil)
	//c.saveResponse(dummyPartialResponse2Marshalled, nil)
	//
	//resp, err := c.retrieveResponse(dummyPartialResponse1Marshalled.RequestID)
	//require.NoError(t, err)
	//
	//var expected []Response
	//expected = append(expected, dummyPartialResponse1Marshalled)
	//expected = append(expected, dummyPartialResponse2Marshalled)
	//
	//assert.Equal(t, resp, expected)
}

// TestResponseDeletion tests the ability for a requester to clean up after retrieving a response after delivery to a client
func TestResponseDeletion(t *testing.T) {
	//c := newClient()
	//
	//c.saveResponse(dummyPartialResponse1Marshalled, nil)
	//c.saveResponse(dummyPartialResponse2Marshalled, nil)
	//
	//c.deleteResponse(dummyPartialResponse1Marshalled.RequestID)
	//
	//_, ok := c.results.Load(dummyPartialResponse1Marshalled.RequestID)
	//assert.False(t, ok)
}

var codes = []struct {
	code int
}{
	{200},
	{204},
	{206},
	{401},
	{407},
	{498},
	{499},
	{500},
	{597},
	{598},
	{599},
	{3434}, // Testing unknown error code
}

// Tests detection of errors and if an error is generated for a specific error code
func TestResponseErrorDetection(t *testing.T) {
	for _, co := range codes {
		dummyResponse := Response{
			RequestID: "",
			Status:    Status{Code: co.code},
			Result:    Result{},
		}
		err := dummyResponse.detectError()
		switch {
		case co.code == 200:
			if err != nil {
				t.Log("Successful response returned error.")
			}
		case co.code == 204:
			if err != nil {
				t.Log("Successful response returned error.")
			}
		case co.code == 206:
			if err != nil {
				t.Log("Successful response returned error.")
			}
		default:
			if err == nil {
				t.Log("Unsuccessful response did not return error.")
			}
		}
	}
}
