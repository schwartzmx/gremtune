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

	errorChannel chan error
}

func New(host string, logger zerolog.Logger) (*Cosmos, error) {

	dialer, err := NewDialer(host)
	if err != nil {
		return nil, err
	}

	cosmos := &Cosmos{
		logger:       logger,
		dialer:       dialer,
		errorChannel: make(chan error),
	}

	go func() {
		err := <-cosmos.errorChannel
		cosmos.logger.Error().Err(err).Msg("Error received")
	}()

	pool, err := NewPool(cosmos.dial, 10, time.Second*30)
	if err != nil {
		return nil, err
	}
	cosmos.pool = pool

	return cosmos, nil
}

func (c *Cosmos) dial() (interfaces.QueryExecutor, error) {
	return Dial(c.dialer, c.errorChannel)
}

func (c *Cosmos) Execute(query string) (resp []interfaces.Response, err error) {
	return c.pool.Execute(query)
}
