package gremtune

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mock_interfaces "github.com/schwartzmx/gremtune/test/mocks/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPurge(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)

	n := time.Now()
	// invalid has timedout and should be cleaned up
	invalid := &idleConnection{t: n.Add(-30 * time.Second), pc: &PooledConnection{Client: newClient(mockedDialer)}}
	// valid has not yet timed out and should remain in the idle pool
	valid := &idleConnection{t: n.Add(30 * time.Second), pc: &PooledConnection{Client: newClient(mockedDialer)}}

	// Pool has a 30 second timeout and an idle connection slice containing both
	// the invalid and valid idle connections
	p := &Pool{IdleTimeout: time.Second * 30, idle: []*idleConnection{invalid, valid}}
	assert.Len(t, p.idle, 2, "Expected 2 idle connections")

	// WHEN
	mockedDialer.EXPECT().Close()
	p.purge()

	// THEN
	assert.Len(t, p.idle, 1, "Expected 1 idle connection after purge")
	assert.Equal(t, valid.t, p.idle[0].t, "Expected the valid connection to remain in idle pool")
}

func TestPurgeErrorClosedConnection(t *testing.T) {
	// GIVEN
	n := time.Now()
	p := &Pool{IdleTimeout: time.Second * 30}
	valid := &idleConnection{t: n.Add(30 * time.Second), pc: &PooledConnection{Client: &Client{}}}
	client := &Client{}
	closed := &idleConnection{t: n.Add(30 * time.Second), pc: &PooledConnection{Pool: p, Client: client}}
	idle := []*idleConnection{valid, closed}
	p.idle = idle

	// Simulate error
	closed.pc.Client.Errored = true
	assert.Len(t, p.idle, 2, "Expected 2 idle connections")

	// WHEN
	p.purge()

	// THEN
	assert.Len(t, p.idle, 1, "Expected 1 idle connection after purge")
	assert.Equal(t, valid.t, p.idle[0].t, "Expected the valid connection to remain in idle pool")
}

func TestPooledConnectionClose(t *testing.T) {
	// GIVEN
	pool := &Pool{}
	pc := &PooledConnection{Pool: pool}
	assert.Len(t, pool.idle, 0, "Expected 0 idle connections")

	// WHEN
	pc.Close()

	// THEN
	assert.Len(t, pool.idle, 1, "Expected 1 idle connection")
	idled := pool.idle[0]
	require.NotNil(t, idled, "Expected to get connection")
	assert.False(t, idled.t.IsZero(), "Expected an idled time")
}

func TestFirst(t *testing.T) {
	// GIVEN
	n := time.Now()
	pool := &Pool{MaxActive: 1, IdleTimeout: 30 * time.Second}
	idled := []*idleConnection{
		&idleConnection{t: n.Add(-45 * time.Second), pc: &PooledConnection{Pool: pool, Client: &Client{}}}, // expired
		&idleConnection{t: n.Add(-45 * time.Second), pc: &PooledConnection{Pool: pool, Client: &Client{}}}, // expired
		&idleConnection{pc: &PooledConnection{Pool: pool, Client: &Client{}}},                              // valid
	}
	pool.idle = idled
	assert.Len(t, pool.idle, 3, "Expected 3 idle connections")

	// WHEN
	// Get should return the last idle connection and purge the others
	c := pool.first()
	assert.Equal(t, c, pool.idle[0], "Expected to get first connection in idle slice")
	// Empty pool should return nil
	emptyPool := &Pool{}
	c = emptyPool.first()

	// THEN
	assert.Nil(t, c)
}

func TestGetAndDial(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)

	n := time.Now()
	pool := &Pool{IdleTimeout: time.Second * 30}
	invalid := &idleConnection{t: n.Add(-30 * time.Second), pc: &PooledConnection{Pool: pool, Client: newClient(mockedDialer)}}
	idle := []*idleConnection{invalid}
	pool.idle = idle

	client := newClient(mockedDialer)
	pool.Dial = func() (*Client, error) {
		return client, nil
	}

	assert.Len(t, pool.idle, 1, "Expected 1 idle connections")
	assert.Equal(t, invalid, pool.idle[0], "Expected invalid connection")

	// WHEN
	mockedDialer.EXPECT().Close()
	conn, err := pool.Get()
	assert.NoError(t, err)
	assert.Len(t, pool.idle, 0, "Expected 0 idle connections")
	assert.Equal(t, client, conn.Client, "Expected correct client to be returned")
	assert.Equal(t, 1, pool.active, "Expected 1 active connections")

	// Close the connection and ensure it was returned to the idle pool
	conn.Close()

	assert.Len(t, pool.idle, 1, "Expected connection to be returned to idle pool")
	assert.Equal(t, 0, pool.active, "Expected 0 active connections")

	// Get a new connection and ensure that it is the now idling connection
	conn, err = pool.Get()
	assert.NoError(t, err)
	require.NotNil(t, conn)
	assert.Equal(t, client, conn.Client, "Expected the same connection to be reused")
	assert.Equal(t, 1, pool.active, "Expected 1 active connections")
}
