package gremcos

import (
	"github.com/prometheus/client_golang/prometheus"
	m "github.com/supplyon/gremcos/metrics"
)

// Metrics represents the collection of metrics internally set by the service.
type Metrics struct {
	statusCodeTotal m.CounterVec
}

// NewMetrics returns the metrics collection
func NewMetrics(namespace string) Metrics {

	statusCode := []string{"code"}
	statusCodeTotal := m.NewWrappedCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "cosmos",
		Name:      "statuscode_total",
		Help:      "Counts the number of responses from cosmos separated by status code.",
	}, statusCode)

	return Metrics{
		statusCodeTotal: statusCodeTotal,
	}
}
