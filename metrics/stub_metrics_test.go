package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStubHistogram(t *testing.T) {
	// GIVEN
	// WHEN
	histogram := NewStubHistogram()

	// THEN
	assert.Equal(t, histogram,nopHistogram)
}

func TestObserve(t *testing.T) {
	// GIVEN
	histogram := NewStubHistogram()

	// WHEN
	histogram.Observe(4.2)

	// THEN nothing happens
}

func TestNewStubHistogramVec(t *testing.T) {

	// WHEN
	histogramVec := NewHistogramVec()

	// THEN
	assert.NotNil(t, histogramVec)
}

func TestWithLabelValues(t *testing.T) {
	// GIVEN
	histogramVec := NewHistogramVec()

	// WHEN
	histogram := histogramVec.WithLabelValues("test", "values")

	// THEN
	assert.Equal(t, histogram,nopHistogram)
}