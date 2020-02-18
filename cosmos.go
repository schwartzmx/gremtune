package gremcos

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/supplyon/gremcos/interfaces"
)

// Cosmos is a connector that can be used to connect to and interact with a CosmosDB
type Cosmos struct {
	logger zerolog.Logger

	errorChannel chan error

	host     string
	username string
	password string

	// dialer is used to create/ dial new connections if needed
	dialer interfaces.Dialer

	// pool the connection pool
	pool                    interfaces.QueryExecutor
	numMaxActiveConnections int
	connectionIdleTimeout   time.Duration

	// metrics for cosmos
	metrics *Metrics

	wg sync.WaitGroup
}

// Option is the struct for defining optional parameters for Cosmos
type Option func(*Cosmos)

// WithAuth sets credentials for an authenticated connection
func WithAuth(username string, password string) Option {
	return func(c *Cosmos) {
		c.username = username
		c.password = password
	}
}

// WithLogger specifies the logger to use
func WithLogger(logger zerolog.Logger) Option {
	return func(c *Cosmos) {
		c.logger = logger
	}
}

// ConnectionIdleTimeout specifies the timeout after which idle
// connections will be removed from the internal connection pool
func ConnectionIdleTimeout(timeout time.Duration) Option {
	return func(c *Cosmos) {
		c.connectionIdleTimeout = timeout
	}
}

// NumMaxActiveConnections specifies the maximum amount of active connections.
func NumMaxActiveConnections(numMaxActiveConnections int) Option {
	return func(c *Cosmos) {
		c.numMaxActiveConnections = numMaxActiveConnections
	}
}

// MetricsPrefix can be used to customize the metrics prefix
// as needed for a specific service. Per default 'gremcos' is used
// as prefix.
func MetricsPrefix(prefix string) Option {
	return func(c *Cosmos) {
		c.metrics = NewMetrics(prefix)
	}
}

// withMetrics can be used to set metrics from the outside.
// This is needed in order to be able to inject mocks for unit-tests.
func withMetrics(metrics *Metrics) Option {
	return func(c *Cosmos) {
		c.metrics = metrics
	}
}

// New creates a new instance of the Cosmos (-DB connector)
func New(host string, options ...Option) (*Cosmos, error) {
	cosmos := &Cosmos{
		logger:                  zerolog.Nop(),
		errorChannel:            make(chan error),
		host:                    host,
		numMaxActiveConnections: 10,
		connectionIdleTimeout:   time.Second * 30,
		metrics:                 nil,
	}

	for _, opt := range options {
		opt(cosmos)
	}

	// if metrics not set via MetricsPrefix instantiate the metrics
	// using the default prefix
	if cosmos.metrics == nil {
		cosmos.metrics = NewMetrics("gremcos")
	}

	// use default settings (timeout, buffersizes etc.) for the websocket
	dialer, err := NewWebsocket(host)
	if err != nil {
		return nil, err
	}
	cosmos.dialer = dialer

	pool, err := NewPool(cosmos.dial, cosmos.numMaxActiveConnections, cosmos.connectionIdleTimeout, cosmos.logger)
	if err != nil {
		return nil, err
	}
	cosmos.pool = pool

	// set up a consumer for all the errors that are posted by the
	// clients on the error channel
	cosmos.wg.Add(1)
	go func() {
		defer cosmos.wg.Done()
		for range cosmos.errorChannel {
			// consume the errors from the channel at the moment it is not needed to post them to the log since they are
			// anyway handed over to the caller. For debugging the following line can be uncommented
			// cosmos.logger.Error().Err(err).Msg("Error from connection pool received")
		}
		cosmos.logger.Debug().Msg("Error channel consumer closed")
	}()

	return cosmos, nil
}

// dial creates new connections. It is called by the pool in case a new connection is demanded.
func (c *Cosmos) dial() (interfaces.QueryExecutor, error) {
	return Dial(c.dialer, c.errorChannel, SetAuth(c.username, c.password), PingInterval(time.Second*30))
}

func (c *Cosmos) Execute(query string) ([]interfaces.Response, error) {

	responses, err := c.pool.Execute(query)

	// try to investigate the responses and to find out if we can find more specific error information
	if respErr := extractFirstError(responses); respErr != nil {
		err = respErr
	}

	updateRequestMetrics(responses, c.metrics)
	return responses, err
}

func (c *Cosmos) ExecuteAsync(query string, responseChannel chan interfaces.AsyncResponse) (err error) {
	return c.pool.ExecuteAsync(query, responseChannel)
}

func (c *Cosmos) IsConnected() bool {
	return c.pool.IsConnected()
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

// IsHealthy returns nil if the Cosmos DB connection is alive, otherwise an error is returned
func (c *Cosmos) IsHealthy() error {
	return c.pool.Ping()
}

// updateRequestMetrics updates the request relevant metrics based on the given chunk of responses
func updateRequestMetrics(respones []interfaces.Response, metrics *Metrics) {

	// nothing to update
	if len(respones) == 0 {
		return
	}

	retryAfter := time.Second * 0
	var requestChargePerQueryTotal float32
	var serverTimePerQueryTotal time.Duration

	for _, response := range respones {
		statusCode := response.Status.Code
		respInfo, err := parseAttributeMap(response.Status.Attributes)

		if err != nil {
			// parsing the response failed -> we use the unspecific status code
			metrics.statusCodeTotal.WithLabelValues(fmt.Sprintf("%d", statusCode)).Inc()
			continue
		}

		// use the more specific status code
		statusCode = respInfo.statusCode
		metrics.statusCodeTotal.WithLabelValues(fmt.Sprintf("%d", statusCode)).Inc()

		// only take the largest waittime of this chunk of responses
		if retryAfter < respInfo.retryAfter {
			retryAfter = respInfo.retryAfter
		}

		// only take the largest value since cosmos already accumulates this value
		if requestChargePerQueryTotal < respInfo.requestChargeTotal {
			requestChargePerQueryTotal = respInfo.requestChargeTotal
		}

		// only take the largest value since cosmos already accumulates this value
		if serverTimePerQueryTotal < respInfo.serverTimeTotal {
			serverTimePerQueryTotal = respInfo.serverTimeTotal
		}
	}

	numResponses := len(respones)
	var requestChargePerQueryResponseAvg float64
	if numResponses > 0 {
		requestChargePerQueryResponseAvg = float64(requestChargePerQueryTotal) / float64(numResponses)
	}

	var serverTimePerQueryResponseAvg float64
	if numResponses > 0 {
		serverTimePerQueryResponseAvg = float64(serverTimePerQueryTotal.Milliseconds()) / float64(numResponses)
	}

	metrics.serverTimePerQueryResponseAvgMS.Set(serverTimePerQueryResponseAvg)
	metrics.serverTimePerQueryMS.Set(float64(serverTimePerQueryTotal.Milliseconds()))
	metrics.requestChargePerQueryResponseAvg.Set(requestChargePerQueryResponseAvg)
	metrics.requestChargePerQuery.Set(float64(requestChargePerQueryTotal))
	metrics.requestChargeTotal.Add(float64(requestChargePerQueryTotal))
	metrics.retryAfterMS.Set(float64(retryAfter.Milliseconds()))
}
