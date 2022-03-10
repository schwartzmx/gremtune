package gremcos

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
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
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, metricMocks := NewMockedMetrics(mockCtrl)
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
	mockCount := mock_metrics.NewMockCounter(mockCtrl)
	mockCount.EXPECT().Inc().Times(2)
	metricMocks.connectionUsageTotal.EXPECT().WithLabelValues("kind", "READ", "error", "true").Return(mockCount).Times(2)

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
	// Closing the QueryExecutors here because they are not pooled and would be ignored by cosmos.Stop
	assert.NoError(t, client1.Close())
	assert.NoError(t, client2.Close())
	assert.NoError(t, cosmos.Stop())
}

func TestNew(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

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
	require.NotNil(t, cImpl.credentialProvider)

	uname, err := cImpl.credentialProvider.Username()
	assert.NoError(t, err)
	pwd, err := cImpl.credentialProvider.Password()
	assert.NoError(t, err)
	assert.Equal(t, username, uname)
	assert.Equal(t, password, pwd)
	assert.NoError(t, cosmos.Stop())
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

	// WHEN closed
	mockedQueryExecutor.EXPECT().Close().Times(1).Return(nil)

	// THEN
	assert.NoError(t, healthyWhenConnected)
	assert.Error(t, healthyWhenNotConnected)
	assert.NoError(t, cosmos.Stop())
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
	assert.NoError(t, cosmos.Stop())
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
	// since there were no responses
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
	metricMocks.retryAfterMS.EXPECT().Observe(float64(0))
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
	metricMocks.retryAfterMS.EXPECT().Observe(float64(33))
	updateRequestMetrics(responses, metrics)

	// THEN
	// expect the calls on the metrics specified above
}

func TestWithResourceTokenAuth(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)
	username := "abcd"
	password := "xyz"

	// WHEN
	cosmos, err := New("ws://host",
		WithResourceTokenAuth(
			StaticCredentialProvider{
				UsernameStatic: username,
				PasswordStatic: password,
			}),
		withMetrics(metrics),
	)

	// THEN
	require.NoError(t, err)
	cImpl := toCosmosImpl(t, cosmos)
	require.NotNil(t, cImpl.credentialProvider)

	uname, err := cImpl.credentialProvider.Username()
	assert.NoError(t, err)
	pwd, err := cImpl.credentialProvider.Password()
	assert.NoError(t, err)
	assert.Equal(t, username, uname)
	assert.Equal(t, password, pwd)
	assert.NoError(t, cosmos.Stop())
}

func TestString(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)
	username := "abcd"
	password := "xyz"

	// WHEN
	cosmos, err := New("ws://host",
		WithResourceTokenAuth(
			StaticCredentialProvider{
				UsernameStatic: username,
				PasswordStatic: password,
			}),
		withMetrics(metrics),
	)

	// THEN
	require.NoError(t, err)
	cImpl := toCosmosImpl(t, cosmos)
	require.NotNil(t, cImpl.credentialProvider)

	assert.Equal(t, "CosmosDB (connected=false, target=ws://host, user=abcd)", cImpl.String())
	assert.NoError(t, cosmos.Stop())
}

func TestWithLogger(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)
	log := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)

	// WHEN
	cosmos, err := New("ws://host", WithLogger(log), withMetrics(metrics))
	require.NoError(t, err)

	// THEN
	cImpl := toCosmosImpl(t, cosmos)
	require.NotNil(t, cImpl.credentialProvider)
	assert.NotEqual(t, zerolog.Nop(), cImpl.logger)
	assert.Equal(t, zerolog.DebugLevel, cImpl.logger.GetLevel())
	assert.NoError(t, cosmos.Stop())
}

func TestAutomaticRetries(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)

	// WHEN
	cosmos, err := New("ws://host", AutomaticRetries(3, time.Second), withMetrics(metrics))
	require.NoError(t, err)

	// THEN
	cImpl := toCosmosImpl(t, cosmos)
	assert.Equal(t, time.Second, cImpl.retryTimeout)
	assert.Equal(t, 3, cImpl.maxRetries)
	assert.NoError(t, cosmos.Stop())
}

func TestAutomaticRetries_DefaultTimeout(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)

	// WHEN
	cosmos, err := New("ws://host", AutomaticRetries(3, 0*time.Second), withMetrics(metrics))
	require.NoError(t, err)

	// THEN
	cImpl := toCosmosImpl(t, cosmos)
	assert.Equal(t, time.Second*30, cImpl.retryTimeout)
	assert.Equal(t, 3, cImpl.maxRetries)
	assert.NoError(t, cosmos.Stop())
}

