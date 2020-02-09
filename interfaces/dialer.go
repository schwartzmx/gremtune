package interfaces

type Dialer interface {
	Connect() error
	IsConnected() bool
	IsDisposed() bool
	Write([]byte) error
	Read() (int, []byte, error)
	Close() error
	Ping() error
	GetQuitChannel() <-chan struct{}
}
