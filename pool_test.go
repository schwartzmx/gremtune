package gremtune

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/schwartzmx/gremtune/interfaces"
	mock_interfaces "github.com/schwartzmx/gremtune/test/mocks/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClose(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)
	clientFactory := func() (interfaces.QueryExecutor, error) {
		return mockedQueryExecutor, nil
	}
	numActiveConnections := 10
	queryExecutor, err := NewPool(clientFactory, numActiveConnections, time.Second*30)
	require.NoError(t, err)
	require.NotNil(t, queryExecutor)
	pool := queryExecutor.(*pool)

	mockedQueryExecutor.EXPECT().HadError().Return(false).AnyTimes()
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
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)
	clientFactory := func() (interfaces.QueryExecutor, error) {
		return mockedQueryExecutor, nil
	}
	queryExecutor, err := NewPool(clientFactory, 10, time.Second*30)
	require.NoError(t, err)
	poolImpl := queryExecutor.(*pool)

	// WHEN
	for i := 0; i < 10; i++ {
		poolImpl.release()
	}

	// THEN
	assert.GreaterOrEqual(t, poolImpl.active, 0)
}

func TestNewPool(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)
	clientFactory := func() (interfaces.QueryExecutor, error) {
		return mockedQueryExecutor, nil
	}
	// WHEN
	queryExecutor, err := NewPool(clientFactory, 10, time.Second*30)

	// THEN
	require.NoError(t, err)
	require.NotNil(t, queryExecutor)
	poolImpl := queryExecutor.(*pool)
	assert.NotNil(t, poolImpl.createQueryExecutor)
	assert.NotNil(t, poolImpl.idleConnections)
}

func TestNewPoolFail(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)
	clientFactory := func() (interfaces.QueryExecutor, error) {
		return mockedQueryExecutor, nil
	}

	// WHEN - create queryexecutor is nil
	pool, err := NewPool(nil, 10, time.Second*30)

	// THEN
	require.Error(t, err)
	require.Nil(t, pool)

	// WHEN - too few connections allowed
	pool, err = NewPool(clientFactory, 0, time.Second*30)

	// THEN
	require.Error(t, err)
	require.Nil(t, pool)

	// WHEN - neg timeout
	pool, err = NewPool(clientFactory, 10, time.Second*-1)

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
	mockedQueryExecutorValid.EXPECT().HadError().Return(false)
	mockedQueryExecutorValid.EXPECT().IsConnected().Return(true)
	mockedQueryExecutorInvalid.EXPECT().HadError().Return(false)
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

	mockedQueryExecutorValid.EXPECT().HadError().Return(false)
	mockedQueryExecutorValid.EXPECT().IsConnected().Return(true)
	// Simulate error
	mockedQueryExecutorWithError.EXPECT().HadError().Return(true)
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

	mockedQueryExecutorValid.EXPECT().HadError().Return(false)
	mockedQueryExecutorValid.EXPECT().IsConnected().Return(true)
	// Simulate error
	mockedQueryExecutorClosed.EXPECT().HadError().Return(false)
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
	mockedQueryExecutor1.EXPECT().HadError().Return(false)
	mockedQueryExecutor1.EXPECT().IsConnected().Return(true)
	mockedQueryExecutor1.EXPECT().Close()
	mockedQueryExecutor2.EXPECT().IsConnected().Return(true)
	mockedQueryExecutor2.EXPECT().HadError().Return(false)
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
