package gremcos

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"

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

	// ExecuteWithBindings can be used to execute a raw query (string) with optional bindings/rebindings. This can be used to issue queries that are not yet supported by the QueryBuilder.
	ExecuteWithBindings(path string, bindings, rebindings map[string]interface{}) (resp []interfaces.Response, err error)

	// IsConnected returns true in case the connection to the CosmosDB is up, false otherwise.
	IsConnected() bool

	// Stop stops the connector, terminates all background go routines and closes open connections.
	Stop() error

	// String returns a string representation of the cosmos connector
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

	// defines the number of times a request is retried if suggested by cosmos
	maxRetries int
	// defines the max duration a request should be retried
	retryTimeout time.Duration
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
// Or you use the default implementation for a static credential provider "StaticCredentialProvider"
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

// AutomaticRetries tries to retry failed requests, if appropriate. Retries are limited to maxRetries. Retrying is stopped after timeout is reached.
// Appropriate error codes are 409, 412, 429, 1007, 1008 see https://docs.microsoft.com/en-us/azure/cosmos-db/graph/gremlin-headers#status-codes
// Hint: Be careful when specifying the values for maxRetries and timeout. They influence how much latency is added on requests that need to be retried.
//       For example if maxRetries = 1 and timeout = 1s the call might take 1s longer to return a potential persistent error.
func AutomaticRetries(maxRetries int, timeout time.Duration) Option {
	return func(c *cosmosImpl) {
		if maxRetries > 0 {
			c.maxRetries = maxRetries
		}

		c.retryTimeout = timeout
		if timeout <= 0 {
			c.retryTimeout = time.Second * 30
		}
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

	return Dial(dialer, c.errorChannel, SetAuth(c.credentialProvider), PingInterval(time.Second*30), WithMetrics(c.metrics))
}

func (c *cosmosImpl) ExecuteQuery(query interfaces.QueryBuilder) ([]interfaces.Response, error) {
	if query == nil {
		return nil, fmt.Errorf("query is nil")
	}
	return c.Execute(query.String())
}

func (c *cosmosImpl) Execute(query string) ([]interfaces.Response, error) {

	doRetry := func() ([]interfaces.Response, error) {
		return c.pool.Execute(query)
	}

	responses, err := retryLoop(doRetry, c.maxRetries, c.retryTimeout, c.metrics, c.logger)

	// try to investigate the responses and to find out if we can find more specific error information
	if respErr := extractFirstError(responses); respErr != nil {
		err = respErr
	}

	return responses, err
}

func (c *cosmosImpl) ExecuteWithBindings(query string, bindings, rebindings map[string]interface{}) ([]interfaces.Response, error) {

	doRetry := func() ([]interfaces.Response, error) {
		return c.pool.ExecuteWithBindings(query, bindings, rebindings)
	}

	responses, err := retryLoop(doRetry, c.maxRetries, c.retryTimeout, c.metrics, c.logger)

	// try to investigate the responses and to find out if we can find more specific error information
	if respErr := extractFirstError(responses); respErr != nil {
		err = respErr
	}

	return responses, err
}

type retryFun func() ([]interfaces.Response, error)

func retryLoop(executeRequest retryFun, maxRetries int, retryTimeout time.Duration, metrics *Metrics, logger zerolog.Logger) (responses []interfaces.Response, err error) {
	if metrics == nil {
		return nil, fmt.Errorf("metrics must not be nil")
	}

	var tryCount int
	shouldRetry := maxRetries > 0
	maxTries := maxRetries + 1

	done := make(chan bool)
	defer close(done)

	timeoutReachedChan := handleTimeout(done, retryTimeout, logger)

	for tryCount = 0; tryCount < maxTries; tryCount++ {
		responses, err = executeRequest()
		isARetry := tryCount > 0
		updateRequestMetrics(responses, metrics, isARetry)

		// error is handled late to ensure an update of the metrics
		if err != nil {
			metrics.requestErrorsTotal.Inc()
			return nil, errors.Wrap(err, "executing request in retry loop")
		}

		if !shouldRetry {
			return responses, nil
		}

		retryInformation := extractRetryConditions(responses)

		// Retry is always on a new or at least active connection,
		// therefore retryInformation.retryOnNewConnection can be used here as well
		if !(retryInformation.retry || retryInformation.retryOnNewConnection) {
			return responses, nil
		}

		if retryInformation.retryAfter > 0 {
			logger.Info().Msgf("retry %d of query after %v because of header status code %d", tryCount+1, retryInformation.retryAfter, retryInformation.responseStatusCode)

			if waitDone := waitForRetry(retryInformation.retryAfter, timeoutReachedChan); !waitDone {
				// timeout occurred
				logger.Warn().Msgf("Timed out while waiting to do a retry after %s (timeout=%s)", retryInformation.retryAfter, retryTimeout)
				metrics.requestRetryTimeoutsTotal.Inc()
				return responses, nil
			}
		}

		// Timeout check in case no waiting is required
		select {
		case <-timeoutReachedChan:
			// we stop here and return what we got so far
			metrics.requestRetryTimeoutsTotal.Inc()
			logger.Warn().Msgf("Timed out while doing a retry (timeout=%s)", retryTimeout)
			return responses, nil
		default:
			continue
			// continue with next retry
		}
	}

	return responses, err
}

func handleTimeout(done <-chan bool, retryTimeout time.Duration, logger zerolog.Logger) (timedOutChan <-chan bool) {
	timeoutReachedChan := make(chan bool)

	go func() {
		retryTimeoutTimer := time.NewTimer(retryTimeout)

		defer close(timeoutReachedChan)

		select {
		case <-retryTimeoutTimer.C:
			// no further retries, we return the current responses
			logger.Debug().Msgf("Specified timout (%v) for retries exceeded. Hence the current request won't be retried in case suggests to retry. This message does not indicate that the request itself failed or timed out.", retryTimeout)
			timeoutReachedChan <- true
			return
		case <-done:
			retryTimeoutTimer.Stop()
			return
		}
	}()
	return timeoutReachedChan
}

func waitForRetry(wait time.Duration, stop <-chan bool) (waitDone bool) {
	waitForRetryTimer := time.NewTimer(wait)
	defer waitForRetryTimer.Stop()

	select {
	case <-stop:
		return false
	case <-waitForRetryTimer.C:
		return true
	}
}

func (c *cosmosImpl) executeAsync(query string, asyncResponses *[]interfaces.AsyncResponse, errorCallback func(err error)) (responses []interfaces.Response, err error) {
	intermediateChannel := make(chan interfaces.AsyncResponse, 100)

	if err := c.pool.ExecuteAsync(query, intermediateChannel); err != nil {
		return nil, err
	}
	errorCallback(err)

	responses = make([]interfaces.Response, 0, 5)
	*asyncResponses = make([]interfaces.AsyncResponse, 0, 5)

	for resp := range intermediateChannel {
		*asyncResponses = append(*asyncResponses, resp)
		responses = append(responses, resp.Response)
		if resp.ErrorMessage != "" {
			if err == nil {
				err = errors.New(resp.ErrorMessage)
				continue
			}
			err = errors.Wrap(err, resp.ErrorMessage)
		}
	}

	return responses, err
}

func (c *cosmosImpl) ExecuteAsync(query string, responseChannel chan interfaces.AsyncResponse) (err error) {

	var asyncResponses []interfaces.AsyncResponse

	wg := sync.WaitGroup{}
	wg.Add(1)

	var returnAfterFirstCall sync.Once
	var firstCallError error
	var firstCallErrorLock sync.Mutex

	errCallback := func(callbackErr error) {
		returnAfterFirstCall.Do(func() {
			firstCallErrorLock.Lock()
			firstCallError = callbackErr
			firstCallErrorLock.Unlock()
			wg.Done()
		})
	}

	doRetry := func() ([]interfaces.Response, error) {
		return c.executeAsync(query, &asyncResponses, errCallback)
	}

	go func() {
		defer close(responseChannel)
		_, retryErr := retryLoop(doRetry, c.maxRetries, c.retryTimeout, c.metrics, c.logger)

		if retryErr != nil {
			// return because the asyncResponses we gathered might be outdated
			return
		}

		// Write final result
		for _, response := range asyncResponses {
			responseChannel <- response
		}
	}()

	wg.Wait()
	firstCallErrorLock.Lock()
	defer firstCallErrorLock.Unlock()
	return firstCallError
}

func (c *cosmosImpl) IsConnected() bool {
	return c.pool.IsConnected()
}

func (c *cosmosImpl) Stop() error {
	c.logger.Info().Msg("Teardown requested")

	poolCloseErr := c.pool.Close()

	close(c.errorChannel)
	c.wg.Wait()

	return poolCloseErr
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
func updateRequestMetrics(responses []interfaces.Response, metrics *Metrics, isARetry bool) {
	if isARetry {
		metrics.requestRetiesTotal.Inc()
	}

	// nothing to update
	if len(responses) == 0 {
		return
	}

	retryAfter := time.Second * 0
	var requestChargePerQueryTotal float32
	var serverTimePerQueryTotal time.Duration

	for _, response := range responses {
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

		// only take the largest wait time of this chunk of responses
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

	numResponses := len(responses)
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
	metrics.retryAfterMS.Observe(float64(retryAfter.Milliseconds()))
}
