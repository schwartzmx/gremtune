package gremtune

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/schwartzmx/gremtune/interfaces"
	mock_interfaces "github.com/schwartzmx/gremtune/test/mocks/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsConnectedRace(t *testing.T) {
	// This test shall detect data races when
	// checking the connection state of the pool
	// and using the pool at the same time

	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor, pool, err := newMockedPool(mockCtrl)
	require.NoError(t, err)
	require.NotNil(t, pool)

	mockedQueryExecutor.EXPECT().LastError().Return(nil).AnyTimes()
	mockedQueryExecutor.EXPECT().IsConnected().Return(true).AnyTimes()
	mockedQueryExecutor.EXPECT().Close().Return(nil).AnyTimes()

	// WHEN
	// now start a goroutine that checks the connection state
	ticker := time.NewTicker(time.Millisecond * 100)
	go func() {
		for range ticker.C {
			pool.IsConnected()
		}
	}()

	numConnectionsToAcquire := 100
	wg := sync.WaitGroup{}
	wg.Add(numConnectionsToAcquire)

	// start n goroutines that use the pool in parallel
	for i := 0; i < numConnectionsToAcquire; i++ {
		go func() {
			defer wg.Done()
			pc, err := pool.Get()
			require.NoError(t, err)
			require.NotNil(t, pc)
			pc.Close()
			millies := rand.Intn(200)
			time.Sleep(time.Millisecond * time.Duration(millies))
		}()
	}
	wg.Wait()
	ticker.Stop()
	pool.Close()

	// THEN
	// No 'THEN' here. The main use case is to find race conditions,
	// which is done by the golang tooling. This means if there is any
	// race condition the go test -race call will assert this test as failed.
}

func TestIsConnectedNoConnection(t *testing.T) {
	// GIVEN - no connections
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	_, pool, err := newMockedPool(mockCtrl)
	require.NoError(t, err)
	require.NotNil(t, pool)

	// WHEN
	connected := pool.IsConnected()

	// THEN
	assert.False(t, connected)
}

func TestIsConnectedActiveConnection(t *testing.T) {
	// GIVEN - one active connection
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	_, pool, err := newMockedPool(mockCtrl)
	require.NoError(t, err)
	require.NotNil(t, pool)

	// acquire one connection
	pConn, err := pool.Get()
	require.NoError(t, err)
	require.NotNil(t, pConn)

	// WHEN
	connected := pool.IsConnected()

	// THEN
	assert.True(t, connected)
}

func TestIsConnectedIdleConnection(t *testing.T) {
	// GIVEN - one idle connection
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor, pool, err := newMockedPool(mockCtrl)
	require.NoError(t, err)
	require.NotNil(t, pool)

	// acquire one connection
	pConn, err := pool.Get()
	require.NoError(t, err)
	require.NotNil(t, pConn)
	// put back the active connection to the idlepool
	pConn.Close()
	mockedQueryExecutor.EXPECT().IsConnected().Return(true)

	// WHEN
	connected := pool.IsConnected()

	// THEN
	assert.True(t, connected)
}

func TestIsConnectedIdleAndFaulty(t *testing.T) {
	// GIVEN - one idle (connected) and one idle (not connected) connection
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor, pool, err := newMockedPool(mockCtrl)
	require.NoError(t, err)
	require.NotNil(t, pool)

	// acquire two connections
	//mockedQueryExecutor.EXPECT().LastError().Return(nil).Times(1)
	//mockedQueryExecutor.EXPECT().IsConnected().Return(true).Times(2)
	pConn1, err := pool.Get()
	require.NoError(t, err)
	require.NotNil(t, pConn1)
	pConn2, err := pool.Get()
	require.NoError(t, err)
	require.NotNil(t, pConn2)

	// put back the active connections to the idlepool
	pConn1.Close()
	pConn2.Close()
	mockedQueryExecutor.EXPECT().IsConnected().Return(false)
	mockedQueryExecutor.EXPECT().IsConnected().Return(true)

	// WHEN
	connected := pool.IsConnected()

	// THEN
	assert.True(t, connected)
}

