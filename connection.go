package gremcos

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"sync"

	"github.com/supplyon/gremcos/interfaces"

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

	mux sync.RWMutex

	// wsDialerFactory is a factory that creates
	// dialers (functions that can establish a websocket connection)
	wsDialerFactory websocketDialerFactory
}

// NewWebsocket returns a WebSocket dialer to use when connecting to Gremlin Server
func NewWebsocket(host string, options ...optionWebsocket) (interfaces.Dialer, error) {
	createdWebsocket := &websocket{
		timeout:         5 * time.Second,
		writingWait:     15 * time.Second,
		readingWait:     15 * time.Second,
		connected:       false,
		readBufSize:     8192,
		writeBufSize:    8192,
		host:            host,
		wsDialerFactory: gorillaWebsocketDialerFactory, // use the gorilla websocket as default
	}

	for _, opt := range options {
		opt(createdWebsocket)
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

	return createdWebsocket, nil
}

// Connect connects to the peer and actually opens the connection.
// This function has to be called before writing/ reading from/ to the socket.
func (ws *websocket) Connect() error {

	// create the function that shall be used for dialing
	dial := ws.wsDialerFactory(ws.writeBufSize, ws.readBufSize, ws.timeout)

	conn, response, err := dial(ws.host, http.Header{})
	if err != nil {
		ws.setConnection(nil)

		errMsg := fmt.Sprintf("dialing '%s' failed with %s. Probably '/gremlin' has to be added to the used hostname.", ws.host, err)
		// try to get some additional information out of the response
		errMsgAdditional := ""
		if err = extractConnectionError(response); err != nil {
			errMsgAdditional = fmt.Sprintf(" Details: %s", err.Error())
		}

		// As of 3.2.2 the URL has changed.
		// https://groups.google.com/forum/#!msg/gremlin-users/x4hiHsmTsHM/Xe4GcPtRCAAJ
		// Probably '/gremlin' has to be added to the used hostname
		return fmt.Errorf("%s%s", errMsg, errMsgAdditional)
	}

	// Install the handler for pong messages from the peer.
	// As stated in the documentation (see :https://github.com/gorilla/websocket/blob/master/conn.go#L1156)
	// the handler has usually to do nothing except of reading the connection.
	// This is one of two parts of the websockets heartbeet protocol.
	conn.SetPongHandler(func(appData string) error {
		return nil
	})

	ws.setConnection(conn)
	return nil
}

func extractConnectionError(resp *http.Response) error {
	if resp == nil {
		return nil
	}
	errMinimal := fmt.Errorf("%s", resp.Status)

	if resp.Body == nil {
		return errMinimal
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return errMinimal
	}

	dataStr := string(data)
	if dataStr == "" {
		return errMinimal
	}
	return fmt.Errorf("%s: %s", resp.Status, dataStr)
}

func (ws *websocket) setConnection(connection interfaces.WebsocketConnection) {
	ws.mux.Lock()
	defer ws.mux.Unlock()
	ws.conn = connection
}

// IsConnected returns whether the underlying WebsocketConnection is connected or not
func (ws *websocket) IsConnected() bool {
	ws.mux.RLock()
	defer ws.mux.RUnlock()
	return ws.conn != nil
}

// Write writes the given data chunk on the socket
func (ws *websocket) Write(msg []byte) error {
	if !ws.IsConnected() {
		return fmt.Errorf("Not connected")
	}

	// ensure that we have the connection during the whole read operation
	ws.mux.RLock()
	defer ws.mux.RUnlock()

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
	if !ws.IsConnected() {
		return 0, nil, fmt.Errorf("Not connected")
	}

	// ensure that we have the connection during the whole read operation
	ws.mux.RLock()
	defer ws.mux.RUnlock()

	// ensure to not block forever
	if err := ws.conn.SetReadDeadline(time.Now().Add(ws.readingWait)); err != nil {
		return 0, nil, err
	}
	return ws.conn.ReadMessage()
}

// Close closes the underlying websocket
func (ws *websocket) Close() error {

	if !ws.IsConnected() {
		return nil
	}

	// clean up in any case
	defer func() {
		if ws.conn != nil {
			ws.conn.Close()
		}
	}()

	// ensure that we have the connection during the whole read operation
	ws.mux.RLock()
	defer ws.mux.RUnlock()

	//Cleanly close the connection with the server
	return ws.conn.WriteMessage(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, ""))
}

// Ping sends a websocket ping frame to the peer.
// This is one of two parts of the websockets heartbeet protocol.
// It has to be ensured that somebody calls this function continuously (e.g. each 60s).
// Otherwise the socket will be closed by the peer.
func (ws *websocket) Ping() error {
	if !ws.IsConnected() {
		return fmt.Errorf("Not connected")
	}

	// ensure that we have the connection during the whole read operation
	disconnected := false
	ws.mux.RLock()
	err := ws.conn.WriteControl(gorilla.PingMessage, []byte{}, time.Now().Add(ws.writingWait))
	if err != nil {
		disconnected = true
	}
	ws.mux.RUnlock()

	if disconnected {
		ws.setConnection(nil)
	}
	return err
}
