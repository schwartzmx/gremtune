package gremcos

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cast"
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

// New creates a new instance of the Cosmos (-DB connector)
func New(host string, options ...Option) (*Cosmos, error) {
	cosmos := &Cosmos{
		logger:                  zerolog.Nop(),
		errorChannel:            make(chan error),
		host:                    host,
		numMaxActiveConnections: 10,
		connectionIdleTimeout:   time.Second * 30,
	}

	for _, opt := range options {
		opt(cosmos)
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
			// consume the errors from the channel
			// at the moment it is not needed to post them to the log since they are
			// anyway handed over to the caller
			// For debugging the following line can be uncommented
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

	resp, err := c.pool.Execute(query)

	// try to investigate the responses and to find out if we can find more specific error information
	if respErr := extractFirstError(resp); respErr != nil {
		err = respErr
	}

	return resp, err
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

// extractFirstError runs through the given responses and returns the first error it finds.
// All information (e.g.)
func extractFirstError(responses []interfaces.Response) error {

	for _, response := range responses {
		statusCode := response.Status.Code

		// everything ok --> skip this response
		if statusCode == interfaces.StatusSuccess || statusCode == interfaces.StatusNoContent || statusCode == interfaces.StatusPartialContent {
			continue
		}

		// since all success codes are already skipped
		// here we have an error

		// Do specific a interpretation on the 500 errors if possible.
		// Usually from CosmosDB we can use additional headers to extract more detail
		if statusCode == interfaces.StatusServerError {
			responseInfo, err := parseAttributeMap(response.Status.Attributes)
			if err != nil {
				// if we can't parse/ interpret the attribute map then we return the full/ unparsed error information
				return fmt.Errorf("Failed parsing attributes of response: '%s'. Unparsed error: %d - %s", err.Error(), response.Status.Code, response.Status.Message)
			}
			return fmt.Errorf("%d (%d) - %s", responseInfo.statusCode, responseInfo.subStatusCode, responseInfo.statusDescription)
		}

		// for the remaining error status codes do the usual error detection mechanism based on the main status code
		return extractError(response)
	}

	return nil
}

func parseAttributeMap(attributes map[string]interface{}) (responseInformation, error) {

	responseInfo := responseInformation{}
	if valueStr, ok := attributes[string(headerStatusCode)]; ok {

		value, err := cast.ToInt16E(valueStr)
		if err != nil {
			return responseInfo, errors.Wrapf(err, "Failed parsing '%s'", headerStatusCode)
		}
		statusCode := int(value)
		responseInfo.statusCode = statusCode
		responseInfo.statusDescription = statusCodeToDescription(statusCode)
	}

	if valueStr, ok := attributes[string(headerSubStatusCode)]; ok {
		responseInfo.subStatusCode = int(cast.ToInt16(valueStr))
	}

	if valueStr, ok := attributes[string(headerRequestCharge)]; ok {
		responseInfo.requestCharge = cast.ToFloat32(valueStr)
	}

	if valueStr, ok := attributes[string(headerRequestChargeTotal)]; ok {
		responseInfo.requestChargeTotal = cast.ToFloat32(valueStr)
	}

	if valueStr, ok := attributes[string(headerServerTimeMS)]; ok {
		responseInfo.serverTime = time.Microsecond * time.Duration(1000*cast.ToFloat32(valueStr))
	}

	if valueStr, ok := attributes[string(headerServerTimeMSTotal)]; ok {
		responseInfo.serverTimeTotal = time.Microsecond * time.Duration(1000*cast.ToFloat32(valueStr))
	}

	if valueStr, ok := attributes[string(headerActivityID)]; ok {
		responseInfo.activityID = cast.ToString(valueStr)
	}

	if valueStr, ok := attributes[string(headerRetryAfterMS)]; ok {
		retryAfter, err := time.Parse("15:04:05.999999999", cast.ToString(valueStr))
		zeroTime, _ := time.Parse("15:04:05.999999999", "00:00:00.000")
		responseInfo.retryAfter = retryAfter.Sub(zeroTime)
		if err != nil {
			responseInfo.retryAfter = 0
		}
	}

	if valueStr, ok := attributes[string(headerSource)]; ok {
		responseInfo.source = cast.ToString(valueStr)
	}

	return responseInfo, nil
}

func statusCodeToDescription(code int) string {
	desc, ok := statusCodeDescription[code]
	if !ok {
		return fmt.Sprintf("Status code %d is unknown", code)
	}
	return desc
}

type responseInformation struct {
	statusCode         int
	subStatusCode      int
	statusDescription  string
	requestCharge      float32
	requestChargeTotal float32
	serverTime         time.Duration
	serverTimeTotal    time.Duration
	activityID         string
	retryAfter         time.Duration
	source             string
}

// statusCodeDescription provides the description for status codes taken from https://docs.microsoft.com/en-us/azure/cosmos-db/gremlin-headers#status-codes
var statusCodeDescription = map[int]string{
	401:  "Error message 'Unauthorized: Invalid credentials provided' is returned when authentication password doesn't match Cosmos DB account key. Navigate to your Cosmos DB Gremlin account in the Azure portal and confirm that the key is correct.",
	404:  "Concurrent operations that attempt to delete and update the same edge or vertex simultaneously. Error message 'Owner resource does not exist' indicates that specified database or collection is incorrect in connection parameters in /dbs/<database name>/colls/<collection or graph name> format.",
	408:  "'Server timeout' indicates that traversal took more than 30 seconds and was canceled by the server. Optimize your traversals to run quickly by filtering vertices or edges on every hop of traversal to narrow down search scope.",
	409:  "'Conflicting request to resource has been attempted. Retry to avoid conflicts.' This usually happens when vertex or an edge with an identifier already exists in the graph.",
	412:  "Status code is complemented with error message 'PreconditionFailedException': One of the specified pre-condition is not met. This error is indicative of an optimistic concurrency control violation between reading an edge or vertex and writing it back to the store after modification. Most common situations when this error occurs is property modification, for example g.V('identifier').property('name','value'). Gremlin engine would read the vertex, modify it, and write it back. If there is another traversal running in parallel trying to write the same vertex or an edge, one of them will receive this error. Application should submit traversal to the server again.",
	429:  "Request was throttled and should be retried after value in x-ms-retry-after-ms",
	500:  "Error message that contains 'NotFoundException: Entity with the specified id does not exist in the system.' indicates that a database and/or collection was re-created with the same name. This error will disappear within 5 minutes as change propagates and invalidates caches in different Cosmos DB components. To avoid this issue, use unique database and collection names every time.",
	1000: "This status code is returned when server successfully parsed a message but wasn't able to execute. It usually indicates a problem with the query.",
	1001: "This code is returned when server completes traversal execution but fails to serialize response back to the client. This error can happen when traversal generates complex result, that is too large or does not conform to TinkerPop protocol specification. Application should simplify the traversal when it encounters this error.",
	1003: "'Query exceeded memory limit. Bytes Consumed: XXX, Max: YYY' is returned when traversal exceeds allowed memory limit. Memory limit is 2 GB per traversal.",
	1004: "This status code indicates malformed graph request. Request can be malformed when it fails deserialization, non-value type is being deserialized as value type or unsupported gremlin operation requested. Application should not retry the request because it will not be successful.",
	1007: "Usually this status code is returned with error message 'Could not process request. Underlying connection has been closed.'. This situation can happen if client driver attempts to use a connection that is being closed by the server. Application should retry the traversal on a different connection.",
	1008: "Cosmos DB Gremlin server can terminate connections to rebalance traffic in the cluster. Client drivers should handle this situation and use only live connections to send requests to the server. Occasionally client drivers may not detect that connection was closed. When application encounters an error, 'Connection is too busy. Please retry after sometime or open more connections.' it should retry traversal on a different connection.",
}

// Responseheaders for CosmosDB, taken from: https://docs.microsoft.com/en-us/azure/cosmos-db/gremlin-headers#headers
type cosmosDBResponseHeader string

const (
	headerRequestCharge      cosmosDBResponseHeader = "x-ms-request-charge"       // double
	headerRequestChargeTotal cosmosDBResponseHeader = "x-ms-total-request-charge" // double
	headerServerTimeMS       cosmosDBResponseHeader = "x-ms-server-time-ms"       // double
	headerServerTimeMSTotal  cosmosDBResponseHeader = "x-ms-total-server-time-ms" // double
	headerStatusCode         cosmosDBResponseHeader = "x-ms-status-code"          // long
	headerSubStatusCode      cosmosDBResponseHeader = "x-ms-substatus-code"       // long
	headerRetryAfterMS       cosmosDBResponseHeader = "x-ms-retry-after-ms"       // string
	headerActivityID         cosmosDBResponseHeader = "x-ms-activity-id"          // string
	headerSource             cosmosDBResponseHeader = "x-ms-source"               // string
)
