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

// client is a container for the gremtune client.
type client struct {
	// conn is the entity that manages the websocket connection
	conn interfaces.Dialer

	// requests takes any request and delivers it to the WriteWorker for dispatch to Gremlin Server
	requests chan []byte

	// results is a container for the responses received from the peer.
	// The responses are stored per request id.
	// For each request (Id) there might be 0..n responses.
	// <RequestID string,responses []Response>
	results *sync.Map

	// responseNotifier notifies the requester that a response has arrived for a specific request
	// <RequestID string,errorChannel chan error>
	// As notification object error is used.
	// In case of an error an error is sent on the channel.
	// !!! In case there is new (unprocessed) data available, nil is sent on the channel.
	responseNotifier *sync.Map

	// responseStatusNotifier notifies the requester that a response has arrived for a specific request with a specific (http like) status code
	// <RequestID string,codeChannel chan int>
	responseStatusNotifier *sync.Map

	// stores the most recent error
	lastError error

	// auth auth information like username and password
	auth auth

	// pingInterval is the interval that is used to check if the connection
	// is still alive. The interval to send the ping frame to the peer.
	pingInterval time.Duration

	wg  sync.WaitGroup
	mux sync.RWMutex

	// quitChannel channel to notify workers that they should stop
	quitChannel chan struct{}
}

// auth is the container for authentication data of Client
type auth struct {
	username string
	password string
}

// clientOption is the struct for defining optional parameters for the Client
type clientOption func(*client)

// SetAuth sets credentials for an authenticated connection
func SetAuth(username string, password string) clientOption {
	return func(c *client) {
		c.auth = auth{username: username, password: password}
	}
}

// PingInterval sets the ping interval, which is the interval to send the ping frame to the peer
func PingInterval(interval time.Duration) clientOption {
	return func(c *client) {
		c.pingInterval = interval
	}
}

func newClient(dialer interfaces.Dialer, options ...clientOption) *client {
	client := &client{
		conn:                   dialer,
		requests:               make(chan []byte, 3),
		results:                &sync.Map{},
		responseNotifier:       &sync.Map{},
		responseStatusNotifier: &sync.Map{},
		pingInterval:           60 * time.Second,
		quitChannel:            make(chan struct{}),
		lastError:              nil,
	}

	for _, opt := range options {
		opt(client)
	}

	return client
}

// Dial returns a client for interaction with the Gremlin Server specified in the host IP.
// The client is already connected.
func Dial(conn interfaces.Dialer, errorChannel chan error, options ...clientOption) (*client, error) {

	if conn == nil {
		return nil, fmt.Errorf("Dialer is nil")
	}
	client := newClient(conn, options...)

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

func (c *client) LastError() error {
	return c.lastError
}

func (c *client) IsConnected() bool {
	return c.conn.IsConnected()
}

func (c *client) executeRequest(query string, bindings, rebindings *map[string]string) ([]interfaces.Response, error) {
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

func (c *client) executeAsync(query string, bindings, rebindings *map[string]string, responseChannel chan interfaces.AsyncResponse) (err error) {
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

func validateCredentials(auth auth) error {
	if len(auth.username) == 0 {
		return fmt.Errorf("Username is missing")
	}

	if len(auth.password) == 0 {
		return fmt.Errorf("Password is missing")
	}
	return nil
}

func (c *client) authenticate(requestID string) error {
	if err := validateCredentials(c.auth); err != nil {
		return err
	}

	req, err := prepareAuthRequest(requestID, c.auth.username, c.auth.password)
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
func (c *client) ExecuteWithBindings(query string, bindings, rebindings map[string]string) (resp []interfaces.Response, err error) {
	if !c.conn.IsConnected() {
		return resp, fmt.Errorf("Can't write - no connection")
	}
	resp, err = c.executeRequest(query, &bindings, &rebindings)
	return
}

// Execute formats a raw Gremlin query, sends it to Gremlin Server, and returns the result.
func (c *client) Execute(query string) (resp []interfaces.Response, err error) {
	if !c.conn.IsConnected() {
		return resp, fmt.Errorf("Can't write - no connection")
	}
	resp, err = c.executeRequest(query, nil, nil)
	return
}

// Execute formats a raw Gremlin query, sends it to Gremlin Server, and the results are streamed to channel provided in method paramater.
func (c *client) ExecuteAsync(query string, responseChannel chan interfaces.AsyncResponse) (err error) {
	if !c.conn.IsConnected() {
		return fmt.Errorf("Can't write - no connection")
	}
	err = c.executeAsync(query, nil, nil, responseChannel)
	return
}

// ExecuteFileWithBindings takes a file path to a Gremlin script, sends it to Gremlin Server with bindings, and returns the result.
func (c *client) ExecuteFileWithBindings(path string, bindings, rebindings map[string]string) (resp []interfaces.Response, err error) {
	if !c.conn.IsConnected() {
		return resp, fmt.Errorf("Can't write - no connection")
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
func (c *client) ExecuteFile(path string) (resp []interfaces.Response, err error) {
	if !c.conn.IsConnected() {
		return resp, fmt.Errorf("Can't write - no connection")
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
func (c *client) Close() error {

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
func (c *client) writeWorker(errs chan error, quit <-chan struct{}) {
	defer c.workerSaveExit("writeWorker", errs)

	for {
		select {
		case msg := <-c.requests:
			c.mux.Lock()
			err := c.conn.Write(msg)
			if err != nil {
				errs <- err
				c.lastError = err
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
func (c *client) readWorker(errs chan error, quit <-chan struct{}) {
	defer c.workerSaveExit("readWorker", errs)

	for {
		msgType, msg, err := c.conn.Read()
		if msgType == -1 { // msgType == -1 is noFrame (close connection)
			err = fmt.Errorf("Received msgType == -1 this is no frame --> close the readworker")
			errs <- err
			c.lastError = err

			// to return at this point is save since we call workerSaveExit() to clean up everything
			// when the function is left
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
			c.lastError = errorToPost
		}

		select {
		case <-quit:
			return
		default:
			continue
		}
	}
}

func (c *client) pingWorker(errs chan error, quit <-chan struct{}) {
	ticker := time.NewTicker(c.pingInterval)
	defer func() {
		ticker.Stop()
		c.workerSaveExit("pingWorker", errs)
	}()

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

// workerSaveExit can be used as defered call on leaving a worker routine.
// It ensures that the client is closed and cleaned up appropriately.
func (c *client) workerSaveExit(name string, errs chan<- error) {
	c.wg.Done()

	// call close to ensure that everything is cleaned up appropriately
	if err := c.Close(); err != nil {
		err = fmt.Errorf("Error closing client while leaving worker '%s'", name)
		errs <- err
	}
}
