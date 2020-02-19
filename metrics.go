package gremcos

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	m "github.com/supplyon/gremcos/metrics"
)

// Metrics represents the collection of metrics internally set by the service.
type Metrics struct {
	statusCodeTotal                  m.CounterVec
	retryAfterMS                     m.Gauge
	requestChargeTotal               m.Counter
	requestChargePerQuery            m.Gauge
	requestChargePerQueryResponseAvg m.Gauge
	serverTimePerQueryMS             m.Gauge
	serverTimePerQueryResponseAvgMS  m.Gauge
}

// NewMetrics returns the metrics collection
func NewMetrics(namespace string) *Metrics {
	statusCode := []string{"code"}
	statusCodeTotal := m.NewWrappedCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "cosmos",
		Name:      "statuscode_total",
		Help:      "Counts the number of responses from cosmos separated by status code.",
	}, statusCode)

	retryAfterMS := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "cosmos",
		Name:      "retry_after_ms",
		Help:      "The time in milliseconds suggested by cosmos to wait before issuing the next query.",
	})

	requestChargePerQuery := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "cosmos",
		Name:      "request_charge_per_query",
		Help:      "Cosmos DB reports a request charge accumulated for all responses of one query. This metric represents that value.",
	})

	requestChargePerQueryResponseAvg := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "cosmos",
		Name:      "request_charge_per_queryresponse_avg",
		Help:      "Cosmos DB reports a request charge each of the responses of one query. This metric represents the average of these values for one query.",
	})

	requestChargeTotal := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "cosmos",
		Name:      "request_charge_total",
		Help:      "The accumulated request charge over all queries issued so far.",
	})

	serverTimePerQueryMS := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "cosmos",
		Name:      "server_time_per_query_ms",
		Help:      "The time spent in ms for one query.",
	})

	serverTimePerQueryResponseAvgMS := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "cosmos",
		Name:      "server_time_per_queryresponse_avg_ms",
		Help:      "The average time spent in ms for one query per response.",
	})

	return &Metrics{
		statusCodeTotal:                  statusCodeTotal,
		retryAfterMS:                     retryAfterMS,
		requestChargeTotal:               requestChargeTotal,
		requestChargePerQuery:            requestChargePerQuery,
		requestChargePerQueryResponseAvg: requestChargePerQueryResponseAvg,
		serverTimePerQueryMS:             serverTimePerQueryMS,
		serverTimePerQueryResponseAvgMS:  serverTimePerQueryResponseAvgMS,
	}
}
