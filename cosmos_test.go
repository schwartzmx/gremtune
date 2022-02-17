package gremcos

import (
	"fmt"
	"os"
	"testing"
	"time"

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
	require.NotNil(t, cImpl.credentialProvider)

	uname, err := cImpl.credentialProvider.Username()
	assert.NoError(t, err)
	pwd, err := cImpl.credentialProvider.Password()
	assert.NoError(t, err)
	assert.Equal(t, username, uname)
	assert.Equal(t, password, pwd)
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
}

func TestAutomaticRetries(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)

	// WHEN
	cosmos, err := New("ws://host", AutomaticRetries(3,time.Second), withMetrics(metrics))
	require.NoError(t, err)

	// THEN
	cImpl := toCosmosImpl(t, cosmos)
	assert.Equal(t, time.Second,cImpl.retryTimeout)
	assert.Equal(t, 3, cImpl.maxRetries)
}

func TestCosmosImpl_Execute_RetriesSuccess(t *testing.T) {
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
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 0
	cosmos := cosmosImpl{
		logger:           zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:             poolMock,
		metrics:          newStubbedMetrics(),
		maxRetries:       maxRetries,
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
		logger:           zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:             poolMock,
		metrics:          newStubbedMetrics(),
		maxRetries:       maxRetries,
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
		logger:           zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:             poolMock,
		metrics:          newStubbedMetrics(),
		maxRetries:       maxRetries,
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
		logger:           zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:             poolMock,
		metrics:          newStubbedMetrics(),
		maxRetries:       maxRetries,
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
		logger:           zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:             poolMock,
		metrics:          newStubbedMetrics(),
		maxRetries:       maxRetries,
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
		logger:           zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:             poolMock,
		metrics:          newStubbedMetrics(),
		maxRetries:       maxRetries,
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
		logger:           zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:             poolMock,
		metrics:          newStubbedMetrics(),
		maxRetries:       maxRetries,
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
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryExecutor, poolMock, err := newMockedPool(mockCtrl)
	require.NoError(t, err)

	const maxRetries = 3
	cosmos := cosmosImpl{
		logger:           zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:             poolMock,
		metrics:          newStubbedMetrics(),
		maxRetries:       maxRetries,
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
		logger:           zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:             poolMock,
		metrics:          newStubbedMetrics(),
		maxRetries:       maxRetries,
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
		logger:           zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:             poolMock,
		metrics:          newStubbedMetrics(),
		maxRetries:       maxRetries,
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
		logger:           zerolog.New(os.Stdout).Level(zerolog.DebugLevel),
		pool:             poolMock,
		metrics:          newStubbedMetrics(),
		maxRetries:       maxRetries,
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