func TestCosmosImpl_ExecuteQuery_NoQuery(t *testing.T) {
	// GIVEN
	cosmos := cosmosImpl{}

	// WHEN
	responses, err := cosmos.ExecuteQuery(nil)

	// THEN
	assert.EqualError(t, err, "query is nil")
	assert.Nil(t, responses)
}

func TestCosmosImpl_Execute_RetriesSuccess(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 3
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:         poolMock,
		metrics:      newStubbedMetrics(),
		maxRetries:   maxRetries,
		retryTimeout: time.Second * 2,
	}

	query := "g.V().has(\"user_id\",\"12345\")"

	success := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusSuccess,
			},
		},
	}
	doRetry := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusServerError,
				Attributes: map[string]interface{}{
					"x-ms-status-code":    429,
					"x-ms-substatus-code": 3200,
					"x-ms-retry-after-ms": "00:00:00.0500000",
				},
			},
		},
	}

	queryExecutor.EXPECT().LastError().AnyTimes().Return(nil)
	queryExecutor.EXPECT().IsConnected().AnyTimes().Return(true)

	gomock.InOrder(
		queryExecutor.EXPECT().Execute(query).Times(maxRetries).
			Return(doRetry, nil),
		queryExecutor.EXPECT().Execute(query).Times(1).
			Return(success, nil),
	)

	// WHEN
	responses, err := cosmos.Execute(query)

	// THEN
	assert.EqualValues(t, success, responses)
	assert.NoError(t, err)
}

func TestCosmosImpl_Execute_NoRetries(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 0
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:         poolMock,
		metrics:      newStubbedMetrics(),
		maxRetries:   maxRetries,
		retryTimeout: time.Second * 2,
	}

	query := "g.V().has(\"user_id\",\"12345\")"

	doRetry := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusServerError,
				Attributes: map[string]interface{}{
					"x-ms-status-code":    429,
					"x-ms-substatus-code": 3200,
					"x-ms-retry-after-ms": "00:00:00.5000000",
				},
			},
		},
	}

	queryExecutor.EXPECT().LastError().AnyTimes().Return(nil)
	queryExecutor.EXPECT().IsConnected().AnyTimes().Return(true)

	gomock.InOrder(
		queryExecutor.EXPECT().Execute(query).Times(1).
			Return(doRetry, nil),
	)

	// WHEN
	responses, err := cosmos.Execute(query)

	// THEN
	assert.EqualValues(t, responses, doRetry)
	assert.EqualError(t, err, "429 (3200) - Request was throttled and should be retried after value in x-ms-retry-after-ms")
}

func TestCosmosImpl_Execute_MaxRetriesFailure(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 3
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:         poolMock,
		metrics:      newStubbedMetrics(),
		maxRetries:   maxRetries,
		retryTimeout: time.Second * 2,
	}

	query := "g.V().has(\"user_id\",\"12345\")"

	doRetry := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusServerError,
				Attributes: map[string]interface{}{
					"x-ms-status-code":    429,
					"x-ms-substatus-code": 3200,
					"x-ms-retry-after-ms": "00:00:00.0500000",
				},
			},
		},
	}

	queryExecutor.EXPECT().LastError().AnyTimes().Return(nil)
	queryExecutor.EXPECT().IsConnected().AnyTimes().Return(true)

	gomock.InOrder(
		queryExecutor.EXPECT().Execute(query).Times(maxRetries).
			Return(doRetry, nil),
		queryExecutor.EXPECT().Execute(query).Times(1).
			Return(doRetry, nil),
	)

	// WHEN
	responses, err := cosmos.Execute(query)

	// THEN
	assert.EqualValues(t, doRetry, responses)
	assert.EqualError(t, err, "429 (3200) - Request was throttled and should be retried after value in x-ms-retry-after-ms")
}

