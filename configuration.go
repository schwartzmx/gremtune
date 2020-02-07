package gremtune

import (
	"time"

	"github.com/schwartzmx/gremtune/interfaces"
)

//DialerConfig is the struct for defining configuration for WebSocket dialer
type DialerConfig func(*websocket)

//SetAuthentication sets on dialer credentials for authentication
func SetAuthentication(username string, password string) DialerConfig {
	return func(ws *websocket) {
		ws.auth = &interfaces.Auth{Username: username, Password: password}
	}
}

//SetTimeout sets the dial handshake timeout
func SetTimeout(timeout time.Duration) DialerConfig {
	return func(ws *websocket) {
		ws.timeout = timeout
	}
}

//SetPingInterval sets the interval of ping sending for know is
//connection is alive and in consequence the client is connected
func SetPingInterval(interval time.Duration) DialerConfig {
	return func(ws *websocket) {
		ws.pingInterval = interval
	}
}

//SetWritingWait sets the time for waiting that writing occur
func SetWritingWait(wait time.Duration) DialerConfig {
	return func(ws *websocket) {
		ws.writingWait = wait
	}
}

//SetReadingWait sets the time for waiting that reading occur
func SetReadingWait(wait time.Duration) DialerConfig {
	return func(ws *websocket) {
		ws.readingWait = wait
	}
}

//SetBufferSize sets the read/write buffer size
func SetBufferSize(readBufferSize int, writeBufferSize int) DialerConfig {
	return func(ws *websocket) {
		ws.readBufSize = readBufferSize
		ws.writeBufSize = writeBufferSize
	}
}

// websocketDialerFactoryFun exchange/ set the factory function used to create the dialer which
// is then used to open the websocket connection.
// This function is not exported on purpose, it should only used for injection and mocking in tests!!
func websocketDialerFactoryFun(wsDialerFactory websocketDialerFactory) DialerConfig {
	return func(ws *websocket) {
		ws.wsDialerFactory = wsDialerFactory
	}
}
