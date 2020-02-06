package gremtune

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDialer(t *testing.T) {
	dialer, err := NewDialer("ws://localhost")
	assert.NotNil(t, dialer)
	assert.NoError(t, err)
}

func TestNewDialerFail(t *testing.T) {

	dialer, err := NewDialer("invalid host")
	assert.Nil(t, dialer)
	assert.Error(t, err)
}

func TestPanicOnMissingAuthCredentials(t *testing.T) {
	c := newClient()
	ws := &Websocket{}
	c.conn = ws

	defer func() {
		if r := recover(); r == nil {
			t.Fail()
		}
	}()

	c.conn.getAuth()
}
