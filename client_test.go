package gremtune

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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

func TestNewClient(t *testing.T) {
	// WHEN
	client := newClient()

	// THEN
	require.NotNil(t, client)

}
