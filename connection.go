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

	// auth auth information like username and password
	auth interfaces.Auth

	// disposed flags the websocket as
	// 'has been closed and can't be reused'
	disposed bool

	// connected flags the websocket as connected or not connected
	connected bool

	// pingInterval is the interval that is used to check if the connection
	// is still alive
	pingInterval time.Duration

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

	// channel for quit notification
	quit chan struct{}
	mux  sync.RWMutex

	// wsDialerFactory is a factory that creates
	// dialers (functions that can establish a websocket connection)
	wsDialerFactory websocketDialerFactory
}

// NewDialer returns a WebSocket dialer to use when connecting to Gremlin Server
func NewDialer(host string, configs ...DialerConfig) (interfaces.Dialer, error) {
	createdWebsocket := &websocket{
		timeout:         5 * time.Second,
		pingInterval:    60 * time.Second,
		writingWait:     15 * time.Second,
		readingWait:     15 * time.Second,
		connected:       false,
		quit:            make(chan struct{}),
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

	return createdWebsocket, nil
}

func (ws *websocket) Connect() error {
	if ws.disposed {
		return fmt.Errorf("This websocket is already disposed (closed). Websockets can't be reused connect() -> close() -> connect() is not permitted")
	}

	// create the function that shall be used for dialing
	dial := ws.wsDialerFactory(ws.writeBufSize, ws.readBufSize, ws.timeout)

	conn, _, err := dial(ws.host, http.Header{})
	ws.conn = conn
	if err != nil {
		ws.connected = false
		// As of 3.2.2 the URL has changed.
		// https://groups.google.com/forum/#!msg/gremlin-users/x4hiHsmTsHM/Xe4GcPtRCAAJ
		// Probably '/gremlin' has to be added to the used hostname
		return fmt.Errorf("Dial failed: %s. Probably '/gremlin' has to be added to the used hostname", err)
	}

	ws.conn.SetPongHandler(func(appData string) error {
		ws.connected = true
		return nil
	})

	ws.connected = true
	return nil
}

func (ws *websocket) GetQuitChannel() <-chan struct{} {
	return ws.quit
}

// IsConnected returns whether the underlying websocket is connected
func (ws *websocket) IsConnected() bool {
	ws.mux.RLock()
	defer ws.mux.RUnlock()

	return ws.connected && ws.conn != nil
}

// IsDisposed returns whether the underlying websocket is disposed
func (ws *websocket) IsDisposed() bool {
	return ws.disposed
}

func (ws *websocket) Write(msg []byte) error {
	return ws.conn.WriteMessage(2, msg)
}

func (ws *websocket) Read() (msgType int, msg []byte, err error) {
	return ws.conn.ReadMessage()
}

// close closes the websocket
// Caution!: After calling this function the whole websocket is invalid
// since the internal quit channel is also closed and won't be recreated.
// Hence after closing a websocket one has to create a new one instead of
// reusing the closed one and call connect on it.
// Caution!: This method can only called once each second call will result in an error.
func (ws *websocket) Close() error {
	if ws.disposed {
		return fmt.Errorf("This websocket is already disposed (closed). Websockets can't be reused close() -> close() is not permitted")
	}

	// clean up in any case
	defer func() {
		// close the channel to send the quit notification
		// to all workers
		close(ws.quit)
		if ws.conn != nil {
			ws.conn.Close()
		}
		ws.disposed = true
	}()

	if !ws.IsConnected() {
		return nil
	}
	return ws.conn.WriteMessage(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, "")) //Cleanly close the connection with the server
}

func (ws *websocket) GetAuth() interfaces.Auth {
	return ws.auth
}

func (ws *websocket) Ping(errs chan error) {
	ticker := time.NewTicker(ws.pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			connected := true
			if err := ws.conn.WriteControl(gorilla.PingMessage, []byte{}, time.Now().Add(ws.writingWait)); err != nil {
				errs <- err
				connected = false
			}
			ws.mux.Lock()
			ws.connected = connected
			ws.mux.Unlock()

		case <-ws.quit:
			return
		}
	}
}
