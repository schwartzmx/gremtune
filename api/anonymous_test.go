package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnonymousInE(t *testing.T) {
	// GIVEN

	// WHEN
	e := InE()

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, "__.inE()", e.String())
}

func TestAnonymousOutE(t *testing.T) {
	// GIVEN

	// WHEN
	e := OutE()

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, "__.outE()", e.String())
}

func TestAnonymousUnfold(t *testing.T) {
	// GIVEN

	// WHEN
	e := Unfold()

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, "__.unfold()", e.String())
}

func TestAnonymousAddV(t *testing.T) {
	// GIVEN

	// WHEN
	e := AddV("some_vertex")

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, "__.addV(\"some_vertex\")", e.String())
}

func TestAnonymousConstant(t *testing.T) {
	// GIVEN

	// WHEN
	e := Constant("1234")

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, "__.constant(\"1234\")", e.String())
}