func TestCosmosImpl_Execute_NoRetriesAfterSuccess(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 3
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:         poolMock,
		metrics:      newStubbedMetrics(),
		maxRetries:   maxRetries,
		retryTimeout: time.Second * 2,
	}

	query := "g.V().has(\"user_id\",\"12345\")"

	success := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusSuccess,
			},
		},
	}
	doRetry := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusServerError,
				Attributes: map[string]interface{}{
					"x-ms-status-code":    429,
					"x-ms-substatus-code": 3200,
					"x-ms-retry-after-ms": "00:00:00.0500000",
				},
			},
		},
	}

	queryExecutor.EXPECT().LastError().AnyTimes().Return(nil)
	queryExecutor.EXPECT().IsConnected().AnyTimes().Return(true)

	gomock.InOrder(
		queryExecutor.EXPECT().Execute(query).Times(1).
			Return(doRetry, nil),
		queryExecutor.EXPECT().Execute(query).Times(1).
			Return(success, nil),
	)

	// WHEN
	responses, err := cosmos.Execute(query)

	// THEN
	assert.EqualValues(t, responses, success)
	assert.NoError(t, err)
}

func TestCosmosImpl_ExecuteWithBindings_RetriesSuccess(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 3
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:         poolMock,
		metrics:      newStubbedMetrics(),
		maxRetries:   maxRetries,
		retryTimeout: time.Second * 2,
	}

	query := "g.V().has(\"user_id\",\"12345\")"

	success := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusSuccess,
			},
		},
	}
	doRetry := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusServerError,
				Attributes: map[string]interface{}{
					"x-ms-status-code":    429,
					"x-ms-substatus-code": 3200,
					"x-ms-retry-after-ms": "00:00:00.0500000",
				},
			},
		},
	}

	queryExecutor.EXPECT().LastError().AnyTimes().Return(nil)
	queryExecutor.EXPECT().IsConnected().AnyTimes().Return(true)

	gomock.InOrder(
		queryExecutor.EXPECT().ExecuteWithBindings(gomock.Any(), gomock.Any(), gomock.Any()).Times(maxRetries).
			Return(doRetry, nil),
		queryExecutor.EXPECT().ExecuteWithBindings(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
			Return(success, nil),
	)

	// WHEN
	responses, err := cosmos.ExecuteWithBindings(query, nil, nil)

	// THEN
	assert.EqualValues(t, success, responses)
	assert.NoError(t, err)
}

func TestCosmosImpl_ExecuteWithBindings_NoRetries(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 0
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:         poolMock,
		metrics:      newStubbedMetrics(),
		maxRetries:   maxRetries,
		retryTimeout: time.Second * 2,
	}

	query := "g.V().has(\"user_id\",\"12345\")"

	doRetry := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusServerError,
				Attributes: map[string]interface{}{
					"x-ms-status-code":    429,
					"x-ms-substatus-code": 3200,
					"x-ms-retry-after-ms": "00:00:00.5000000",
				},
			},
		},
	}

	queryExecutor.EXPECT().LastError().AnyTimes().Return(nil)
	queryExecutor.EXPECT().IsConnected().AnyTimes().Return(true)

	gomock.InOrder(
		queryExecutor.EXPECT().ExecuteWithBindings(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
			Return(doRetry, nil),
	)

	// WHEN
	responses, err := cosmos.ExecuteWithBindings(query, nil, nil)

	// THEN
	assert.EqualValues(t, responses, doRetry)
	assert.EqualError(t, err, "429 (3200) - Request was throttled and should be retried after value in x-ms-retry-after-ms")
}

func TestCosmosImpl_ExecuteWithBindings_MaxRetriesFailure(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 3
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:         poolMock,
		metrics:      newStubbedMetrics(),
		maxRetries:   maxRetries,
		retryTimeout: time.Second * 2,
	}

	query := "g.V().has(\"user_id\",\"12345\")"

	doRetry := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusServerError,
				Attributes: map[string]interface{}{
					"x-ms-status-code":    429,
					"x-ms-substatus-code": 3200,
					"x-ms-retry-after-ms": "00:00:00.0500000",
				},
			},
		},
	}

	queryExecutor.EXPECT().LastError().AnyTimes().Return(nil)
	queryExecutor.EXPECT().IsConnected().AnyTimes().Return(true)

	gomock.InOrder(
		queryExecutor.EXPECT().ExecuteWithBindings(gomock.Any(), gomock.Any(), gomock.Any()).Times(maxRetries).
			Return(doRetry, nil),
		queryExecutor.EXPECT().ExecuteWithBindings(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
			Return(doRetry, nil),
	)

	// WHEN
	responses, err := cosmos.ExecuteWithBindings(query, nil, nil)

	// THEN
	assert.EqualValues(t, doRetry, responses)
	assert.EqualError(t, err, "429 (3200) - Request was throttled and should be retried after value in x-ms-retry-after-ms")
}

