package gremtune

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/schwartzmx/gremtune/interfaces"
)

// Client is a container for the gremtune client.
type Client struct {
	conn                   interfaces.Dialer
	requests               chan []byte
	responses              chan []byte
	results                *sync.Map
	responseNotifier       *sync.Map // responseNotifier notifies the requester that a response has arrived for the request
	responseStatusNotifier *sync.Map // responseStatusNotifier notifies the requester that a response has arrived for the request with the code
	mux                    sync.RWMutex
	Errored                bool

	// auth auth information like username and password
	auth Auth

	// pingInterval is the interval that is used to check if the connection
	// is still alive. The interval to send the ping frame to the peer.
	pingInterval time.Duration

	wg sync.WaitGroup

	// quitChannel channel to notify workers that they should stop
	quitChannel chan struct{}
}

//Auth is the container for authentication data of Client
type Auth struct {
	Username string
	Password string
}

// ClientOption is the struct for defining optional parameters for the Client
type ClientOption func(*Client)

// SetAuth sets credentials for an authenticated connection
func SetAuth(username string, password string) ClientOption {
	return func(c *Client) {
		c.auth = Auth{Username: username, Password: password}
	}
}

// PingInterval sets the ping interval, which is the interval to send the ping frame to the peer
func PingInterval(interval time.Duration) ClientOption {
	return func(c *Client) {
		c.pingInterval = interval
	}
}

func newClient(dialer interfaces.Dialer, options ...ClientOption) *Client {
	client := &Client{
		conn:                   dialer,
		requests:               make(chan []byte, 3), // c.requests takes any request and delivers it to the WriteWorker for dispatch to Gremlin Server
		responses:              make(chan []byte, 3), // c.responses takes raw responses from ReadWorker and delivers it for sorting to handelResponse
		results:                &sync.Map{},
		responseNotifier:       &sync.Map{},
		responseStatusNotifier: &sync.Map{},
		pingInterval:           60 * time.Second,
		quitChannel:            make(chan struct{}),
	}

	for _, opt := range options {
		opt(client)
	}

	return client
}

// Dial returns a gremtune client for interaction with the Gremlin Server specified in the host IP.
func Dial(conn interfaces.Dialer, errorChannel chan error) (*Client, error) {

	if conn == nil {
		return nil, fmt.Errorf("Dialer is nil")
	}
	client := newClient(conn)

	err := client.conn.Connect()
	if err != nil {
		return nil, err
	}

	// Start all worker (run async)
	client.wg.Add(3)
	go client.writeWorker(errorChannel, client.quitChannel)
	go client.readWorker(errorChannel, client.quitChannel)
	go client.pingWorker(errorChannel, client.quitChannel)

	return client, nil
}

func (c *Client) pingWorker(errs chan error, quit <-chan struct{}) {
	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()
	defer c.wg.Done()

	for {
		select {
		case <-ticker.C:
			if err := c.conn.Ping(); err != nil {
				errs <- err
			}
		case <-quit:
			return
		}
	}
}

func (c *Client) executeRequest(query string, bindings, rebindings *map[string]string) ([]Response, error) {
	var req request
	var id string
	var err error

	if bindings != nil && rebindings != nil {
		req, id, err = prepareRequestWithBindings(query, *bindings, *rebindings)
	} else {
		req, id, err = prepareRequest(query)
	}

	if err != nil {
		return nil, err
	}

	msg, err := packageRequest(req)
	if err != nil {
		return nil, err
	}

	c.responseNotifier.Store(id, make(chan error, 1))
	c.responseStatusNotifier.Store(id, make(chan int, 1))
	c.dispatchRequest(msg)

	// this call blocks until the response has been retrieved from the server
	resp, err := c.retrieveResponse(id)

	if err != nil {
		err = errors.Wrapf(err, "query: %s", query)
	}
	return resp, err
}

func (c *Client) executeAsync(query string, bindings, rebindings *map[string]string, responseChannel chan AsyncResponse) (err error) {
	var req request
	var id string
	if bindings != nil && rebindings != nil {
		req, id, err = prepareRequestWithBindings(query, *bindings, *rebindings)
	} else {
		req, id, err = prepareRequest(query)
	}
	if err != nil {
		return
	}

	msg, err := packageRequest(req)
	if err != nil {
		log.Println(err)
		return
	}
	c.responseNotifier.Store(id, make(chan error, 1))
	c.responseStatusNotifier.Store(id, make(chan int, 1))
	c.dispatchRequest(msg)
	go c.retrieveResponseAsync(id, responseChannel)
	return
}

func validateCredentials(auth Auth) error {
	if len(auth.Username) == 0 {
		return fmt.Errorf("Username is missing")
	}

	if len(auth.Password) == 0 {
		return fmt.Errorf("Password is missing")
	}
	return nil
}

