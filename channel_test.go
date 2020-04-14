package gremcos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCloseOnceChannel(t *testing.T) {
	// GIVEN
	// WHEN
	safeCloseErrChan := newSafeCloseErrorChannel(1)

	// THEN
	assert.NotNil(t, safeCloseErrChan)
}

func TestMultiCloseErrChannel(t *testing.T) {
	// GIVEN
	safeCloseErrChan := newSafeCloseErrorChannel(1)

	// WHEN
	safeCloseErrChan.Close()
	safeCloseErrChan.Close()

	// THEN
	// there should be no panic
}