func TestCosmosImpl_ExecuteWithBindings_NoRetriesAfterSuccess(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 3
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:         poolMock,
		metrics:      newStubbedMetrics(),
		maxRetries:   maxRetries,
		retryTimeout: time.Second * 2,
	}

	query := "g.V().has(\"user_id\",\"12345\")"

	success := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusSuccess,
			},
		},
	}
	doRetry := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusServerError,
				Attributes: map[string]interface{}{
					"x-ms-status-code":    429,
					"x-ms-substatus-code": 3200,
					"x-ms-retry-after-ms": "00:00:00.0500000",
				},
			},
		},
	}

	queryExecutor.EXPECT().LastError().AnyTimes().Return(nil)
	queryExecutor.EXPECT().IsConnected().AnyTimes().Return(true)

	gomock.InOrder(
		queryExecutor.EXPECT().ExecuteWithBindings(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
			Return(doRetry, nil),
		queryExecutor.EXPECT().ExecuteWithBindings(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
			Return(success, nil),
	)

	// WHEN
	responses, err := cosmos.ExecuteWithBindings(query, nil, nil)

	// THEN
	assert.EqualValues(t, responses, success)
	assert.NoError(t, err)
}

func TestCosmosImpl_ExecuteAsync_RetriesSuccess(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 3
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:         poolMock,
		metrics:      newStubbedMetrics(),
		maxRetries:   maxRetries,
		retryTimeout: time.Second * 2,
	}

	query := "g.V().has(\"user_id\",\"12345\")"

	success := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusSuccess,
			},
		},
	}
	doRetry := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusServerError,
				Attributes: map[string]interface{}{
					"x-ms-status-code":    429,
					"x-ms-substatus-code": 3200,
					"x-ms-retry-after-ms": "00:00:00.0500000",
				},
			},
		},
	}

	queryExecutor.EXPECT().LastError().AnyTimes().Return(nil)
	queryExecutor.EXPECT().IsConnected().AnyTimes().Return(true)

	gomock.InOrder(
		queryExecutor.EXPECT().ExecuteAsync(gomock.Any(), gomock.Any()).Times(maxRetries).DoAndReturn(func(q string, resp chan interfaces.AsyncResponse) error {
			resp <- interfaces.AsyncResponse{Response: doRetry[0]}
			close(resp)
			return nil
		}),
		queryExecutor.EXPECT().ExecuteAsync(gomock.Any(), gomock.Any()).Times(1).
			DoAndReturn(func(q string, resp chan interfaces.AsyncResponse) error {
				resp <- interfaces.AsyncResponse{Response: success[0]}
				close(resp)
				return nil
			}),
	)

	// WHEN
	responseChannel := make(chan interfaces.AsyncResponse, 100)

	err = cosmos.ExecuteAsync(query, responseChannel)

	responses := make([]interfaces.Response, 0, 1)
	for resp := range responseChannel {
		responses = append(responses, resp.Response)
	}

	// THEN
	assert.EqualValues(t, success, responses)
	assert.NoError(t, err)
}

func TestCosmosImpl_ExecuteAsync_NoRetries(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 0
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:         poolMock,
		metrics:      newStubbedMetrics(),
		maxRetries:   maxRetries,
		retryTimeout: time.Second * 2,
	}

	query := "g.V().has(\"user_id\",\"12345\")"

	doRetry := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusServerError,
				Attributes: map[string]interface{}{
					"x-ms-status-code":    429,
					"x-ms-substatus-code": 3200,
					"x-ms-retry-after-ms": "00:00:00.5000000",
				},
			},
		},
	}

	queryExecutor.EXPECT().LastError().AnyTimes().Return(nil)
	queryExecutor.EXPECT().IsConnected().AnyTimes().Return(true)

	gomock.InOrder(
		queryExecutor.EXPECT().ExecuteAsync(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(func(q string, resp chan interfaces.AsyncResponse) error {
			resp <- interfaces.AsyncResponse{Response: doRetry[0]}
			close(resp)
			return nil
		}),
	)

	// WHEN
	responseChannel := make(chan interfaces.AsyncResponse, 100)

	err = cosmos.ExecuteAsync(query, responseChannel)

	responses := make([]interfaces.Response, 0, 1)
	for resp := range responseChannel {
		responses = append(responses, resp.Response)
	}

	// THEN
	assert.EqualValues(t, responses, doRetry)
	assert.NoError(t, err)
}

