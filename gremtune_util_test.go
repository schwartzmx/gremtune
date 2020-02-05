package gremtune

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, errChan chan error) *Client {
	dialer := NewDialer("ws://127.0.0.1:8182")
	require.NotNil(t, dialer, "Dialer is nil")
	client, err := Dial(dialer, errChan)
	require.NoError(t, err, "Failed to create client")
	return &client
}

func newTestPool(t *testing.T, errChan chan error) *Pool {
	dialFn := func() (*Client, error) {
		dialer := NewDialer("ws://127.0.0.1:8182")
		c, err := Dial(dialer, errChan)
		require.NoError(t, err)
		return &c, err
	}

	return &Pool{
		Dial:        dialFn,
		MaxActive:   10,
		IdleTimeout: time.Duration(10 * time.Second),
	}
}
