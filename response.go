package gremgo

import (
	"encoding/json"
	"fmt"
)

const (
	statusSuccess                  = 200
	statusNoContent                = 204
	statusPartialContent           = 206
	statusUnauthorized             = 401
	statusAuthenticate             = 407
	statusMalformedRequest         = 498
	statusInvalidRequestArguments  = 499
	statusServerError              = 500
	statusScriptEvaluationError    = 597
	statusServerTimeout            = 598
	statusServerSerializationError = 599
)

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

// Response structs holds the entire response from requests to the gremlin server
type Response struct {
	RequestID string `json:"requestId"`
	Status    Status `json:"status"`
	Result    Result `json:"result"`
}

// ToString returns a string representation of the Response struct
func (r Response) ToString() string {
	return fmt.Sprintf("Response \nRequestID: %v, \nStatus: {%#v}, \nResult: {%#v}\n", r.RequestID, r.Status, r.Result)
}

func (c *Client) handleResponse(msg []byte) (err error) {
	resp, err := marshalResponse(msg)

	if resp.Status.Code == statusAuthenticate { //Server request authentication
		return c.authenticate(resp.RequestID)
	}

	c.saveResponse(resp, err)
	return
}

// marshalResponse creates a response struct for every incoming response for further manipulation
func marshalResponse(msg []byte) (resp Response, err error) {
	err = json.Unmarshal(msg, &resp)
	if err != nil {
		return
	}

	err = resp.detectError()
	return
}

// saveResponse makes the response available for retrieval by the requester. Mutexes are used for thread safety.
func (c *Client) saveResponse(resp Response, err error) {
	c.respMutex.Lock()
	var container []interface{}
	existingData, ok := c.results.Load(resp.RequestID) // Retrieve old data container (for requests with multiple responses)
	if ok {
		container = existingData.([]interface{})
	}
	newdata := append(container, resp)       // Create new data container with new data
	c.results.Store(resp.RequestID, newdata) // Add new data to buffer for future retrieval
	respNotifier, load := c.responseNotifier.LoadOrStore(resp.RequestID, make(chan error, 1))
	_ = load
	if resp.Status.Code != statusPartialContent {
		respNotifier.(chan error) <- err
	}
	c.respMutex.Unlock()
}

// retrieveResponse retrieves the response saved by saveResponse.
func (c *Client) retrieveResponse(id string) (data []Response, err error) {
	resp, _ := c.responseNotifier.Load(id)
	err = <-resp.(chan error)
	if err == nil {
		if dataI, ok := c.results.Load(id); ok {
			d := dataI.([]interface{})
			data = make([]Response, len(d))
			for i := range d {
				data[i] = d[i].(Response)
			}
			close(resp.(chan error))
			c.responseNotifier.Delete(id)
			c.deleteResponse(id)
		}
	}
	return
}

// deleteRespones deletes the response from the container. Used for cleanup purposes by requester.
func (c *Client) deleteResponse(id string) {
	c.results.Delete(id)
	return
}

// responseDetectError detects any possible errors in responses from Gremlin Server and generates an error for each code
func (r *Response) detectError() (err error) {
	switch r.Status.Code {
	case statusSuccess, statusNoContent, statusPartialContent:
		break
	case statusUnauthorized:
		err = fmt.Errorf("UNAUTHORIZED - Response Message: %s", r.Status.Message)
	case statusAuthenticate:
		err = fmt.Errorf("AUTHENTICATE - Response Message: %s", r.Status.Message)
	case statusMalformedRequest:
		err = fmt.Errorf("MALFORMED REQUEST - Response Message: %s", r.Status.Message)
	case statusInvalidRequestArguments:
		err = fmt.Errorf("INVALID REQUEST ARGUMENTS - Response Message: %s", r.Status.Message)
	case statusServerError:
		err = fmt.Errorf("SERVER ERROR - Response Message: %s", r.Status.Message)
	case statusScriptEvaluationError:
		err = fmt.Errorf("SCRIPT EVALUATION ERROR - Response Message: %s", r.Status.Message)
	case statusServerTimeout:
		err = fmt.Errorf("SERVER TIMEOUT - Response Message: %s", r.Status.Message)
	case statusServerSerializationError:
		err = fmt.Errorf("SERVER SERIALIZATION ERROR - Response Message: %s", r.Status.Message)
	default:
		err = fmt.Errorf("UNKNOWN ERROR - Response Message: %s", r.Status.Message)
	}
	return
}
