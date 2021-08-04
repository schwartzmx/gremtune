package gremcos

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supplyon/gremcos/interfaces"
	mock_interfaces "github.com/supplyon/gremcos/test/mocks/interfaces"
	mock_metrics "github.com/supplyon/gremcos/test/mocks/metrics"
)

type dialerMock struct {
}

func (d *dialerMock) Connect() error {
	return nil
}
func (d *dialerMock) IsConnected() bool {
	return true
}
func (d *dialerMock) Write([]byte) error {
	return nil
}
func (d *dialerMock) Read() (int, []byte, error) {
	return 1, nil, nil
}
func (d *dialerMock) Close() error {
	return nil
}
func (d *dialerMock) Ping() error {
	return nil
}

var websocketGenerator = func(host string, options ...optionWebsocket) (interfaces.Dialer, error) {
	mock := &dialerMock{}
	return mock, nil
}

func toCosmosImpl(t *testing.T, cosmos Cosmos) *cosmosImpl {
	require.NotNil(t, cosmos, "Cosmos must not be nil")
	cImpl, ok := cosmos.(*cosmosImpl)
	require.True(t, ok, "Failed to cast to *cosmosImpl")
	require.NotNil(t, cImpl, "Casted to nil of cosmosImpl")
	return cImpl
}

func TestDialUsingDifferentWebsockets(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)
	idleTimeout := time.Second * 12
	maxActiveConnections := 10
	username := "abcd"
	password := "xyz"

	cosmos, err := New("ws://host",
		ConnectionIdleTimeout(idleTimeout),
		NumMaxActiveConnections(maxActiveConnections),
		WithAuth(username, password),
		withMetrics(metrics),
		wsGenerator(websocketGenerator),
	)
	require.NoError(t, err)
	cImpl := toCosmosImpl(t, cosmos)

	// WHEN
	queryExecutor1, err1 := cImpl.dial()
	queryExecutor2, err2 := cImpl.dial()

	// THEN
	require.NoError(t, err1)
	require.NotNil(t, queryExecutor1)
	require.NoError(t, err2)
	require.NotNil(t, queryExecutor2)

	client1 := queryExecutor1.(*client)
	client2 := queryExecutor2.(*client)
	assert.False(t, &client1.conn == &client2.conn)
}

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
	cImpl := toCosmosImpl(t, cosmos)
	assert.Equal(t, idleTimeout, cImpl.connectionIdleTimeout)
	assert.Equal(t, maxActiveConnections, cImpl.numMaxActiveConnections)
	assert.Equal(t, username, cImpl.credentialProvider.Username())
	assert.Equal(t, password, cImpl.credentialProvider.Password())
}

func TestStop(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)

	cosmos, err := New("ws://host", withMetrics(metrics))
	require.NoError(t, err)
	cImpl := toCosmosImpl(t, cosmos)
	cImpl.pool = mockedQueryExecutor
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
	cImpl := toCosmosImpl(t, cosmos)
	cImpl.pool = mockedQueryExecutor

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
	cImpl := toCosmosImpl(t, cosmos)
	assert.NotNil(t, cImpl.metrics)
}

func TestUpdateMetricsNoResponses(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)

	var responses []interfaces.Response

	// WHEN
	updateRequestMetrics(responses, metrics)

	// THEN
	// there should be no invocation on the metrics mock
	// since there where no responses
}

func TestUpdateMetricsZero(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, metricMocks := NewMockedMetrics(mockCtrl)

	noErrorNoAttrs := interfaces.Response{
		Status: interfaces.Status{
			Code: interfaces.StatusSuccess,
		},
	}

	noError := interfaces.Response{
		Status: interfaces.Status{
			Code: interfaces.StatusSuccess,
			Attributes: map[string]interface{}{
				"x-ms-status-code": 200,
			},
		},
	}

	responses := []interfaces.Response{noError, noErrorNoAttrs}

	// WHEN
	mockCount200 := mock_metrics.NewMockCounter(mockCtrl)
	mockCount200.EXPECT().Inc().Times(2)
	metricMocks.statusCodeTotal.EXPECT().WithLabelValues("200").Return(mockCount200).Times(2)
	metricMocks.serverTimePerQueryResponseAvgMS.EXPECT().Set(float64(0))
	metricMocks.serverTimePerQueryMS.EXPECT().Set(float64(0))
	metricMocks.requestChargePerQueryResponseAvg.EXPECT().Set(float64(0))
	metricMocks.requestChargePerQuery.EXPECT().Set(float64(0))
	metricMocks.requestChargeTotal.EXPECT().Add(float64(0))
	metricMocks.retryAfterMS.EXPECT().Set(float64(0))
	updateRequestMetrics(responses, metrics)

	// THEN
	// expect the calls on the metrics specified above
}

func TestUpdateMetricsFull(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, metricMocks := NewMockedMetrics(mockCtrl)

	withError := interfaces.Response{
		Status: interfaces.Status{
			Code: interfaces.StatusSuccess,
			Attributes: map[string]interface{}{
				"x-ms-status-code":          429,
				"x-ms-substatus-code":       3200,
				"x-ms-total-request-charge": 11,
				"x-ms-total-server-time-ms": 22,
				"x-ms-retry-after-ms":       "00:00:00.033",
			},
		},
	}

	responses := []interfaces.Response{withError}

	// WHEN
	mockCount200 := mock_metrics.NewMockCounter(mockCtrl)
	mockCount200.EXPECT().Inc()
	metricMocks.statusCodeTotal.EXPECT().WithLabelValues("429").Return(mockCount200)
	metricMocks.serverTimePerQueryResponseAvgMS.EXPECT().Set(float64(22))
	metricMocks.serverTimePerQueryMS.EXPECT().Set(float64(22))
	metricMocks.requestChargePerQueryResponseAvg.EXPECT().Set(float64(11))
	metricMocks.requestChargePerQuery.EXPECT().Set(float64(11))
	metricMocks.requestChargeTotal.EXPECT().Add(float64(11))
	metricMocks.retryAfterMS.EXPECT().Set(float64(33))
	updateRequestMetrics(responses, metrics)

	// THEN
	// expect the calls on the metrics specified above
}
