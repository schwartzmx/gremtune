package gremtune

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mock_connection "github.com/schwartzmx/gremtune/test/mocks/connection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestConnect(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_connection.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection)

	dialer, err := NewDialer("ws://localhost", websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, dialer)

}

func newMockedDialerFactory(websocketConnection WebsocketConnection) websocketDialerFactory {

	dialerFunc := func(urlStr string, requestHeader http.Header) (WebsocketConnection, *http.Response, error) {
		return websocketConnection, nil, nil
	}

	return func(wBufSize, rBifSize int, timeout time.Duration) websocketDialer {
		return dialerFunc
	}
}
