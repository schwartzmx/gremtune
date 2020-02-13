package gremcos

import (
	"net/http"
	"time"

	gorilla "github.com/gorilla/websocket"
	"github.com/supplyon/gremcos/interfaces"
)

// websocketDialer is a function type for dialing/ connecting to a websocket server and creating a WebsocketConnection
type websocketDialer func(urlStr string, requestHeader http.Header) (interfaces.WebsocketConnection, *http.Response, error)

// websocketDialerFactory is a function type that is able to create websocketDialer's
type websocketDialerFactory func(writeBufferSize, readBufferSize int, handshakeTimout time.Duration) websocketDialer

// gorillaWebsocketDialerFactory is a function that is able to create websocketDialer's using the websocket implementation
// of github.com/gorilla/websocket
var gorillaWebsocketDialerFactory = func(writeBufferSize, readBufferSize int, handshakeTimout time.Duration) websocketDialer {
	// create the gorilla websocket dialer
	dialer := gorilla.Dialer{
		WriteBufferSize:  writeBufferSize,
		ReadBufferSize:   readBufferSize,
		HandshakeTimeout: handshakeTimout,
	}

	// return the websocketDialer, wrapping the gorilla websocket dial call
	return func(urlStr string, requestHeader http.Header) (interfaces.WebsocketConnection, *http.Response, error) {
		return dialer.Dial(urlStr, requestHeader)
	}
}
