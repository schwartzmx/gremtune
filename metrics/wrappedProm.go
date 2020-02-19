package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// WrappedGaugeVec wraps a prometheus GaugeVec
type WrappedGaugeVec struct {
	prom *prometheus.GaugeVec
}

// WithLabelValues implements the WithLabelValues to meet the GaugeVec interface
func (wG *WrappedGaugeVec) WithLabelValues(lvs ...string) Gauge {
	return wG.prom.WithLabelValues(lvs...)
}

// NewWrappedGaugeVec creates a prometheus GaugeVec that is wrapped
func NewWrappedGaugeVec(opts prometheus.GaugeOpts, labelNames []string) *WrappedGaugeVec {
	return &WrappedGaugeVec{
		prom: promauto.NewGaugeVec(opts, labelNames),
	}
}

// WrappedCounterVec wraps a prometheus CounterVec
type WrappedCounterVec struct {
	prom *prometheus.CounterVec
}

// WithLabelValues implements the WithLabelValues to meet the CounterVec interface
func (wG *WrappedCounterVec) WithLabelValues(lvs ...string) Counter {
	return wG.prom.WithLabelValues(lvs...)
}

// NewWrappedCounterVec creates a prometheus CounterVec that is wrapped
func NewWrappedCounterVec(opts prometheus.CounterOpts, labelNames []string) *WrappedCounterVec {
	return &WrappedCounterVec{
		prom: promauto.NewCounterVec(opts, labelNames),
	}
}
