package gremcos

import (
	"bytes"
	"fmt"
	"go.uber.org/atomic"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supplyon/gremcos/interfaces"
	mock_interfaces "github.com/supplyon/gremcos/test/mocks/interfaces"
)

func TestNewWebsocket(t *testing.T) {
	// GIVEN

	// WHEN
	websocket, err := NewWebsocket("ws://localhost")

	// THEN
	assert.NotNil(t, websocket)
	assert.NoError(t, err)
}

func TestNewWebsocketFail(t *testing.T) {

	// WHEN - invalid host
	websocket, err := NewWebsocket("invalid host")
	assert.Nil(t, websocket)
	assert.Error(t, err)

	// WHEN - read buffer invalid
	websocket, err = NewWebsocket("ws://host", SetBufferSize(0, 10))
	assert.Nil(t, websocket)
	assert.Error(t, err)

	// WHEN - write buffer invalid
	websocket, err = NewWebsocket("ws://host", SetBufferSize(10, 0))
	assert.Nil(t, websocket)
	assert.Error(t, err)

	// WHEN - websocketFactory is nil
	websocket, err = NewWebsocket("ws://host", websocketDialerFactoryFun(nil))
	assert.Nil(t, websocket)
	assert.Error(t, err)
}

func TestConnect(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)

	websocket, err := NewWebsocket("ws://localhost", websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, websocket)

	// WHEN
	mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())
	err = websocket.Connect()

	// THEN
	assert.NoError(t, err)
	assert.True(t, websocket.IsConnected())
}



func TestConnectFail(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, true)

	socket, err := NewWebsocket("ws://localhost", websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, socket)

	// WHEN
	err = socket.Connect()

	// THEN
	assert.Error(t, err)
	assert.False(t, socket.IsConnected())
	assert.False(t, socket.(*websocket).connected.Load())
}

func TestConnectReconnect(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)

	dialer, err := NewWebsocket("ws://localhost", websocketDialerFactoryFun(mockedDialerFactory))
	websocket := dialer.(*websocket)
	require.NoError(t, err)
	require.NotNil(t, websocket)

	// WHEN - first connect successful
	mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())
	err = websocket.Connect()

	// THEN
	assert.NoError(t, err)
	assert.True(t, websocket.IsConnected())
	assert.True(t, websocket.connected.Load())


	// WHEN - second connect fails
	websocket.wsDialerFactory = newMockedDialerFactory(mockedWebsocketConnection, true)
	err = websocket.Connect()

	// THEN
	assert.Error(t, err)
	assert.False(t, websocket.IsConnected())
	assert.False(t, websocket.connected.Load())
}

func TestConnectClose(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)

	websocket, err := NewWebsocket("ws://localhost", websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, websocket)

	// WHEN connected
	mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())
	err = websocket.Connect()
	require.NoError(t, err)

	mockedWebsocketConnection.EXPECT().WriteMessage(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, "")).Return(nil)
	mockedWebsocketConnection.EXPECT().Close()
	err = websocket.Close()

	// THEN
	assert.NoError(t, err)
}

func TestConnectCloseOnNotConnectedWebsocket(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)

	websocket, err := NewWebsocket("ws://localhost", websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, websocket)

	// WHEN - not connected
	err = websocket.Close()

	// THEN
	assert.NoError(t, err)
}

func TestPing(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	websocket := &websocket{
		conn:      mockedWebsocketConnection,
		connected: atomic.NewBool(true),
	}

	// WHEN
	mockedWebsocketConnection.EXPECT().WriteControl(gorilla.PingMessage, gomock.Any(), gomock.Any()).Return(nil)
	err := websocket.Ping()

	// THEN
	assert.NoError(t, err)
	assert.True(t, websocket.IsConnected())
}

func TestPingFail(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	websocket := &websocket{
		conn:      mockedWebsocketConnection,
		connected: atomic.NewBool(true),
	}

	// WHEN
	mockedWebsocketConnection.EXPECT().WriteControl(gorilla.PingMessage, gomock.Any(), gomock.Any()).Return(fmt.Errorf("ERROR"))
	err := websocket.Ping()

	// THEN
	assert.Error(t, err)
	assert.False(t, websocket.IsConnected())
	assert.False(t, websocket.connected.Load())

}

func TestPingFailWhenNotConnected(t *testing.T) {
	// GIVEN
	websocket := &websocket{connected: atomic.NewBool(false)}

	// WHEN
	err := websocket.Ping()

	// THEN
	assert.Error(t, err)
	assert.False(t, websocket.IsConnected())
}

func TestWrite(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	websocket := &websocket{
		conn:      mockedWebsocketConnection,
		connected: atomic.NewBool(true),
	}
	data := []byte("hello")

	// WHEN
	mockedWebsocketConnection.EXPECT().SetWriteDeadline(gomock.Any()).Return(nil)
	mockedWebsocketConnection.EXPECT().WriteMessage(gorilla.BinaryMessage, data).Return(nil)
	err := websocket.Write(data)

	// THEN
	assert.NoError(t, err)
}

