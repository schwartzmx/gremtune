package metrics // nolint: golint,stylecheck // used in too many projects already

// StubCounter is a stub of Counter interface
type StubCounter struct {
}

// NewStubCounter creates a new stub instance
func NewStubCounter() *StubCounter {
	stub := &StubCounter{}
	return stub
}

// Inc stubs base method
func (m *StubCounter) Inc() {
	// do nothing, as it is a stub
}

// Add stubs base method
func (m *StubCounter) Add(arg0 float64) {
	// do nothing, as it is a stub
}

// StubGauge is a stub of Gauge interface
type StubGauge struct {
}

// NewStubGauge creates a new stub instance
func NewStubGauge() *StubGauge {
	stub := &StubGauge{}
	return stub
}

// Set stubs base method
func (m *StubGauge) Set(arg0 float64) {
	// do nothing, as it is a stub
}

// Add stubs base method
func (m *StubGauge) Add(arg0 float64) {
	// do nothing, as it is a stub
}

// StubGaugeVec is a stub of GaugeVec interface
type StubGaugeVec struct {
}

// NewStubGaugeVec creates a new stub instance
func NewStubGaugeVec() *StubGaugeVec {
	stub := &StubGaugeVec{}
	return stub
}

var nopGauge = NewStubGauge()

// WithLabelValues stubs base method
func (m *StubGaugeVec) WithLabelValues(lvs ...string) Gauge {
	return nopGauge
}

// StubCounterVec is a stub of CounterVec interface
type StubCounterVec struct {
}

// NewStubCounterVec creates a new stub instance
func NewStubCounterVec() *StubCounterVec {
	stub := &StubCounterVec{}
	return stub
}

var nopCounter = NewStubCounter()

// WithLabelValues stubs base method
func (m *StubCounterVec) WithLabelValues(lvs ...string) Counter {
	return nopCounter
}

// StubHistogram is a stub of Histogram interface
type StubHistogram struct {
}

// NewStubHistogram creates a new stub instance
func NewStubHistogram() *StubHistogram {
	stub := &StubHistogram{}
	return stub
}

// Observe stubs base method
func (m *StubHistogram) Observe(arg0 float64) {
	// just a Stub
}

// StubHistogramVec is a stub of HistogramVec interface
type StubHistogramVec struct {
}

// NewHistogramVec creates a new stub instance
func NewHistogramVec() *StubHistogramVec {
	stub := &StubHistogramVec{}
	return stub
}

var nopHistogram = NewStubHistogram()

// WithLabelValues stubs base method
func (m *StubHistogramVec) WithLabelValues(lvs ...string) Histogram {
	return nopHistogram
}
