package gremtune

import (
	"fmt"
	"sync"
	"time"

	"github.com/schwartzmx/gremtune/interfaces"
)

type QueryExecutorFactoryFunc func() (interfaces.QueryExecutor, error)

// pool maintains a pool of connections to the cosmos db.
type pool struct {
	// createQueryExecutor function that returns new connected QueryExecutors
	createQueryExecutor QueryExecutorFactoryFunc

	// maxActive is the maximum number of allowed active connections
	maxActive int

	// idleTimeout is the maximum time a idle connection will be kept in the pool.
	// If the timeout has been expired, the connection will be closed and removed
	// from the pool.
	// If this timeout is set to 0, the timeout is unlimited -> no expiration of connections.
	idleTimeout time.Duration

	// idleConnections list of idle connections
	idleConnections []*idleConnection

	// active is the number of currently active connections
	active int

	closed bool
	cond   *sync.Cond
	mu     sync.Mutex
}

// pooledConnection represents a shared and reusable connection.
type pooledConnection struct {
	pool   *pool
	client interfaces.QueryExecutor
}

// NewPool creates a new pool which is a QueryExecutor
func NewPool(createQueryExecutor QueryExecutorFactoryFunc, maxActiveConnections int, idleTimeout time.Duration) (*pool, error) {

	if createQueryExecutor == nil {
		return nil, fmt.Errorf("Given createQueryExecutor is nil")
	}

	if maxActiveConnections < 1 {
		return nil, fmt.Errorf("maxActiveConnections has to be >=1")
	}

	if idleTimeout < time.Second*0 {
		return nil, fmt.Errorf("maxActiveConnections has to be >=0")
	}

	return &pool{
		createQueryExecutor: createQueryExecutor,
		maxActive:           maxActiveConnections,
		active:              0,
		closed:              false,
		idleTimeout:         idleTimeout,
		idleConnections:     make([]*idleConnection, 0),
	}, nil
}

type idleConnection struct {
	pc *pooledConnection

	// idleSince is the time the connection was idled
	idleSince time.Time
}

func (p *pool) IsConnected() bool {
	// TODO: Implement
	return true
}

func (p *pool) LastError() error {
	// TODO: Implement
	return nil
}

// Get will return an available pooled connection. Either an idle connection or
// by dialing a new one if the pool does not currently have a maximum number
// of active connections.
func (p *pool) Get() (*pooledConnection, error) {
	// Lock the pool to keep the kids out.
	p.mu.Lock()

	// Clean this place up.
	p.purge()

	// Wait loop
	for {
		// TODO: Ensure to return only clients that are connected

		// Try to grab first available idle connection
		if conn := p.first(); conn != nil {

			// Remove the connection from the idle slice
			p.idleConnections = append(p.idleConnections[:0], p.idleConnections[1:]...)
			p.active++
			p.mu.Unlock()
			pc := &pooledConnection{pool: p, client: conn.pc.client}
			return pc, nil

		}

		// No idle connections, try dialing a new one
		if p.maxActive == 0 || p.active < p.maxActive {
			p.active++
			createQueryExecutor := p.createQueryExecutor

			// Unlock here so that any other connections that need to be
			// dialed do not have to wait.
			p.mu.Unlock()

			dc, err := createQueryExecutor()
			if err != nil {
				p.mu.Lock()
				p.release()
				p.mu.Unlock()
				return nil, err
			}

			pc := &pooledConnection{pool: p, client: dc}
			return pc, nil
		}

		//No idle connections and max active connections, let's wait.
		if p.cond == nil {
			p.cond = sync.NewCond(&p.mu)
		}

		p.cond.Wait()
	}
}

// put pushes the supplied pooledConnection to the top of the idle slice to be reused.
// It is not threadsafe. The caller should manage locking the pool.
func (p *pool) put(pc *pooledConnection) {
	if p.closed {
		pc.client.Close()
		return
	}
	idle := &idleConnection{pc: pc, idleSince: time.Now()}
	// Prepend the connection to the front of the slice
	p.idleConnections = append([]*idleConnection{idle}, p.idleConnections...)

}