func TestRead(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	websocket := &websocket{
		conn:      mockedWebsocketConnection,
		connected: atomic.NewBool(true),
	}
	data := []byte("hello")
	datalen := len(data)

	// WHEN
	mockedWebsocketConnection.EXPECT().SetReadDeadline(gomock.Any()).Return(nil)
	mockedWebsocketConnection.EXPECT().ReadMessage().Return(datalen, data, nil)
	nBytes, dataReceived, err := websocket.Read()

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, data, dataReceived)
	assert.Equal(t, datalen, nBytes)
}

func TestMultiReconnectAndParallelRead(t *testing.T) {
	// This test shall ensure that there are no race conditions when creating and reading from a connection.
	// Hence, it should be run with -race.
	// The test reconnects multiple times and checks for the connection in parallel.

	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)

	websocket, err := NewWebsocket("ws://localhost", websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, websocket)

	// WHEN - multiple reconnects and parallel checks
	mockedWebsocketConnection.EXPECT().SetReadDeadline(gomock.Any()).AnyTimes()
	mockedWebsocketConnection.EXPECT().SetWriteDeadline(gomock.Any()).AnyTimes()
	mockedWebsocketConnection.EXPECT().WriteControl(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockedWebsocketConnection.EXPECT().ReadMessage().AnyTimes()
	mockedWebsocketConnection.EXPECT().WriteMessage(gomock.Any(), gomock.Any()).AnyTimes()

	quitChannel := make(chan struct{})
	// parallel checker
	go func() {
		for {
			websocket.IsConnected()
			select {
			case <-quitChannel:
				return
			default:
				continue
			}
		}
	}()
	// parallel read, ping write
	go func() {
		for {
			err = websocket.Ping()
			_, _, err = websocket.Read()
			err = websocket.Write([]byte("HUHU"))
			_ = err
			select {
			case <-quitChannel:
				return
			default:
				continue
			}
		}
	}()

	for i := 0; i < 100; i++ {
		mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())
		require.NoError(t, websocket.Connect())
		require.True(t, websocket.IsConnected())
	}
	close(quitChannel)
}

func newMockedDialerFactory(websocketConnection interfaces.WebsocketConnection, fail bool) websocketDialerFactory {

	websocketFuncSuccess := func(urlStr string, requestHeader http.Header) (interfaces.WebsocketConnection, *http.Response, error) {
		return websocketConnection, nil, nil
	}

	websocketFuncError := func(urlStr string, requestHeader http.Header) (interfaces.WebsocketConnection, *http.Response, error) {
		return nil, nil, fmt.Errorf("Timeout")
	}

	// if needed return a websocket that can't create a connection successfully
	if fail {
		return func(wBufSize, rBifSize int, timeout time.Duration) websocketDialer {
			return websocketFuncError
		}
	}

	return func(wBufSize, rBifSize int, timeout time.Duration) websocketDialer {
		return websocketFuncSuccess
	}
}

func TestExtractConnectionError(t *testing.T) {
	assert.Nil(t, extractConnectionError(nil))

	resp := &http.Response{}
	assert.Error(t, extractConnectionError(resp))

	resp.Body = ioutil.NopCloser(bytes.NewReader([]byte("")))
	assert.Error(t, extractConnectionError(resp))

	resp.Body = ioutil.NopCloser(bytes.NewReader([]byte("hello")))
	assert.Error(t, extractConnectionError(resp))
}

func TestConcurrentWriteAndCloseOnConnection(t *testing.T) {
	// This test shall ensure that there are no race conditions when writing and closing a connection.
	// Hence, it should be run with -race.

	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)

	websocket, err := NewWebsocket("ws://localhost", websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, websocket)

	// WHEN - multiple reconnects and parallel checks
	mockedWebsocketConnection.EXPECT().SetReadDeadline(gomock.Any()).AnyTimes()
	mockedWebsocketConnection.EXPECT().SetWriteDeadline(gomock.Any()).AnyTimes()
	mockedWebsocketConnection.EXPECT().WriteControl(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockedWebsocketConnection.EXPECT().ReadMessage().AnyTimes()
	mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())

	sending := false

	mockedWebsocketConnection.EXPECT().WriteMessage(gomock.Any(),gomock.Any()).MinTimes(1).Do(func(dataType interface{},data interface{}) error {
		require.False(t,sending)
		sending = true
		time.Sleep(time.Millisecond*500)
		sending = false
		return nil
	})
	mockedWebsocketConnection.EXPECT().Close().MinTimes(1).Do(func() error {
		require.False(t,sending)
		sending = true
		time.Sleep(time.Millisecond*1)
		sending = false
		return nil
	})

	err = websocket.Connect()
	require.NoError(t, err)

	go func() {
			_ = websocket.Write([]byte("HUHU"))
	}()

	time.Sleep(time.Millisecond*50)
	websocket.Close()

	time.Sleep(time.Millisecond*50)

}