package gremcos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSafeCloseErrorChannel(t *testing.T) {
	// GIVEN
	// WHEN
	safeCloseErrChan := newSafeCloseErrorChannel(1)

	// THEN
	assert.NotNil(t, safeCloseErrChan)
}

func TestMultiCloseSafeCloseErrorChannel(t *testing.T) {
	// GIVEN
	safeCloseErrChan := newSafeCloseErrorChannel(1)

	// WHEN
	safeCloseErrChan.Close()
	safeCloseErrChan.Close()

	// THEN
	// there should be no panic
}

func TestNewSafeCloseIntChannel(t *testing.T) {
	// GIVEN
	// WHEN
	safeCloseIntChan := newSafeCloseIntChannel(1)

	// THEN
	assert.NotNil(t, safeCloseIntChan)
}

func TestMultiCloseSafeCloseIntChannel(t *testing.T) {
	// GIVEN
	safeCloseIntChan := newSafeCloseIntChannel(1)

	// WHEN
	safeCloseIntChan.Close()
	safeCloseIntChan.Close()

	// THEN
	// there should be no panic
}