func TestClose(t *testing.T) {
	// GIVEN
	logger := zerolog.Nop()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)
	clientFactory := func() (interfaces.QueryExecutor, error) {
		return mockedQueryExecutor, nil
	}
	numActiveConnections := 10
	pool, err := NewPool(clientFactory, numActiveConnections, time.Second*30, logger)
	require.NoError(t, err)
	require.NotNil(t, pool)

	mockedQueryExecutor.EXPECT().LastError().Return(nil).AnyTimes()
	mockedQueryExecutor.EXPECT().IsConnected().Return(true).AnyTimes()
	var pooledConnections []*pooledConnection
	// open n connections
	for i := 0; i < numActiveConnections; i++ {
		pooledConnection, err := pool.Get()
		assert.NotNil(t, pooledConnection)
		assert.NoError(t, err)
		pooledConnections = append(pooledConnections, pooledConnection)
	}
	// close the connections -> generate n idle connections
	for _, pooledConnection := range pooledConnections {
		pooledConnection.Close()
	}

	// WHEN
	mockedQueryExecutor.EXPECT().Close().Return(nil).Times(numActiveConnections)
	err = pool.Close()

	// THEN
	assert.NoError(t, err)
	assert.True(t, pool.closed)
}

func TestRelease(t *testing.T) {
	// GIVEN
	logger := zerolog.Nop()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)
	clientFactory := func() (interfaces.QueryExecutor, error) {
		return mockedQueryExecutor, nil
	}
	pool, err := NewPool(clientFactory, 10, time.Second*30, logger)
	require.NoError(t, err)

	// WHEN
	for i := 0; i < 10; i++ {
		pool.release()
	}

	// THEN
	assert.GreaterOrEqual(t, pool.active, 0)
}

func TestNewPool(t *testing.T) {
	// GIVEN
	logger := zerolog.Nop()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)
	clientFactory := func() (interfaces.QueryExecutor, error) {
		return mockedQueryExecutor, nil
	}

	// WHEN
	pool, err := NewPool(clientFactory, 10, time.Second*30, logger)

	// THEN
	require.NoError(t, err)
	require.NotNil(t, pool)
	assert.NotNil(t, pool.createQueryExecutor)
	assert.NotNil(t, pool.idleConnections)
}

func TestNewPoolFail(t *testing.T) {
	// GIVEN
	logger := zerolog.Nop()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)
	clientFactory := func() (interfaces.QueryExecutor, error) {
		return mockedQueryExecutor, nil
	}

	// WHEN - create queryexecutor is nil
	pool, err := NewPool(nil, 10, time.Second*30, logger)

	// THEN
	require.Error(t, err)
	require.Nil(t, pool)

	// WHEN - too few connections allowed
	pool, err = NewPool(clientFactory, 0, time.Second*30, logger)

	// THEN
	require.Error(t, err)
	require.Nil(t, pool)

	// WHEN - neg timeout
	pool, err = NewPool(clientFactory, 10, time.Second*-1, logger)

	// THEN
	require.Error(t, err)
	require.Nil(t, pool)
}

func TestPurge(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutorInvalid := mock_interfaces.NewMockQueryExecutor(mockCtrl)
	mockedQueryExecutorValid := mock_interfaces.NewMockQueryExecutor(mockCtrl)

	n := time.Now()
	// invalid has timedout and should be cleaned up
	invalid := &idleConnection{idleSince: n.Add(-30 * time.Second), pc: &pooledConnection{client: mockedQueryExecutorInvalid}}
	// valid has not yet timed out and should remain in the idle pool
	valid := &idleConnection{idleSince: n.Add(30 * time.Second), pc: &pooledConnection{client: mockedQueryExecutorValid}}

	// pool has a 30 second timeout and an idle connection slice containing both
	// the invalid and valid idle connections
	p := &pool{idleTimeout: time.Second * 30, idleConnections: []*idleConnection{invalid, valid}}
	assert.Len(t, p.idleConnections, 2, "Expected 2 idle connections")

	// WHEN
	mockedQueryExecutorValid.EXPECT().LastError().Return(nil)
	mockedQueryExecutorValid.EXPECT().IsConnected().Return(true)
	mockedQueryExecutorInvalid.EXPECT().LastError().Return(nil)
	mockedQueryExecutorInvalid.EXPECT().IsConnected().Return(true)
	mockedQueryExecutorInvalid.EXPECT().Close()
	p.purge()

	// THEN
	assert.Len(t, p.idleConnections, 1, "Expected 1 idle connection after purge")
	assert.Equal(t, valid.idleSince, p.idleConnections[0].idleSince, "Expected the valid connection to remain in idle pool")
}

