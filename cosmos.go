package gremtune

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/schwartzmx/gremtune/interfaces"
)

type Cosmos struct {
	logger zerolog.Logger
	dialer interfaces.Dialer
	pool   interfaces.QueryExecutor

	errorChannel chan error
	host         string
	username     string
	password     string

	wg sync.WaitGroup
}

func New(host, username, password string, logger zerolog.Logger) (*Cosmos, error) {

	dialer, err := NewDialer(host)
	if err != nil {
		return nil, err
	}

	cosmos := &Cosmos{
		logger:       logger,
		dialer:       dialer,
		errorChannel: make(chan error),
		host:         host,
		username:     username,
		password:     password,
	}

	cosmos.wg.Add(1)
	go func() {
		defer cosmos.wg.Done()
		for err := range cosmos.errorChannel {
			cosmos.logger.Error().Err(err).Msg("Error received")
		}
		cosmos.logger.Debug().Msg("Error channel consumer closed")
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

func (c *Cosmos) IsConnected() bool {
	// TODO: Implement
	return true
}

func (c *Cosmos) Stop() error {
	defer func() {
		close(c.errorChannel)
		c.wg.Wait()
	}()
	c.logger.Info().Msg("Teardown requested")

	return c.pool.Close()
}

func (c *Cosmos) String() string {
	return fmt.Sprintf("CosmosDB (connected=%t, target=%s, user=%s)", c.IsConnected(), c.host, c.username)
}
