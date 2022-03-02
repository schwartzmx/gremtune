package gremcos

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/supplyon/gremcos/interfaces"
)

//nolint:structcheck // reports responseStatusCode wrong
type cosmosStatusCodeDescription struct {
	responseStatusCode            int
	retry                         bool
	// retryOnNewConnection defines whether a retry has to be done on a new connection or not.
	// This is always given when we act on the pool, but has to be regarded when the client is exported/used.
	retryOnNewConnection          bool
	responseStatusCodeDescription string
}

type cosmosRetryInformation struct {
	cosmosStatusCodeDescription
	retryAfter time.Duration
}

var noRetry = cosmosStatusCodeDescription{retry: false}

// statusCodeDescription provides the description for status codes taken from https://docs.microsoft.com/en-us/azure/cosmos-db/gremlin-headers#status-codes
var statusCodeDescription = map[int]cosmosStatusCodeDescription{
	401:  {401, false, false, "Error message 'Unauthorized: Invalid credentials provided' is returned when authentication password doesn't match Cosmos DB account key. Navigate to your Cosmos DB Gremlin account in the Azure portal and confirm that the key is correct."},
	404:  {404, false, false, "Concurrent operations that attempt to delete and update the same edge or vertex simultaneously. Error message 'Owner resource does not exist' indicates that specified database or collection is incorrect in connection parameters in /dbs/<database name>/colls/<collection or graph name> format."},
	408:  {408, false, false, "'Server timeout' indicates that traversal took more than 30 seconds and was canceled by the server. Optimize your traversals to run quickly by filtering vertices or edges on every hop of traversal to narrow down search scope."},
	409:  {409, true, false, "'Conflicting request to resource has been attempted. Retry to avoid conflicts.' This usually happens when vertex or an edge with an identifier already exists in the graph."},
	412:  {412, true, false, "Status code is complemented with error message 'PreconditionFailedException': One of the specified pre-condition is not met. This error is indicative of an optimistic concurrency control violation between reading an edge or vertex and writing it back to the store after modification. Most common situations when this error occurs is property modification, for example g.V('identifier').property('name','value'). Gremlin engine would read the vertex, modify it, and write it back. If there is another traversal running in parallel trying to write the same vertex or an edge, one of them will receive this error. Application should submit traversal to the server again."},
	429:  {429, true, false, "Request was throttled and should be retried after value in x-ms-retry-after-ms"},
	500:  {500, false, false, "Error message that contains 'NotFoundException: Entity with the specified id does not exist in the system.' indicates that a database and/or collection was re-created with the same name. This error will disappear within 5 minutes as change propagates and invalidates caches in different Cosmos DB components. To avoid this issue, use unique database and collection names every time."},
	1000: {1000, false, false, "This status code is returned when server successfully parsed a message but wasn't able to execute. It usually indicates a problem with the query."},
	1001: {1001, false, false, "This code is returned when server completes traversal execution but fails to serialize response back to the client. This error can happen when traversal generates complex result, that is too large or does not conform to TinkerPop protocol specification. Application should simplify the traversal when it encounters this error."},
	1003: {1003, false, false, "'Query exceeded memory limit. Bytes Consumed: XXX, Max: YYY' is returned when traversal exceeds allowed memory limit. Memory limit is 2 GB per traversal."},
	1004: {1004, false, false, "This status code indicates malformed graph request. Request can be malformed when it fails deserialization, non-value type is being deserialized as value type or unsupported gremlin operation requested. Application should not retry the request because it will not be successful."},
	1007: {1007, true, true, "Usually this status code is returned with error message 'Could not process request. Underlying connection has been closed.'. This situation can happen if client driver attempts to use a connection that is being closed by the server. Application should retry the traversal on a different connection."},
	1008: {1008, true, true, "Cosmos DB Gremlin server can terminate connections to rebalance traffic in the cluster. Client drivers should handle this situation and use only live connections to send requests to the server. Occasionally client drivers may not detect that connection was closed. When application encounters an error, 'Connection is too busy. Please retry after sometime or open more connections.' it should retry traversal on a different connection."},
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

// extractFirstError runs through the given responses and returns the first error it finds.
func extractFirstError(responses []interfaces.Response) error {

	for _, response := range responses {
		statusCode := response.Status.Code

		// everything ok --> skip this response
		if statusCode == interfaces.StatusSuccess || statusCode == interfaces.StatusNoContent || statusCode == interfaces.StatusPartialContent {
			continue
		}

		// since all success codes are already skipped, now we have an error

		// For the non 500 error status codes, do the usual error detection mechanism based on the main status code.
		if statusCode != interfaces.StatusServerError {
			return extractError(response)
		}

		// Try to provide a more specific error message for the 500 errors if possible.
		// Usually from CosmosDB we can use additional headers to extract more details.
		responseInfo, err := parseAttributeMap(response.Status.Attributes)
		if err != nil {
			// if we can't parse/ interpret the attribute map then we return the full/ unparsed error information
			return fmt.Errorf("Failed parsing attributes of response: '%s'. Unparsed error: %d - %s", err.Error(), response.Status.Code, response.Status.Message)
		}
		return fmt.Errorf("%d (%d) - %s", responseInfo.statusCode, responseInfo.subStatusCode, responseInfo.statusDescription)

	}

	// no error was found
	return nil
}

func extractRetryConditions(responses []interfaces.Response) (c cosmosRetryInformation) {
	lastRetryResponseStatusCode := noRetry
	retryAfter := time.Second * 0
	for _, response := range responses {
		statusCode := response.Status.Code

		// everything ok --> skip this response
		if statusCode == interfaces.StatusSuccess || statusCode == interfaces.StatusNoContent || statusCode == interfaces.StatusPartialContent {
			continue
		}

		responseInfo, err := parseAttributeMap(response.Status.Attributes)
		if err != nil {
			// if we can't parse/ interpret the attribute map then we ignore it
			continue
		}

		responseCosmosStatusCode,ok := statusCodeDescription[responseInfo.statusCode]

		if !ok {
			// status code not found, ignore it
			continue
		}

		if responseCosmosStatusCode.retry {
			lastRetryResponseStatusCode = responseCosmosStatusCode

			if responseInfo.retryAfter > retryAfter {
				retryAfter = responseInfo.retryAfter
			}
		}
	}
	return cosmosRetryInformation{lastRetryResponseStatusCode, retryAfter}
}

// parseAttributeMap parses the given attribute map assuming that it contains CosmosDB specific headers.
func parseAttributeMap(attributes map[string]interface{}) (responseInformation, error) {
	responseInfo := responseInformation{}

	// immediately return in case the header status code is missing
	if _, ok := attributes[string(headerStatusCode)]; !ok {
		return responseInfo, fmt.Errorf("'%s' is missing", headerStatusCode)
	}

	valueStr := attributes[string(headerStatusCode)]
	value, err := cast.ToInt16E(valueStr)
	if err != nil {
		return responseInfo, errors.Wrapf(err, "Failed parsing '%s'", headerStatusCode)
	}
	statusCode := int(value)
	responseInfo.statusCode = statusCode
	responseInfo.statusDescription = statusCodeToDescription(statusCode)

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
	return desc.responseStatusCodeDescription
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
