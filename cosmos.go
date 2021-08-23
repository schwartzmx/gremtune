package gremcos

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/supplyon/gremcos/interfaces"
)

// Cosmos is an abstraction of the CosmosDB
type Cosmos interface {
	// ExecuteQuery executes the given query and returns the according responses from the CosmosDB
	ExecuteQuery(query interfaces.QueryBuilder) ([]interfaces.Response, error)

	// Execute can be used to execute a raw query (string). This can be used to issue queries that are not yet supported by the QueryBuilder.
	Execute(query string) ([]interfaces.Response, error)

	// ExecuteAsync can be used to issue a query and streaming in the responses as they are available / are provided by the CosmosDB
	ExecuteAsync(query string, responseChannel chan interfaces.AsyncResponse) (err error)

	// IsConnected returns true in case the connection to the CosmosDB is up, false otherwise.
	IsConnected() bool

	// Stop stops the connector, terminates all background go routines and closes open connections.
	Stop() error

	// String
	String() string

	// IsHealthy returns nil in case the connection to the CosmosDB is up, the according error otherwise.
	IsHealthy() error
}

// cosmos is a connector that can be used to connect to and interact with a CosmosDB
type cosmosImpl struct {
	logger zerolog.Logger

	errorChannel chan error

	host string

	// pool the connection pool
	pool                    interfaces.QueryExecutor
	numMaxActiveConnections int
	connectionIdleTimeout   time.Duration

	// websocketGenerator is a function that is responsible to spawn new websocket
	// connections if needed.
	websocketGenerator websocketGeneratorFun

	// metrics for cosmos
	metrics *Metrics

	wg sync.WaitGroup

	credentialProvider CredentialProvider
}

type websocketGeneratorFun func(host string, options ...optionWebsocket) (interfaces.Dialer, error)

// Option is the struct for defining optional parameters for Cosmos
type Option func(*cosmosImpl)

// WithAuth sets credentials for an authenticated connection using static credentials (primary-/ secondary cosmos key as password)
func WithAuth(username string, password string) Option {
	return func(c *cosmosImpl) {
		c.credentialProvider = StaticCredentialProvider{
			UsernameStatic: username,
			PasswordStatic: password,
		}
	}
}

// WithResourceTokenAuth sets credential provider that is used to authenticate the requests to cosmos.
// With this approach dynamic credentials (cosmos resource tokens) can be used for authentication.
// To do this you have to provide a CredentialProvider implementation that takes care for providing a valid (not yet expired) resource token
//	myResourceTokenProvider := MyDynamicCredentialProvider{}
//	New("wss://example.com", WithResourceTokenAuth(myResourceTokenProvider))
//
// If you want to use static credentials (primary-/ secondary cosmos key as password) instead you can either use "WithAuth".
//	New("wss://example.com", WithAuth("username","primary-key"))
// Or you use the default implementation for a static credentials provider "StaticCredentialProvider"
//	staticCredProvider := StaticCredentialProvider{UsernameStatic: "username", PasswordStatic: "primary-key"}
//	New("wss://example.com", WithResourceTokenAuth(staticCredProvider))
func WithResourceTokenAuth(credentialProvider CredentialProvider) Option {
	return func(c *cosmosImpl) {
		c.credentialProvider = credentialProvider
	}
}

// WithLogger specifies the logger to use
func WithLogger(logger zerolog.Logger) Option {
	return func(c *cosmosImpl) {
		c.logger = logger
	}
}

// ConnectionIdleTimeout specifies the timeout after which idle
// connections will be removed from the internal connection pool
func ConnectionIdleTimeout(timeout time.Duration) Option {
	return func(c *cosmosImpl) {
		c.connectionIdleTimeout = timeout
	}
}

// NumMaxActiveConnections specifies the maximum amount of active connections.
func NumMaxActiveConnections(numMaxActiveConnections int) Option {
	return func(c *cosmosImpl) {
		c.numMaxActiveConnections = numMaxActiveConnections
	}
}

// MetricsPrefix can be used to customize the metrics prefix
// as needed for a specific service. Per default 'gremcos' is used
// as prefix.
func MetricsPrefix(prefix string) Option {
	return func(c *cosmosImpl) {
		c.metrics = NewMetrics(prefix)
	}
}

// withMetrics can be used to set metrics from the outside.
// This is needed in order to be able to inject mocks for unit-tests.
func withMetrics(metrics *Metrics) Option {
	return func(c *cosmosImpl) {
		c.metrics = metrics
	}
}

