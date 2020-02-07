package gremtune

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	gorilla "github.com/gorilla/websocket"
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
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_connection.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)

	dialer, err := NewDialer("ws://localhost", websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, dialer)

	// WHEN
	mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())
	err = dialer.connect()

	// THEN
	assert.NoError(t, err)
	assert.True(t, dialer.IsConnected())
}

func TestConnectFail(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_connection.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, true)

	dialer, err := NewDialer("ws://localhost", websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, dialer)

	// WHEN
	err = dialer.connect()

	// THEN
	assert.Error(t, err)
	assert.False(t, dialer.IsConnected())
}

func TestConnectReconnect(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_connection.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)

	dialer, err := NewDialer("ws://localhost", websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, dialer)

	// WHEN - first connect successful
	mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())
	err = dialer.connect()

	// THEN
	assert.NoError(t, err)
	assert.True(t, dialer.IsConnected())

	// WHEN - second connect fails
	dialer.(*websocket).wsDialerFactory = newMockedDialerFactory(mockedWebsocketConnection, true)
	err = dialer.connect()

	// THEN
	assert.Error(t, err)
	assert.False(t, dialer.IsConnected())
}

func TestConnectClose(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_connection.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)

	dialer, err := NewDialer("ws://localhost", websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, dialer)

	// WHEN connected
	mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())
	err = dialer.connect()
	require.NoError(t, err)

	mockedWebsocketConnection.EXPECT().WriteMessage(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, "")).Return(nil)
	mockedWebsocketConnection.EXPECT().Close()
	err = dialer.close()

	// THEN
	assert.NoError(t, err)
}

func TestConnectCloseOnNotConnectedWebsocket(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_connection.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)

	dialer, err := NewDialer("ws://localhost", websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, dialer)

	// WHEN - not connected
	err = dialer.close()

	// THEN
	assert.NoError(t, err)
}

func TestConnectCloseFail(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_connection.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)

	dialer, err := NewDialer("ws://localhost", websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, dialer)

	// WHEN connected
	mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())
	err = dialer.connect()
	require.NoError(t, err)

	mockedWebsocketConnection.EXPECT().WriteMessage(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, "")).Return(nil)
	mockedWebsocketConnection.EXPECT().Close()
	err = dialer.close()
	require.NoError(t, err)

	// WHEN close is called again on a disposed websocket
	err = dialer.close()

	// THEN
	assert.Error(t, err)
}

func newMockedDialerFactory(websocketConnection WebsocketConnection, fail bool) websocketDialerFactory {

	dialerFuncSuccess := func(urlStr string, requestHeader http.Header) (WebsocketConnection, *http.Response, error) {
		return websocketConnection, nil, nil
	}

	dialerFuncError := func(urlStr string, requestHeader http.Header) (WebsocketConnection, *http.Response, error) {
		return nil, nil, fmt.Errorf("Timeout")
	}

	// if needed return a dialer that can't create a connection successfully
	if fail {
		return func(wBufSize, rBifSize int, timeout time.Duration) websocketDialer {
			return dialerFuncError
		}
	}

	return func(wBufSize, rBifSize int, timeout time.Duration) websocketDialer {
		return dialerFuncSuccess
	}
}
