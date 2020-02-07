package gremtune

import (
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	mock_interfaces "github.com/schwartzmx/gremtune/test/mocks/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRequestPreparation tests the ability to package a query and a set of bindings into a request struct for further manipulation
func TestRequestPreparation(t *testing.T) {
	query := "g.V(x)"
	bindings := map[string]string{"x": "10"}
	rebindings := map[string]string{}
	req, id, err := prepareRequestWithBindings(query, bindings, rebindings)
	require.NoError(t, err)

	expectedRequest := request{
		RequestID: id,
		Op:        "eval",
		Processor: "",
		Args: map[string]interface{}{
			"gremlin":    query,
			"bindings":   bindings,
			"language":   "gremlin-groovy",
			"rebindings": rebindings,
		},
	}

	assert.Equal(t, req, expectedRequest)
}

// TestRequestPackaging tests the ability for gremtune to format a request using the established Gremlin Server WebSockets protocol for delivery to the server
func TestRequestPackaging(t *testing.T) {
	testRequest := request{
		RequestID: "1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
		Op:        "eval",
		Processor: "",
		Args: map[string]interface{}{
			"gremlin":  "g.V(x)",
			"bindings": map[string]string{"x": "10"},
			"language": "gremlin-groovy",
		},
	}

	msg, err := packageRequest(testRequest)
	require.NoError(t, err)

	j, err := json.Marshal(testRequest)
	require.NoError(t, err)

	var expected []byte

	mimetype := []byte("application/vnd.gremlin-v2.0+json")
	mimetypelen := byte(len(mimetype))

	expected = append(expected, mimetypelen)
	expected = append(expected, mimetype...)
	expected = append(expected, j...)

	assert.Equal(t, msg, expected)
}

// TestRequestDispatch tests the ability for a requester to send a request to the client for writing to Gremlin Server
func TestRequestDispatch(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	testRequest := request{
		RequestID: "1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
		Op:        "eval",
		Processor: "",
		Args: map[string]interface{}{
			"gremlin":  "g.V(x)",
			"bindings": map[string]string{"x": "10"},
			"language": "gremlin-groovy",
		},
	}
	c := newClient(mockedDialer)
	msg, err := packageRequest(testRequest)
	require.NoError(t, err)

	// WHEN
	c.dispatchRequest(msg)
	// c.requests is the channel where all requests are sent for writing
	// to Gremlin Server, write workers listen on this channel
	req := <-c.requests

	// THEN
	assert.Equal(t, msg, req)
}

// TestAuthRequestDispatch tests the ability for a requester to send a request to the client for writing to Gremlin Server
func TestAuthRequestDispatch(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	id := "1d6d02bd-8e56-421d-9438-3bd6d0079ff1"
	testRequest, _ := prepareAuthRequest(id, "test", "root")

	c := newClient(mockedDialer)
	msg, err := packageRequest(testRequest)
	require.NoError(t, err)

	// WHEN
	c.dispatchRequest(msg)
	// c.requests is the channel where all requests are sent for writing
	// to Gremlin Server, write workers listen on this channel
	req := <-c.requests
	// THEN
	assert.Equal(t, msg, req)
}

// TestAuthRequestPreparation tests the ability to create successful authentication request
func TestAuthRequestPreparation(t *testing.T) {
	id := "1d6d02bd-8e56-421d-9438-3bd6d0079ff1"
	testRequest, err := prepareAuthRequest(id, "test", "root")
	require.NoError(t, err)

	assert.Equal(t, testRequest.RequestID, id)
	assert.Equal(t, "trasversal", testRequest.Processor)
	assert.Equal(t, "authentication", testRequest.Op)

	assert.Len(t, testRequest.Args, 1)
	assert.NotEmpty(t, testRequest.Args["sasl"])
	assert.Equal(t, "AHRlc3QAcm9vdA==", testRequest.Args["sasl"])
}

func TestPrepareRequest(t *testing.T) {
	query := "g.V()"
	testRequest, id, err := prepareRequest(query)
	require.NoError(t, err)

	assert.NotEmpty(t, id)
	assert.Equal(t, "", testRequest.Processor)
	assert.Equal(t, "eval", testRequest.Op)

	assert.Len(t, testRequest.Args, 2)
	assert.NotEmpty(t, testRequest.Args["language"])
	assert.Equal(t, "gremlin-groovy", testRequest.Args["language"])
	assert.NotEmpty(t, testRequest.Args["gremlin"])
	assert.Equal(t, query, testRequest.Args["gremlin"])
}