func TestNoPurge(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutorValid := mock_interfaces.NewMockQueryExecutor(mockCtrl)

	n := time.Now()
	// valid has not yet timed out and should remain in the idle pool
	valid := &idleConnection{idleSince: n.Add(-30 * time.Second), pc: &pooledConnection{client: mockedQueryExecutorValid}}

	// pool has a 30 second timeout and an idle connection slice containing both
	// the invalid and valid idle connections
	p := &pool{idleTimeout: 0, idleConnections: []*idleConnection{valid}}
	assert.Len(t, p.idleConnections, 1, "Expected 1 idle connections")

	// WHEN
	p.purge()

	// THEN
	assert.Len(t, p.idleConnections, 1, "Expected 1 idle connection after purge")
	assert.Equal(t, valid.idleSince, p.idleConnections[0].idleSince, "Expected the valid connection to remain in idle pool")
}
func TestPurgeOnErroredConnection(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutorValid := mock_interfaces.NewMockQueryExecutor(mockCtrl)
	mockedQueryExecutorWithError := mock_interfaces.NewMockQueryExecutor(mockCtrl)

	n := time.Now()
	p := &pool{idleTimeout: time.Second * 30}
	valid := &idleConnection{idleSince: n.Add(30 * time.Second), pc: &pooledConnection{client: mockedQueryExecutorValid}}
	closed := &idleConnection{idleSince: n.Add(30 * time.Second), pc: &pooledConnection{pool: p, client: mockedQueryExecutorWithError}}
	idle := []*idleConnection{valid, closed}
	p.idleConnections = idle

	mockedQueryExecutorValid.EXPECT().LastError().Return(nil)
	mockedQueryExecutorValid.EXPECT().IsConnected().Return(true)
	// Simulate error
	mockedQueryExecutorWithError.EXPECT().LastError().Return(fmt.Errorf("ERROR"))
	mockedQueryExecutorWithError.EXPECT().Close().Return(nil)
	assert.Len(t, p.idleConnections, 2, "Expected 2 idle connections")

	// WHEN
	p.purge()

	// THEN
	assert.Len(t, p.idleConnections, 1, "Expected 1 idle connection after purge")
	assert.Equal(t, valid.idleSince, p.idleConnections[0].idleSince, "Expected the valid connection to remain in idle pool")
}

func TestPurgeOnClosedConnection(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutorValid := mock_interfaces.NewMockQueryExecutor(mockCtrl)
	mockedQueryExecutorClosed := mock_interfaces.NewMockQueryExecutor(mockCtrl)

	n := time.Now()
	p := &pool{idleTimeout: time.Second * 30}
	valid := &idleConnection{idleSince: n.Add(30 * time.Second), pc: &pooledConnection{client: mockedQueryExecutorValid}}
	closed := &idleConnection{idleSince: n.Add(30 * time.Second), pc: &pooledConnection{pool: p, client: mockedQueryExecutorClosed}}
	idle := []*idleConnection{valid, closed}
	p.idleConnections = idle

	mockedQueryExecutorValid.EXPECT().LastError().Return(nil)
	mockedQueryExecutorValid.EXPECT().IsConnected().Return(true)
	// Simulate error
	mockedQueryExecutorClosed.EXPECT().LastError().Return(nil)
	mockedQueryExecutorClosed.EXPECT().IsConnected().Return(false)
	assert.Len(t, p.idleConnections, 2, "Expected 2 idle connections")

	// WHEN
	p.purge()

	// THEN
	assert.Len(t, p.idleConnections, 1, "Expected 1 idle connection after purge")
	assert.Equal(t, valid.idleSince, p.idleConnections[0].idleSince, "Expected the valid connection to remain in idle pool")
}