// wsGenerator can be used to set the generator to create websockets for the outside.
// This is needed in order to be able to inject mocks for unit-tests.
func wsGenerator(wsGenerator websocketGeneratorFun) Option {
	return func(c *cosmosImpl) {
		c.websocketGenerator = wsGenerator
	}
}

// New creates a new instance of the Cosmos (-DB connector)
func New(host string, options ...Option) (Cosmos, error) {
	cosmos := &cosmosImpl{
		logger:                  zerolog.Nop(),
		errorChannel:            make(chan error),
		host:                    host,
		numMaxActiveConnections: 10,
		connectionIdleTimeout:   time.Second * 30,
		metrics:                 nil,
		websocketGenerator:      NewWebsocket,
		credentialProvider:      noCredentials{},
	}

	for _, opt := range options {
		opt(cosmos)
	}

	// if metrics not set via MetricsPrefix instantiate the metrics
	// using the default prefix
	if cosmos.metrics == nil {
		cosmos.metrics = NewMetrics("gremcos")
	}

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
			if err != nil {
				cosmos.logger.Error().Err(err).Msg("Error from connection pool received")
			}
		}
		cosmos.logger.Debug().Msg("Error channel consumer closed")
	}()

	return cosmos, nil
}

// dial creates new connections. It is called by the pool in case a new connection is demanded.
func (c *cosmosImpl) dial() (interfaces.QueryExecutor, error) {

	// create a new websocket dialer to avoid using the same websocket connection for
	// multiple queries at the same time
	// use default settings (timeout, buffersizes etc.) for the websocket
	dialer, err := c.websocketGenerator(c.host)
	if err != nil {
		return nil, err
	}

	return Dial(dialer, c.errorChannel, SetAuth(c.credentialProvider), PingInterval(time.Second*30))
}

func (c *cosmosImpl) ExecuteQuery(query interfaces.QueryBuilder) ([]interfaces.Response, error) {
	if query == nil {
		return nil, fmt.Errorf("Query is nil")
	}
	return c.Execute(query.String())
}

func (c *cosmosImpl) Execute(query string) ([]interfaces.Response, error) {

	responses, err := c.pool.Execute(query)

	// try to investigate the responses and to find out if we can find more specific error information
	if respErr := extractFirstError(responses); respErr != nil {
		err = respErr
	}

	updateRequestMetrics(responses, c.metrics)
	return responses, err
}

func (c *cosmosImpl) ExecuteAsync(query string, responseChannel chan interfaces.AsyncResponse) (err error) {
	return c.pool.ExecuteAsync(query, responseChannel)
}

func (c *cosmosImpl) IsConnected() bool {
	return c.pool.IsConnected()
}

func (c *cosmosImpl) Stop() error {
	defer func() {
		close(c.errorChannel)
		c.wg.Wait()
	}()
	c.logger.Info().Msg("Teardown requested")

	return c.pool.Close()
}

func (c *cosmosImpl) String() string {
	username, err := c.credentialProvider.Username()
	if err != nil {
		username = fmt.Sprintf("failed to obtain username: %v", err)
	}
	return fmt.Sprintf("CosmosDB (connected=%t, target=%s, user=%s)", c.IsConnected(), c.host, username)
}

// IsHealthy returns nil if the Cosmos DB connection is alive, otherwise an error is returned
func (c *cosmosImpl) IsHealthy() error {
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
	var serverTimePerQueryResponseAvg float64
	if numResponses > 0 {
		requestChargePerQueryResponseAvg = float64(requestChargePerQueryTotal) / float64(numResponses)
		serverTimePerQueryResponseAvg = float64(serverTimePerQueryTotal.Milliseconds()) / float64(numResponses)
	}

	metrics.serverTimePerQueryResponseAvgMS.Set(serverTimePerQueryResponseAvg)
	metrics.serverTimePerQueryMS.Set(float64(serverTimePerQueryTotal.Milliseconds()))
	metrics.requestChargePerQueryResponseAvg.Set(requestChargePerQueryResponseAvg)
	metrics.requestChargePerQuery.Set(float64(requestChargePerQueryTotal))
	metrics.requestChargeTotal.Add(float64(requestChargePerQueryTotal))
	metrics.retryAfterMS.Set(float64(retryAfter.Milliseconds()))
}