func TestCosmosImpl_ExecuteAsync_MaxRetriesFailure(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 3
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:         poolMock,
		metrics:      newStubbedMetrics(),
		maxRetries:   maxRetries,
		retryTimeout: time.Second * 2,
	}

	query := "g.V().has(\"user_id\",\"12345\")"

	doRetry := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusServerError,
				Attributes: map[string]interface{}{
					"x-ms-status-code":    429,
					"x-ms-substatus-code": 3200,
					"x-ms-retry-after-ms": "00:00:00.0500000",
				},
			},
		},
	}

	queryExecutor.EXPECT().LastError().AnyTimes().Return(nil)
	queryExecutor.EXPECT().IsConnected().AnyTimes().Return(true)

	gomock.InOrder(
		queryExecutor.EXPECT().ExecuteAsync(gomock.Any(), gomock.Any()).Times(maxRetries).DoAndReturn(func(q string, resp chan interfaces.AsyncResponse) error {
			resp <- interfaces.AsyncResponse{Response: doRetry[0]}
			close(resp)
			return nil
		}),
		queryExecutor.EXPECT().ExecuteAsync(gomock.Any(), gomock.Any()).Times(1).
			DoAndReturn(func(q string, resp chan interfaces.AsyncResponse) error {
				resp <- interfaces.AsyncResponse{Response: doRetry[0]}
				close(resp)
				return nil
			}),
	)

	// WHEN
	responseChannel := make(chan interfaces.AsyncResponse, 100)

	err = cosmos.ExecuteAsync(query, responseChannel)

	responses := make([]interfaces.Response, 0, 1)
	for resp := range responseChannel {
		responses = append(responses, resp.Response)
	}

	// THEN
	assert.EqualValues(t, doRetry, responses)
	assert.NoError(t, err)
}

func TestCosmosImpl_ExecuteAsync_NoRetriesAfterSuccess(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 3
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:         poolMock,
		metrics:      newStubbedMetrics(),
		maxRetries:   maxRetries,
		retryTimeout: time.Second * 2,
	}

	query := "g.V().has(\"user_id\",\"12345\")"

	success := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusSuccess,
			},
		},
	}
	doRetry := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusServerError,
				Attributes: map[string]interface{}{
					"x-ms-status-code":    429,
					"x-ms-substatus-code": 3200,
					"x-ms-retry-after-ms": "00:00:00.0500000",
				},
			},
		},
	}

	queryExecutor.EXPECT().LastError().AnyTimes().Return(nil)
	queryExecutor.EXPECT().IsConnected().AnyTimes().Return(true)

	gomock.InOrder(
		queryExecutor.EXPECT().ExecuteAsync(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(func(q string, resp chan interfaces.AsyncResponse) error {
			resp <- interfaces.AsyncResponse{Response: doRetry[0]}
			close(resp)
			return nil
		}),
		queryExecutor.EXPECT().ExecuteAsync(gomock.Any(), gomock.Any()).Times(1).
			DoAndReturn(func(q string, resp chan interfaces.AsyncResponse) error {
				resp <- interfaces.AsyncResponse{Response: success[0]}
				close(resp)
				return nil
			}),
	)

	// WHEN
	responseChannel := make(chan interfaces.AsyncResponse, 100)

	err = cosmos.ExecuteAsync(query, responseChannel)

	responses := make([]interfaces.Response, 0, 1)
	for resp := range responseChannel {
		responses = append(responses, resp.Response)
	}

	// THEN
	assert.EqualValues(t, responses, success)
	assert.NoError(t, err)
}

func TestCosmosImpl_ExecuteAsync_NoRetriesAfterTimeout(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 3
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:         poolMock,
		metrics:      newStubbedMetrics(),
		maxRetries:   maxRetries,
		retryTimeout: time.Millisecond * 90,
	}

	query := "g.V().has(\"user_id\",\"12345\")"

	doRetry := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusServerError,
				Attributes: map[string]interface{}{
					"x-ms-status-code":    429,
					"x-ms-substatus-code": 3200,
					"x-ms-retry-after-ms": "00:00:00.0500000",
				},
			},
		},
	}

	queryExecutor.EXPECT().LastError().AnyTimes().Return(nil)
	queryExecutor.EXPECT().IsConnected().AnyTimes().Return(true)

	// Called only two times, as it is aborted before
	queryExecutor.EXPECT().ExecuteAsync(gomock.Any(), gomock.Any()).Times(2).DoAndReturn(func(q string, resp chan interfaces.AsyncResponse) error {
		resp <- interfaces.AsyncResponse{Response: doRetry[0]}
		close(resp)
		return nil
	})

	// WHEN
	responseChannel := make(chan interfaces.AsyncResponse, 100)

	err = cosmos.ExecuteAsync(query, responseChannel)

	responses := make([]interfaces.Response, 0, 1)
	for resp := range responseChannel {
		responses = append(responses, resp.Response)
	}

	// THEN
	assert.EqualValues(t, responses, doRetry)
	assert.NoError(t, err)
}

