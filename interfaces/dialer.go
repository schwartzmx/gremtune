package interfaces

// Dialer represents a entity that is able to open a websocket and work (read/ write) on it.
type Dialer interface {
	Connect() error
	IsConnected() bool
	Write([]byte) error
	Read() (int, []byte, error)
	Close() error
	Ping() error
}
