package gremcos

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supplyon/gremcos/interfaces"
)

type myNetError string // implements net.Error interface

func (mne myNetError) Error() string {
	return fmt.Sprintf("net.Error: %s", string(mne))
}

func (mne myNetError) Timeout() bool {
	return false
}

func (mne myNetError) Temporary() bool {
	return false
}

func TestIsNetError(t *testing.T) {
	// GIVEN
	netErr := myNetError("failed")
	noNetErr := fmt.Errorf("failed")

	// WHEN
	isNetErr := isNetError(netErr)
	isNoNetErr := isNetError(noNetErr)
	isNilErr := isNetError(nil)

	// THEN
	assert.True(t, isNetErr)
	assert.False(t, isNoNetErr)
	assert.False(t, isNilErr)
}

func TestIsNetworkError(t *testing.T) {
	// GIVEN
	netErr := myNetError("failed")
	noNetErr := fmt.Errorf("failed")
	noConnectionError := ErrNoConnection
	connectivityError := Error{Wrapped: fmt.Errorf("Conn failed"), Category: ErrorCategoryConnectivity}
	noConnectivityError := Error{Wrapped: fmt.Errorf("some failure"), Category: ErrorCategoryAuth}

	// WHEN
	isNetErr := IsNetworkErr(netErr)
	isNoNwError := IsNetworkErr(noNetErr)
	isNilError := IsNetworkErr(nil)
	isNoConnectionError := IsNetworkErr(noConnectionError)
	isConnectivityError := IsNetworkErr(connectivityError)
	isNoConnectivityError := IsNetworkErr(noConnectivityError)

	// THEN
	assert.True(t, isNetErr)
	assert.False(t, isNoNwError)
	assert.False(t, isNilError)
	assert.True(t, isNoConnectionError)
	assert.True(t, isConnectivityError)
	assert.False(t, isNoConnectivityError)
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
		dummyResponse := interfaces.Response{
			RequestID: "",
			Status:    interfaces.Status{Code: co.code},
			Result:    interfaces.Result{},
		}
		err := extractError(dummyResponse)
		switch {
		case co.code == 200:
			assert.NoError(t, err, "Successful response returned error (code %d).", co.code)
		case co.code == 204:
			assert.NoError(t, err, "Successful response returned error (code %d).", co.code)
		case co.code == 206:
			assert.NoError(t, err, "Successful response returned error (code %d).", co.code)
		case co.code == 3434:
			assert.Error(t, err, "Unsuccessful response did not return error (code %d).", co.code)
		default:
			require.Error(t, err, "Unsuccessful response did not return error (code %d).", co.code)
			cerr, ok := err.(Error)
			assert.True(t, ok)
			assert.NotEqual(t, ErrorCategoryGeneral, cerr.Category) // should only be general for unknown errors
			assert.NotNil(t, cerr.Wrapped)
		}
	}
}
