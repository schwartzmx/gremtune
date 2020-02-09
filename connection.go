package gremtune

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"sync"

	"github.com/schwartzmx/gremtune/interfaces"

	gorilla "github.com/gorilla/websocket"
)

// websocket is the dialer for a WebsocketConnection
type websocket struct {
	// the host to establish the connection with
	// it is expected to specify the protocol as part of the host
	// supported protocols are ws and wss
	// example: ws://localhost:8182/gremlin
	host string

	// conn is the actual connection
	conn interfaces.WebsocketConnection

	// disposed flags the websocket as
	// 'has been closed and can't be reused'
	disposed bool

	// connected flags the websocket as connected or not connected
	connected bool

	// writingWait is the maximum time a write operation will wait to start
	// sending data on the socket. If this duration has been exceeded
	// the operation will fail with an error.
	writingWait time.Duration

	// readingWait is the maximum time a read operation will wait until
	// data is received on the socket. If this duration has been exceeded
	// the operation will fail with an error.
	readingWait time.Duration

	// timeout for the initial handshake
	timeout time.Duration

	readBufSize  int
	writeBufSize int

	// quitChannel channel for quit notification
	quitChannel chan struct{}
	mux         sync.RWMutex

	// wsDialerFactory is a factory that creates
	// dialers (functions that can establish a websocket connection)
	wsDialerFactory websocketDialerFactory
}

// NewDialer returns a WebSocket dialer to use when connecting to Gremlin Server
func NewDialer(host string, configs ...DialerConfig) (interfaces.Dialer, error) {
	createdWebsocket := &websocket{
		timeout:         5 * time.Second,
		writingWait:     15 * time.Second,
		readingWait:     15 * time.Second,
		connected:       false,
		disposed:        false,
		quitChannel:     make(chan struct{}),
		readBufSize:     8192,
		writeBufSize:    8192,
		host:            host,
		wsDialerFactory: gorillaWebsocketDialerFactory, // use the gorilla websocket as default
	}

	for _, conf := range configs {
		conf(createdWebsocket)
	}

	// verify setup and fail as early as possible
	if !strings.HasPrefix(createdWebsocket.host, "ws://") && !strings.HasPrefix(createdWebsocket.host, "wss://") {
		return nil, fmt.Errorf("Host '%s' is invalid, expected protocol 'ws://' or 'wss://' missing", createdWebsocket.host)
	}

	if createdWebsocket.readBufSize <= 0 {
		return nil, fmt.Errorf("Invalid size for read buffer: %d", createdWebsocket.readBufSize)
	}

	if createdWebsocket.writeBufSize <= 0 {
		return nil, fmt.Errorf("Invalid size for write buffer: %d", createdWebsocket.writeBufSize)
	}

	if createdWebsocket.wsDialerFactory == nil {
		return nil, fmt.Errorf("The factory for websocket dialers is nil")
	}

	// TODO: Check if it makes more sense to call Connect() at this point.
	// Anyway Connect() should only be called once. Hence it could be refactored
	// to an internal function which is called right when the dialer is created.
	// But probably one want's to create a dialer and then want's to do a delayed connect (e.g. pool)?!
	// This allows the separation of dialer:= NewDialer() and Dial(dialer)
	return createdWebsocket, nil
}

