package interfaces

import (
	"encoding/json"
	"fmt"
)

type QueryExecutor interface {
	HadError() bool
	Close() error
	Execute(query string) (resp []Response, err error)
	IsConnected() bool
	ExecuteAsync(query string, responseChannel chan AsyncResponse) (err error)
	ExecuteFileWithBindings(path string, bindings, rebindings map[string]string) (resp []Response, err error)
	ExecuteFile(path string) (resp []Response, err error)
	ExecuteWithBindings(query string, bindings, rebindings map[string]string) (resp []Response, err error)
}

const (
	StatusSuccess                  = 200
	StatusNoContent                = 204
	StatusPartialContent           = 206
	StatusUnauthorized             = 401
	StatusAuthenticate             = 407
	StatusMalformedRequest         = 498
	StatusInvalidRequestArguments  = 499
	StatusServerError              = 500
	StatusScriptEvaluationError    = 597
	StatusServerTimeout            = 598
	StatusServerSerializationError = 599
)

// Response structs holds the entire response from requests to the gremlin server
type Response struct {
	RequestID string `json:"requestId"`
	Status    Status `json:"status"`
	Result    Result `json:"result"`
}

// Status struct is used to hold properties returned from requests to the gremlin server
type Status struct {
	Message    string                 `json:"message"`
	Code       int                    `json:"code"`
	Attributes map[string]interface{} `json:"attributes"`
}

// Result struct is used to hold properties returned for results from requests to the gremlin server
type Result struct {
	// Query Response Data
	Data json.RawMessage        `json:"data"`
	Meta map[string]interface{} `json:"meta"`
}

// AsyncResponse structs holds the entire response from requests to the gremlin server
type AsyncResponse struct {
	Response     Response `json:"response"`     //Partial Response object
	ErrorMessage string   `json:"errorMessage"` // Error message if there was an error
}

// String returns a string representation of the Response struct
func (r Response) String() string {
	return fmt.Sprintf("Response \nRequestID: %v, \nStatus: {%#v}, \nResult: {%#v}\n", r.RequestID, r.Status, r.Result)
}

// DetectError detects any possible errors in responses from Gremlin Server and generates an error for each code
func (r *Response) DetectError() (err error) {
	switch r.Status.Code {
	case StatusSuccess, StatusNoContent, StatusPartialContent:
		break
	case StatusUnauthorized:
		err = fmt.Errorf("UNAUTHORIZED - Response Message: %s", r.Status.Message)
	case StatusAuthenticate:
		err = fmt.Errorf("AUTHENTICATE - Response Message: %s", r.Status.Message)
	case StatusMalformedRequest:
		err = fmt.Errorf("MALFORMED REQUEST - Response Message: %s", r.Status.Message)
	case StatusInvalidRequestArguments:
		err = fmt.Errorf("INVALID REQUEST ARGUMENTS - Response Message: %s", r.Status.Message)
	case StatusServerError:
		err = fmt.Errorf("SERVER ERROR - Response Message: %s", r.Status.Message)
	case StatusScriptEvaluationError:
		err = fmt.Errorf("SCRIPT EVALUATION ERROR - Response Message: %s", r.Status.Message)
	case StatusServerTimeout:
		err = fmt.Errorf("SERVER TIMEOUT - Response Message: %s", r.Status.Message)
	case StatusServerSerializationError:
		err = fmt.Errorf("SERVER SERIALIZATION ERROR - Response Message: %s", r.Status.Message)
	default:
		err = fmt.Errorf("UNKNOWN ERROR - Response Message: %s", r.Status.Message)
	}
	return
}