func TestPooledConnectionClose(t *testing.T) {
	// GIVEN
	pool := &pool{}
	pc := &pooledConnection{pool: pool}
	assert.Len(t, pool.idleConnections, 0, "Expected 0 idle connections")

	// WHEN
	pc.Close()

	// THEN
	assert.Len(t, pool.idleConnections, 1, "Expected 1 idle connection")
	idled := pool.idleConnections[0]
	require.NotNil(t, idled, "Expected to get connection")
	assert.False(t, idled.idleSince.IsZero(), "Expected an idled time")
}

func TestFirst(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)

	n := time.Now()
	filledPool := &pool{maxActive: 1, idleTimeout: 30 * time.Second}
	idled := []*idleConnection{
		&idleConnection{idleSince: n.Add(-45 * time.Second), pc: &pooledConnection{pool: filledPool, client: mockedQueryExecutor}}, // expired
		&idleConnection{idleSince: n.Add(-45 * time.Second), pc: &pooledConnection{pool: filledPool, client: mockedQueryExecutor}}, // expired
		&idleConnection{pc: &pooledConnection{pool: filledPool, client: mockedQueryExecutor}},                                      // valid
	}
	filledPool.idleConnections = idled
	assert.Len(t, filledPool.idleConnections, 3, "Expected 3 idle connections")

	// WHEN
	// Get should return the last idle connection and purge the others
	c := filledPool.first()
	assert.Equal(t, c, filledPool.idleConnections[0], "Expected to get first connection in idle slice")
	// Empty pool should return nil
	emptypool := &pool{}
	c = emptypool.first()

	// THEN
	assert.Nil(t, c)
}

func TestGetAndDial(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor1 := mock_interfaces.NewMockQueryExecutor(mockCtrl)
	mockedQueryExecutor2 := mock_interfaces.NewMockQueryExecutor(mockCtrl)

	n := time.Now()
	pool := &pool{idleTimeout: time.Second * 30}
	invalid := &idleConnection{idleSince: n.Add(-30 * time.Second), pc: &pooledConnection{pool: pool, client: mockedQueryExecutor1}}
	idle := []*idleConnection{invalid}
	pool.idleConnections = idle

	pool.createQueryExecutor = func() (interfaces.QueryExecutor, error) {
		return mockedQueryExecutor2, nil
	}

	assert.Len(t, pool.idleConnections, 1, "Expected 1 idle connections")
	assert.Equal(t, invalid, pool.idleConnections[0], "Expected invalid connection")

	// WHEN
	mockedQueryExecutor1.EXPECT().LastError().Return(nil)
	mockedQueryExecutor1.EXPECT().IsConnected().Return(true)
	mockedQueryExecutor1.EXPECT().Close()
	mockedQueryExecutor2.EXPECT().IsConnected().Return(true)
	mockedQueryExecutor2.EXPECT().LastError().Return(nil)
	conn, err := pool.Get()
	assert.NoError(t, err)
	assert.Len(t, pool.idleConnections, 0, "Expected 0 idle connections")
	assert.Equal(t, mockedQueryExecutor1, conn.client, "Expected correct client to be returned")
	assert.Equal(t, 1, pool.active, "Expected 1 active connections")

	// Close the connection and ensure it was returned to the idle pool
	conn.Close()

	assert.Len(t, pool.idleConnections, 1, "Expected connection to be returned to idle pool")
	assert.Equal(t, 0, pool.active, "Expected 0 active connections")

	// Get a new connection and ensure that it is the now idling connection
	conn, err = pool.Get()
	assert.NoError(t, err)
	require.NotNil(t, conn)
	assert.Equal(t, mockedQueryExecutor1, conn.client, "Expected the same connection to be reused")
	assert.Equal(t, 1, pool.active, "Expected 1 active connections")
}

func newMockedPool(mockCtrl *gomock.Controller) (*mock_interfaces.MockQueryExecutor, *pool, error) {
	logger := zerolog.Nop()
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)
	clientFactory := func() (interfaces.QueryExecutor, error) {
		return mockedQueryExecutor, nil
	}
	pool, err := NewPool(clientFactory, 2, time.Second*30, logger)
	return mockedQueryExecutor, pool, err
}
