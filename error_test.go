package gremcos

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
	connectivityError := ErrConnectivity{Wrapped: fmt.Errorf("Conn failed")}

	// WHEN
	isNetErr := IsNetworkErr(netErr)
	isNoNwError := IsNetworkErr(noNetErr)
	isNilError := IsNetworkErr(nil)
	isNoConnectionError := IsNetworkErr(noConnectionError)
	isConnectivityError := IsNetworkErr(connectivityError)

	// THEN
	assert.True(t, isNetErr)
	assert.False(t, isNoNwError)
	assert.False(t, isNilError)
	assert.True(t, isNoConnectionError)
	assert.True(t, isConnectivityError)
}
