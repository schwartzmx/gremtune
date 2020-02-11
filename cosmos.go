package gremtune

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/schwartzmx/gremtune/interfaces"
)

type Cosmos struct {
	logger zerolog.Logger
	dialer interfaces.Dialer
	pool   interfaces.QueryExecutor
}

func New(host string, logger zerolog.Logger) (*Cosmos, error) {

	dialer, err := NewDialer(host)
	if err != nil {
		return nil, err
	}

	cosmos := &Cosmos{
		logger: logger,
		dialer: dialer,
	}

	pool, err := NewPool(cosmos.dial, 10, time.Second*30)
	if err != nil {
		return nil, err
	}
	cosmos.pool = pool

	return cosmos, nil
}

func (c *Cosmos) dial() (interfaces.QueryExecutor, error) {
	return Dial(c.dialer, nil)
}

func (c *Cosmos) Execute(query string) (resp []interfaces.Response, err error) {
	return c.pool.Execute(query)
}
