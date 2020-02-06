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

	// WHEN - invalid host
	dialer, err := NewDialer("invalid host")
	assert.Nil(t, dialer)
	assert.Error(t, err)

	// WHEN - read buffer invalid
	dialer, err = NewDialer("ws://host", SetBufferSize(0, 10))
	assert.Nil(t, dialer)
	assert.Error(t, err)

	// WHEN - write buffer invalid
	dialer, err = NewDialer("ws://host", SetBufferSize(10, 0))
	assert.Nil(t, dialer)
	assert.Error(t, err)

	// WHEN - dialerFactory is nil
	dialer, err = NewDialer("ws://host", websocketDialerFactoryFun(nil))
	assert.Nil(t, dialer)
	assert.Error(t, err)
}

func TestPanicOnMissingAuthCredentials(t *testing.T) {
	c := newClient()
	ws := &websocket{}
	c.conn = ws

	defer func() {
		if r := recover(); r == nil {
			t.Fail()
		}
	}()

	c.conn.getAuth()
}
