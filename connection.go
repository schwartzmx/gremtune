package gremtune

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"sync"

	gorilla "github.com/gorilla/websocket"
)

// WebsocketConnection is the minimal interface needed to act on a websocket
type WebsocketConnection interface {
	SetPongHandler(handler func(appData string) error)
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
	WriteControl(messageType int, data []byte, deadline time.Time) error
}

// websocket is the dialer for a WebsocketConnection
type websocket struct {
	host         string
	conn         WebsocketConnection
	auth         *auth
	disposed     bool
	connected    bool
	pingInterval time.Duration
	writingWait  time.Duration
	readingWait  time.Duration

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
func NewDialer(host string, configs ...DialerConfig) (dialer, error) {
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

func (ws *websocket) connect() error {
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

func (ws *websocket) write(msg []byte) error {
	return ws.conn.WriteMessage(2, msg)
}

func (ws *websocket) read() (msgType int, msg []byte, err error) {
	return ws.conn.ReadMessage()
}

// close closes the websocket
// Caution!: After calling this function the whole websocket is invalid
// since the internal quit channel is also closed and won't be recreated.
// Hence after closing a websocket one has to create a new one instead of
// reusing the closed one and call connect on it.
// Caution!: This method can only called once each second call will result in an error.
func (ws *websocket) close() error {
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

func (ws *websocket) getAuth() *auth {
	if ws.auth == nil {
		panic("You must create a Secure Dialer for authenticate with the server")
	}
	return ws.auth
}

func (ws *websocket) ping(errs chan error) {
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
