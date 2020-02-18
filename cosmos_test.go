package gremcos

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_interfaces "github.com/supplyon/gremcos/test/mocks/interfaces"
)

func TestNew(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)
	idleTimeout := time.Second * 12
	maxActiveConnections := 10
	username := "abcd"
	password := "xyz"

	// WHEN
	cosmos, err := New("ws://host",
		ConnectionIdleTimeout(idleTimeout),
		NumMaxActiveConnections(maxActiveConnections),
		WithAuth(username, password),
		withMetrics(metrics),
	)

	// THEN
	require.NoError(t, err)
	require.NotNil(t, cosmos)
	assert.Equal(t, idleTimeout, cosmos.connectionIdleTimeout)
	assert.Equal(t, maxActiveConnections, cosmos.numMaxActiveConnections)
	assert.Equal(t, username, cosmos.username)
	assert.Equal(t, password, cosmos.password)
}

func TestStop(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)

	cosmos, err := New("ws://host", withMetrics(metrics))
	require.NoError(t, err)
	require.NotNil(t, cosmos)
	cosmos.pool = mockedQueryExecutor
	mockedQueryExecutor.EXPECT().Close().Return(nil)

	// WHEN
	err = cosmos.Stop()

	// THEN
	assert.NoError(t, err)
}

func TestIsHealthy(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)

	cosmos, err := New("ws://host", withMetrics(metrics))
	require.NoError(t, err)
	require.NotNil(t, cosmos)
	cosmos.pool = mockedQueryExecutor

	// WHEN -- connected --> healthy
	mockedQueryExecutor.EXPECT().Ping().Return(nil)
	healthyWhenConnected := cosmos.IsHealthy()

	// WHEN -- not connected --> not healthy
	mockedQueryExecutor.EXPECT().Ping().Return(fmt.Errorf("Not connected"))
	healthyWhenNotConnected := cosmos.IsHealthy()

	// THEN
	assert.NoError(t, healthyWhenConnected)
	assert.Error(t, healthyWhenNotConnected)
}

func TestNewWithMetrics(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// WHEN
	cosmos, err := New("ws://host", MetricsPrefix("prefix"))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, cosmos)
	assert.NotNil(t, cosmos.metrics)
}
