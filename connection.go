package gremtune

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"sync"

	gorilla "github.com/gorilla/websocket"
	"github.com/pkg/errors"
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
	sync.RWMutex

	wsDialerFactory websocketDialerFactory
}

//Auth is the container for authentication data of dialer
type auth struct {
	username string
	password string
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

func (ws *websocket) connect() (err error) {
	// create the function that shall be used for dialing
	dial := ws.wsDialerFactory(ws.writeBufSize, ws.readBufSize, ws.timeout)

	ws.conn, _, err = dial(ws.host, http.Header{})
	if err != nil {

		// As of 3.2.2 the URL has changed.
		// https://groups.google.com/forum/#!msg/gremlin-users/x4hiHsmTsHM/Xe4GcPtRCAAJ
		ws.host = ws.host + "/gremlin"
		ws.conn, _, err = dial(ws.host, http.Header{})

		if err != nil {
			return err
		}
	}

	if err == nil {
		ws.connected = true
		ws.conn.SetPongHandler(func(appData string) error {
			ws.connected = true
			return nil
		})
	}
	return nil
}

// IsConnected returns whether the underlying websocket is connected
func (ws *websocket) IsConnected() bool {
	return ws.connected
}

// IsDisposed returns whether the underlying websocket is disposed
func (ws *websocket) IsDisposed() bool {
	return ws.disposed
}

func (ws *websocket) write(msg []byte) (err error) {
	err = ws.conn.WriteMessage(2, msg)
	return
}

func (ws *websocket) read() (msgType int, msg []byte, err error) {
	msgType, msg, err = ws.conn.ReadMessage()
	return
}

func (ws *websocket) close() (err error) {
	defer func() {
		close(ws.quit)
		ws.conn.Close()
		ws.disposed = true
	}()

	err = ws.conn.WriteMessage(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, "")) //Cleanly close the connection with the server
	return
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
			ws.Lock()
			ws.connected = connected
			ws.Unlock()

		case <-ws.quit:
			return
		}
	}
}

func (c *Client) writeWorker(errs chan error, quit chan struct{}) { // writeWorker works on a loop and dispatches messages as soon as it receives them
	for {
		select {
		case msg := <-c.requests:
			c.Lock()
			err := c.conn.write(msg)
			if err != nil {
				errs <- err
				c.Errored = true
				c.Unlock()
				break
			}
			c.Unlock()

		case <-quit:
			return
		}
	}
}

func (c *Client) readWorker(errs chan error, quit chan struct{}) { // readWorker works on a loop and sorts messages as soon as it receives them
	for {
		msgType, msg, err := c.conn.read()
		if msgType == -1 { // msgType == -1 is noFrame (close connection)
			return
		}
		if err != nil {
			errs <- errors.Wrapf(err, "Receive message type: %d", msgType)
			c.Errored = true
			break
		}
		if msg != nil {
			// FIXME: At the moment the error returned by handle response is just ignored.
			err = c.handleResponse(msg)
			_ = err
		}

		select {
		case <-quit:
			return
		default:
			continue
		}
	}
}
