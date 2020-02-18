package gremcos

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock_metrics "github.com/supplyon/gremcos/test/mocks/metrics"
)

type MetricsMocks struct {
	statusCodeTotal                  *mock_metrics.MockCounterVec
	retryAfterMS                     *mock_metrics.MockGauge
	requestChargeTotal               *mock_metrics.MockCounter
	requestChargePerQuery            *mock_metrics.MockGauge
	requestChargePerQueryResponseAvg *mock_metrics.MockGauge
	serverTimePerQueryMS             *mock_metrics.MockGauge
	serverTimePerQueryResponseAvgMS  *mock_metrics.MockGauge
}

// NewMockedMetrics creates and returns mocked metrics that can be used
// for unit-testing.
// Example:
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()
// 		metrics, mocks := NewMockedMetrics(mockCtrl)
// 		mocks.scaleCounter.EXPECT().Set(10)
// use metrics...
func NewMockedMetrics(mockCtrl *gomock.Controller) (*Metrics, *MetricsMocks) {
	mStatusCodeTotal := mock_metrics.NewMockCounterVec(mockCtrl)
	mRetryAfterMS := mock_metrics.NewMockGauge(mockCtrl)
	mRequestChargeTotal := mock_metrics.NewMockCounter(mockCtrl)
	mRequestChargePerQuery := mock_metrics.NewMockGauge(mockCtrl)
	mRequestChargePerQueryResponseAvg := mock_metrics.NewMockGauge(mockCtrl)
	mServerTimePerQueryMS := mock_metrics.NewMockGauge(mockCtrl)
	mServerTimePerQueryResponseAvgMS := mock_metrics.NewMockGauge(mockCtrl)

	metrics := &Metrics{
		statusCodeTotal:                  mStatusCodeTotal,
		retryAfterMS:                     mRetryAfterMS,
		requestChargeTotal:               mRequestChargeTotal,
		requestChargePerQuery:            mRequestChargePerQuery,
		requestChargePerQueryResponseAvg: mRequestChargePerQueryResponseAvg,
		serverTimePerQueryMS:             mServerTimePerQueryMS,
		serverTimePerQueryResponseAvgMS:  mServerTimePerQueryResponseAvgMS,
	}

	mocks := &MetricsMocks{
		statusCodeTotal:                  mStatusCodeTotal,
		retryAfterMS:                     mRetryAfterMS,
		requestChargeTotal:               mRequestChargeTotal,
		requestChargePerQuery:            mRequestChargePerQuery,
		requestChargePerQueryResponseAvg: mRequestChargePerQueryResponseAvg,
		serverTimePerQueryMS:             mServerTimePerQueryMS,
		serverTimePerQueryResponseAvgMS:  mServerTimePerQueryResponseAvgMS,
	}

	return metrics, mocks
}

func Test_NewMetrics(t *testing.T) {
	metrics := NewMetrics("gremcos")
	assert.NotNil(t, metrics.statusCodeTotal)
	assert.NotNil(t, metrics.retryAfterMS)
	assert.NotNil(t, metrics.requestChargeTotal)
	assert.NotNil(t, metrics.requestChargePerQuery)
	assert.NotNil(t, metrics.requestChargePerQueryResponseAvg)
	assert.NotNil(t, metrics.serverTimePerQueryMS)
	assert.NotNil(t, metrics.serverTimePerQueryResponseAvgMS)
}