func TestCosmosImpl_ExecuteAsync_AbortRetryIfRetryAfterTooLong(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 3
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:         poolMock,
		metrics:      newStubbedMetrics(),
		maxRetries:   maxRetries,
		retryTimeout: time.Millisecond * 110,
	}

	query := "g.V().has(\"user_id\",\"12345\")"

	doRetry := []interfaces.Response{
		{
			Status: interfaces.Status{
				Code: interfaces.StatusServerError,
				Attributes: map[string]interface{}{
					"x-ms-status-code":    429,
					"x-ms-substatus-code": 3200,
					"x-ms-retry-after-ms": "00:00:00.2000000",
				},
			},
		},
	}

	queryExecutor.EXPECT().LastError().AnyTimes().Return(nil)
	queryExecutor.EXPECT().IsConnected().AnyTimes().Return(true)

	// Called only one times, as it is aborted before due to too long retry-after
	queryExecutor.EXPECT().ExecuteAsync(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(func(q string, resp chan interfaces.AsyncResponse) error {
		resp <- interfaces.AsyncResponse{Response: doRetry[0]}
		close(resp)
		return nil
	})

	// WHEN
	responseChannel := make(chan interfaces.AsyncResponse, 100)

	err = cosmos.ExecuteAsync(query, responseChannel)

	responses := make([]interfaces.Response, 0, 1)
	for resp := range responseChannel {
		responses = append(responses, resp.Response)
	}

	// THEN
	assert.EqualValues(t, responses, doRetry)
	assert.NoError(t, err)
}

func TestWaitForRetry(t *testing.T) {
	// GIVEN
	waitTime := time.Millisecond * 20
	stop := make(chan bool)
	defer close(stop)
	now := time.Now()
	// WHEN
	waitDone := waitForRetry(waitTime, stop)
	duration := time.Since(now)

	// THEN
	assert.True(t, waitDone)
	assert.True(t, duration >= waitTime)
}

func TestWaitForRetry_Abort(t *testing.T) {

	// GIVEN
	waitTime := time.Millisecond * 20
	stop := make(chan bool)
	defer close(stop)
	now := time.Now()
	waitDone := false
	called := false
	mu := sync.Mutex{}
	// WHEN
	go func() {
		mu.Lock()
		defer mu.Unlock()
		waitDone = waitForRetry(waitTime, stop)
		called = true
	}()
	stop <- true

	duration := time.Since(now)

	// THEN
	mu.Lock()
	assert.False(t, waitDone)
	assert.True(t, called)
	mu.Unlock()
	assert.True(t, duration <= waitTime)
}

func TestHandleTimeout(t *testing.T) {
	// GIVEN
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		retryTimeout: time.Millisecond * 50,
	}
	done := make(chan bool)
	defer close(done)

	timedOut := false
	mu := sync.Mutex{}

	// WHEN
	go func() {
		timedOutChan := cosmos.handleTimeout(done)

		timeOut := <-timedOutChan
		mu.Lock()
		timedOut = timeOut
		mu.Unlock()
	}()
	time.Sleep(time.Millisecond * 20)

	// THEN
	mu.Lock()
	assert.False(t, timedOut)
	mu.Unlock()

	time.Sleep(time.Millisecond * 31)
	mu.Lock()
	assert.True(t, timedOut)
	mu.Unlock()
}

func TestHandleTimeout_Abort(t *testing.T) {
	// GIVEN
	cosmos := cosmosImpl{
		logger:       zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		retryTimeout: time.Millisecond * 50,
	}
	done := make(chan bool)

	timedOut := false
	closed := false
	mu := sync.Mutex{}

	// WHEN
	go func() {
		timedOutChan := cosmos.handleTimeout(done)

		for isTimedOut := range timedOutChan {
			mu.Lock()
			timedOut = isTimedOut
			mu.Unlock()
		}

		mu.Lock()
		closed = true
		mu.Unlock()
	}()
	time.Sleep(time.Millisecond * 20)
	close(done)

	// THEN
	time.Sleep(time.Millisecond * 1)

	mu.Lock()
	assert.False(t, timedOut)
	assert.True(t, closed)
	mu.Unlock()
}
