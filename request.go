package gremtune

import (
	"encoding/base64"
	"encoding/json"

	"github.com/gofrs/uuid"
)

type requester interface {
	prepare() error
	getID() string
	getRequest() request
}

// request is a container for all evaluation request parameters to be sent to the Gremlin Server.
type request struct {
	RequestID string                 `json:"requestId"`
	Op        string                 `json:"op"`
	Processor string                 `json:"processor"`
	Args      map[string]interface{} `json:"args"`
}

// prepareRequest packages a query and binding into the format that Gremlin Server accepts
func prepareRequest(query string) (req request, id string, err error) {
	var uuID uuid.UUID
	uuID, _ = uuid.NewV4()
	id = uuID.String()

	req.RequestID = id
	req.Op = "eval"
	req.Processor = ""

	req.Args = make(map[string]interface{})
	req.Args["language"] = "gremlin-groovy"
	req.Args["gremlin"] = query

	return
}

// prepareRequest packages a query and binding into the format that Gremlin Server accepts
func prepareRequestWithBindings(query string, bindings, rebindings map[string]string) (req request, id string, err error) {
	var uuID uuid.UUID
	uuID, _ = uuid.NewV4()
	id = uuID.String()

	req.RequestID = id
	req.Op = "eval"
	req.Processor = ""

	req.Args = make(map[string]interface{})
	req.Args["language"] = "gremlin-groovy"
	req.Args["gremlin"] = query
	req.Args["bindings"] = bindings
	req.Args["rebindings"] = rebindings
	return
}

// prepareRequestWithSession packages a query and sessionID into the format that Gremlin Server accepts
func prepareRequestWithSession(query string, sessionID string) (req request, id string, err error) {

	if len(sessionID) > 0 {
		var uuID uuid.UUID
		uuID, _ = uuid.NewV4()
		id = uuID.String()

		req.RequestID = id
		req.Op = "eval"
		req.Processor = "session"
		req.Args = make(map[string]interface{})
		req.Args["language"] = "gremlin-groovy"
		req.Args["gremlin"] = query
		req.Args["manageTransaction"] = false
		req.Args["session"] = sessionID
		req.Args["batchSize"] = 64
	} else {
		req, id, err = prepareRequest(query)
	}
	return
}

// prepareRequestWithSessionAndTimeout packages a query and sessionID into the format that Gremlin Server accepts
func prepareRequestWithSessionAndTimeout(query string, sessionID string, timeout int) (req request, id string, err error) {

	if len(sessionID) > 0 && timeout > 0 {
		var uuID uuid.UUID
		uuID, _ = uuid.NewV4()
		id = uuID.String()

		req.RequestID = id
		req.Op = "eval"
		req.Processor = "session"
		req.Args = make(map[string]interface{})
		req.Args["language"] = "gremlin-groovy"
		req.Args["gremlin"] = query
		req.Args["manageTransaction"] = false
		req.Args["session"] = sessionID
		req.Args["batchSize"] = 64
		req.Args["scriptEvaluationTimeout"] = timeout

	} else if len(sessionID) <= 0 && timeout > 0 {
		var uuID uuid.UUID
		uuID, _ = uuid.NewV4()
		id = uuID.String()

		req.RequestID = id
		req.Op = "eval"
		req.Processor = ""

		req.Args = make(map[string]interface{})
		req.Args["language"] = "gremlin-groovy"
		req.Args["gremlin"] = query
		req.Args["scriptEvaluationTimeout"] = timeout

	} else if len(sessionID) > 0 && timeout <= 0 {
		req, id, err = prepareRequestWithSession(query, sessionID)

	} else {
		req, id, err = prepareRequest(query)
	}
	return
}

// prepareRequest packages a query and binding into the format that Gremlin Server accepts
func prepareCommitSessionRequest(sessionID string) (req request, id string, err error) {

	if len(sessionID) > 0 {
		var uuID uuid.UUID
		uuID, _ = uuid.NewV4()
		id = uuID.String()

		req.RequestID = id
		req.Op = "close"
		req.Processor = "session"
		req.Args = make(map[string]interface{})
		req.Args["language"] = "gremlin-groovy"
		req.Args["manageTransaction"] = false
		req.Args["session"] = sessionID
		req.Args["force"] = false
	} else {
		req, id, err = prepareRequest("")
	}
	return
}

//prepareAuthRequest creates a ws request for Gremlin Server
func prepareAuthRequest(requestID string, username string, password string) (req request, err error) {
	req.RequestID = requestID
	req.Op = "authentication"
	req.Processor = "trasversal"

	var simpleAuth []byte
	user := []byte(username)
	pass := []byte(password)

	simpleAuth = append(simpleAuth, 0)
	simpleAuth = append(simpleAuth, user...)
	simpleAuth = append(simpleAuth, 0)
	simpleAuth = append(simpleAuth, pass...)

	req.Args = make(map[string]interface{})
	req.Args["sasl"] = base64.StdEncoding.EncodeToString(simpleAuth)

	return
}

// formatMessage takes a request type and formats it into being able to be delivered to Gremlin Server
func packageRequest(req request) (msg []byte, err error) {
	j, err := json.Marshal(req) // Formats request into byte format
	if err != nil {
		return
	}
	mimeType := []byte("application/vnd.gremlin-v3.0+json")
	msg = append([]byte{0x21}, mimeType...) //0x21 is the fixed length of mimeType in hex
	msg = append(msg, j...)

	return
}

// dispactchRequest sends the request for writing to the remote Gremlin Server
func (c *Client) dispatchRequest(msg []byte) {
	c.requests <- msg
}