// purge removes expired idle connections from the pool.
// It is not threadsafe. The caller should manage locking the pool.
func (p *pool) purge() {
	timeout := p.idleTimeout
	// don't clean up in case there is no timeout specified
	if timeout <= 0 {
		return
	}

	var idleConnectionsAfterPurge []*idleConnection
	now := time.Now()
	for _, idleConnection := range p.idleConnections {
		// If the client has an error then exclude it from the pool
		if err := idleConnection.pc.client.LastError(); err != nil {
			// TODO: Print error to log

			// Force underlying connection closed
			idleConnection.pc.client.Close()
			continue
		}

		// If the client is not connected any more then exclude it from the pool
		if !idleConnection.pc.client.IsConnected() {
			continue
		}

		if idleConnection.idleSince.Add(timeout).After(now) {
			// not expired -> keep it in the idle connection list
			idleConnectionsAfterPurge = append(idleConnectionsAfterPurge, idleConnection)
		} else {
			// expired -> don't add it to the idle connection list
			// Force underlying connection closed
			idleConnection.pc.client.Close()
		}
	}
	p.idleConnections = idleConnectionsAfterPurge
}

// release decrements active and alerts waiters.
// It is not threadsafe. The caller should manage locking the pool.
func (p *pool) release() {
	if p.closed {
		return
	}

	// can't release a more connections
	// since there are no active ones any more
	if p.active == 0 {
		return
	}

	p.active--

	if p.cond != nil {
		p.cond.Signal()
	}
}

// It is not threadsafe. The caller should manage locking the pool.
func (p *pool) first() *idleConnection {
	if len(p.idleConnections) == 0 {
		return nil
	}
	return p.idleConnections[0]
}

// Close closes the pool.
func (p *pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, c := range p.idleConnections {
		c.pc.client.Close()
	}

	p.closed = true
	return nil
}

// ExecuteWithBindings formats a raw Gremlin query, sends it to Gremlin Server, and returns the result.
func (p *pool) ExecuteWithBindings(query string, bindings, rebindings map[string]string) (resp []interfaces.Response, err error) {
	pc, err := p.Get()
	if err != nil {
		return nil, err
	}
	defer pc.Close()
	return pc.client.ExecuteWithBindings(query, bindings, rebindings)
}

// Execute grabs a connection from the pool, formats a raw Gremlin query, sends it to Gremlin Server, and returns the result.
func (p *pool) Execute(query string) (resp []interfaces.Response, err error) {
	pc, err := p.Get()
	if err != nil {
		return nil, err
	}
	// put the connection back into the idle pool
	defer pc.Close()

	return pc.client.Execute(query)
}

func (p *pool) ExecuteAsync(query string, responseChannel chan interfaces.AsyncResponse) (err error) {
	pc, err := p.Get()
	if err != nil {
		return err
	}
	// put the connection back into the idle pool
	defer pc.Close()

	return pc.client.ExecuteAsync(query, responseChannel)
}

func (p *pool) ExecuteFile(path string) (resp []interfaces.Response, err error) {
	pc, err := p.Get()
	if err != nil {
		return nil, err
	}
	// put the connection back into the idle pool
	defer pc.Close()

	return pc.client.ExecuteFile(path)
}

func (p *pool) ExecuteFileWithBindings(path string, bindings, rebindings map[string]string) (resp []interfaces.Response, err error) {
	pc, err := p.Get()
	if err != nil {
		return nil, err
	}
	// put the connection back into the idle pool
	defer pc.Close()

	return pc.client.ExecuteFileWithBindings(path, bindings, rebindings)
}

// Close signals that the caller is finished with the connection and should be
// returned to the pool for future use.
func (pc *pooledConnection) Close() {
	pc.pool.mu.Lock()
	defer pc.pool.mu.Unlock()

	pc.pool.put(pc)
	pc.pool.release()
}
