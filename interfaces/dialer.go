package interfaces

type Dialer interface {
	Connect() error
	IsConnected() bool
	Write([]byte) error
	Read() (int, []byte, error)
	Close() error
	Ping() error
}
