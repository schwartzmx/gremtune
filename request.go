package gremcos

import (
	"encoding/base64"
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// MimeType used for communication with the gremlin server.
var MimeType = []byte("application/vnd.gremlin-v2.0+json")

// request is a container for all evaluation request parameters to be sent to the Gremlin Server.
type request struct {
	RequestID string                 `json:"requestId"`
	Op        string                 `json:"op"`
	Processor string                 `json:"processor"`
	Args      map[string]interface{} `json:"args"`
}

// prepareRequest packages a query and binding into the format that Gremlin Server accepts
func prepareRequest(query string) (request, string, error) {
	var uuID uuid.UUID
	uuID, err := uuid.NewV4()
	if err != nil {
		return request{}, "", err
	}

	req := request{}
	req.RequestID = uuID.String()
	req.Op = "eval"
	req.Processor = ""

	req.Args = make(map[string]interface{})
	req.Args["language"] = "gremlin-groovy"
	req.Args["gremlin"] = query

	return req, req.RequestID, nil
}

// prepareRequest packages a query and binding into the format that Gremlin Server accepts
func prepareRequestWithBindings(query string, bindings, rebindings map[string]interface{}) (request, string, error) {
	uuID, err := uuid.NewV4()
	if err != nil {
		return request{}, "", err
	}

	req := request{}
	req.RequestID = uuID.String()
	req.Op = "eval"
	req.Processor = ""

	req.Args = make(map[string]interface{})
	req.Args["language"] = "gremlin-groovy"
	req.Args["gremlin"] = query
	req.Args["bindings"] = bindings
	req.Args["rebindings"] = rebindings

	return req, req.RequestID, nil
}

//prepareAuthRequest creates a ws request for Gremlin Server
func prepareAuthRequest(requestID string, username string, password string) request {
	req := request{}
	req.RequestID = requestID
	req.Op = "authentication"
	req.Processor = "traversal"

	var simpleAuth []byte
	user := []byte(username)
	pass := []byte(password)

	simpleAuth = append(simpleAuth, 0)
	simpleAuth = append(simpleAuth, user...)
	simpleAuth = append(simpleAuth, 0)
	simpleAuth = append(simpleAuth, pass...)

	req.Args = make(map[string]interface{})
	req.Args["sasl"] = base64.StdEncoding.EncodeToString(simpleAuth)

	return req
}

// formatMessage takes a request type and formats it into being able to be delivered to Gremlin Server
func packageRequest(req request) ([]byte, error) {
	j, err := json.Marshal(req) // Formats request into byte format
	if err != nil {
		return nil, errors.Wrap(err, "marshalling request")
	}
	lenMimeType := byte(len(MimeType))

	//lenMimeType is the fixed length of mimeType in hex
	msg := append([]byte{lenMimeType}, MimeType...)
	msg = append(msg, j...)

	return msg, nil
}

// dispatchRequest sends the request for writing to the remote Gremlin Server
func (c *client) dispatchRequest(msg []byte) {
	c.requests <- msg
}
