package interfaces

import "time"

// WebsocketConnection is the minimal interface needed to act on a websocket
type WebsocketConnection interface {
	SetPongHandler(handler func(appData string) error)
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
	WriteControl(messageType int, data []byte, deadline time.Time) error
}
