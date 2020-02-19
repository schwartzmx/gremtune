package metrics

// Counter represents a counter metric
type Counter interface {
	Inc()
	Add(float64)
}

// Gauge represents a gauge metric
type Gauge interface {
	Set(float64)
	Add(float64)
}

// GaugeVec represents a vector of labelled gauges
type GaugeVec interface {
	WithLabelValues(lvs ...string) Gauge
}

// CounterVec represents a vector of labelled counters
type CounterVec interface {
	WithLabelValues(lvs ...string) Counter
}

// Histogram represents a histogram metric
type Histogram interface {
	Observe(float64)
}