// Connect connects to the peer and actually opens the connection.
// This function has to be called before writing/ reading from/ to the socket.
// This function should not be called if the websocket is already disposed.
// In case of an error it is safer to just create a new dialer via NewDialer
func (ws *websocket) Connect() error {
	if ws.disposed {
		return fmt.Errorf("This websocket is already disposed (closed). Websockets can't be reused connect() -> close() -> connect() is not permitted")
	}

	// create the function that shall be used for dialing
	dial := ws.wsDialerFactory(ws.writeBufSize, ws.readBufSize, ws.timeout)

	conn, response, err := dial(ws.host, http.Header{})
	ws.conn = conn
	if err != nil {
		ws.setConnected(false)

		errMsg := fmt.Sprintf("Dial failed: %s. Probably '/gremlin' has to be added to the used hostname.", err)
		// try to get some additional information out of the response
		if response != nil {
			errMsg = fmt.Sprintf("%s - Response from Server %d", errMsg, response.StatusCode)
		}

		// As of 3.2.2 the URL has changed.
		// https://groups.google.com/forum/#!msg/gremlin-users/x4hiHsmTsHM/Xe4GcPtRCAAJ
		// Probably '/gremlin' has to be added to the used hostname
		return fmt.Errorf("%s", errMsg)
	}

	// Install the handler for pong messages from the peer.
	// As stated in the documentation (see :https://github.com/gorilla/websocket/blob/master/conn.go#L1156)
	// the handler has usually to do nothing except of reading the connection.
	// Here we update additionally the connection state to connected.
	// This is one of two parts of the websockets heartbeet protocol.
	ws.conn.SetPongHandler(func(appData string) error {
		ws.setConnected(true)
		return nil
	})

	ws.setConnected(true)
	return nil
}

func (ws *websocket) setConnected(connected bool) {
	ws.mux.Lock()
	defer ws.mux.Unlock()
	ws.connected = connected
}

// GetQuitChannel returns the channel where a quit messages is send as soon as the underlying WebsocketConnection
// has been closed.
func (ws *websocket) GetQuitChannel() <-chan struct{} {
	return ws.quitChannel
}

// IsConnected returns whether the underlying WebsocketConnection is connected or not
func (ws *websocket) IsConnected() bool {
	ws.mux.RLock()
	defer ws.mux.RUnlock()
	return ws.connected && ws.conn != nil
}

// IsDisposed returns whether the underlying websocket is disposed or not.
// Disposed websockets are dead, they can't be reused by calling Connect() again.
func (ws *websocket) IsDisposed() bool {
	return ws.disposed
}

// Write writes the given data chunk on the socket
func (ws *websocket) Write(msg []byte) error {

	// ensure to not block forever
	if err := ws.conn.SetWriteDeadline(time.Now().Add(ws.writingWait)); err != nil {
		return err
	}

	return ws.conn.WriteMessage(gorilla.BinaryMessage, msg)
}

// Read reads data from the websocket.
// Supported message types, are:
// - gorilla.TextMessage
// - gorilla.BinaryMessage
// - gorilla.CloseMessage
// - gorilla.PingMessage
// - gorilla.PongMessage
func (ws *websocket) Read() (messageType int, msg []byte, err error) {

	// ensure to not block forever
	if err := ws.conn.SetReadDeadline(time.Now().Add(ws.readingWait)); err != nil {
		return 0, nil, err
	}

	return ws.conn.ReadMessage()
}

// Close closes the websocket
// Caution!: After calling this function the whole websocket is invalid/ dead
// since the internal quit channel is also closed and won't be recreated.
// Hence after closing a websocket one has to create a new one instead of
// reusing the closed one and call connect on it.
// Caution!: This method can only called once. Each second call will result in an error.
func (ws *websocket) Close() error {
	if ws.disposed {
		return fmt.Errorf("This websocket is already disposed (closed). Websockets can't be reused close() -> close() is not permitted")
	}

	// clean up in any case
	defer func() {
		// close the channel to send the quit notification
		// to all workers
		close(ws.quitChannel)
		if ws.conn != nil {
			ws.conn.Close()
		}
		ws.disposed = true
	}()

	if !ws.IsConnected() {
		return nil
	}
	//Cleanly close the connection with the server
	return ws.conn.WriteMessage(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, ""))
}

// Ping sends a websocket ping frame to the peer.
// This is one of two parts of the websockets heartbeet protocol.
// It has to be ensured that somebody calls this function continuously (e.g. each 60s).
// Otherwise the socket will be closed by the peer.
func (ws *websocket) Ping() error {
	if ws.conn == nil {
		return fmt.Errorf("Not connected")
	}

	connected := true
	err := ws.conn.WriteControl(gorilla.PingMessage, []byte{}, time.Now().Add(ws.writingWait))
	if err != nil {
		connected = false
	}

	ws.setConnected(connected)
	return err
}