func (c *Client) authenticate(requestID string) error {
	if err := validateCredentials(c.auth); err != nil {
		return err
	}

	req, err := prepareAuthRequest(requestID, c.auth.Username, c.auth.Password)
	if err != nil {
		return err
	}

	msg, err := packageRequest(req)
	if err != nil {
		log.Println(err)
		return err
	}

	c.dispatchRequest(msg)
	return nil
}

// ExecuteWithBindings formats a raw Gremlin query, sends it to Gremlin Server, and returns the result.
func (c *Client) ExecuteWithBindings(query string, bindings, rebindings map[string]string) (resp []Response, err error) {
	if c.conn.IsDisposed() {
		return resp, errors.New("you cannot write on disposed connection")
	}
	resp, err = c.executeRequest(query, &bindings, &rebindings)
	return
}

// Execute formats a raw Gremlin query, sends it to Gremlin Server, and returns the result.
func (c *Client) Execute(query string) (resp []Response, err error) {
	if c.conn.IsDisposed() {
		return resp, errors.New("you cannot write on disposed connection")
	}
	resp, err = c.executeRequest(query, nil, nil)
	return
}

// Execute formats a raw Gremlin query, sends it to Gremlin Server, and the results are streamed to channel provided in method paramater.
func (c *Client) ExecuteAsync(query string, responseChannel chan AsyncResponse) (err error) {
	if c.conn.IsDisposed() {
		return errors.New("you cannot write on disposed connection")
	}
	err = c.executeAsync(query, nil, nil, responseChannel)
	return
}

// ExecuteFileWithBindings takes a file path to a Gremlin script, sends it to Gremlin Server with bindings, and returns the result.
func (c *Client) ExecuteFileWithBindings(path string, bindings, rebindings map[string]string) (resp []Response, err error) {
	if c.conn.IsDisposed() {
		return resp, errors.New("you cannot write on disposed connection")
	}
	d, err := ioutil.ReadFile(path) // Read script from file
	if err != nil {
		log.Println(err)
		return
	}
	query := string(d)
	resp, err = c.executeRequest(query, &bindings, &rebindings)
	return
}

// ExecuteFile takes a file path to a Gremlin script, sends it to Gremlin Server, and returns the result.
func (c *Client) ExecuteFile(path string) (resp []Response, err error) {
	if c.conn.IsDisposed() {
		return resp, errors.New("you cannot write on disposed connection")
	}
	d, err := ioutil.ReadFile(path) // Read script from file
	if err != nil {
		log.Println(err)
		return
	}
	query := string(d)
	resp, err = c.executeRequest(query, nil, nil)
	return
}

// Close closes the underlying connection and marks the client as closed.
func (c *Client) Close() error {

	// notify the workers to stop working
	close(c.quitChannel)

	if c.conn == nil {
		return fmt.Errorf("Connection is nil")
	}
	// wait for cleanup of all started go routines
	defer c.wg.Wait()

	return c.conn.Close()
}

// writeWorker works on a loop and dispatches messages as soon as it receives them
func (c *Client) writeWorker(errs chan error, quit <-chan struct{}) {
	defer c.wg.Done()
	for {
		select {
		case msg := <-c.requests:
			c.mux.Lock()
			err := c.conn.Write(msg)
			if err != nil {
				errs <- err
				c.Errored = true
				c.mux.Unlock()
				break
			}
			c.mux.Unlock()

		case <-quit:
			return
		}
	}
}

// readWorker works on a loop and sorts messages as soon as it receives them
func (c *Client) readWorker(errs chan error, quit <-chan struct{}) {
	defer c.wg.Done()
	for {
		msgType, msg, err := c.conn.Read()
		if msgType == -1 { // msgType == -1 is noFrame (close connection)
			errs <- fmt.Errorf("Received msgType == -1 this is no frame --> close the readworker")
			c.Errored = true

			// FIXME: This looks weird. In case a malformed package is sent here the readWorker
			// is just closed. But what happens afterwards? No one is reading any more?!
			// And the connection is not really closed. Hence no reconnect will happen.
			// The only chance would be that the one who consumes the error messages
			// of the error channel closes the connection immediately if an error arrives.
			return
		}

		var errorToPost error
		if err != nil {
			errorToPost = err
		} else if msg == nil {
			errorToPost = fmt.Errorf("Receive message type: %d, but message was nil", msgType)
		} else {
			// handle the message
			errorToPost = c.handleResponse(msg)
		}

		if errorToPost != nil {
			errs <- errorToPost
			c.Errored = true
		}

		select {
		case <-quit:
			return
		default:
			continue
		}
	}
}
