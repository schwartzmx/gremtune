package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnonymousEq(t *testing.T) {
	// GIVEN

	// WHEN
	e := Eq(123)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, "eq(123)", e.String())
}

func TestAnonymousNeq(t *testing.T) {
	// GIVEN

	// WHEN
	e := Neq(123)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, "neq(123)", e.String())
}

func TestAnonymousLt(t *testing.T) {
	// GIVEN

	// WHEN
	e := Lt(123)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, "lt(123)", e.String())
}

func TestAnonymousLte(t *testing.T) {
	// GIVEN

	// WHEN
	e := Lte(123)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, "lte(123)", e.String())
}

func TestAnonymousGt(t *testing.T) {
	// GIVEN

	// WHEN
	e := Gt(123)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, "gt(123)", e.String())
}

func TestAnonymousGte(t *testing.T) {
	// GIVEN

	// WHEN
	e := Gte(123)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, "gte(123)", e.String())
}

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

func TestAnonymousInV(t *testing.T) {
	// GIVEN

	// WHEN
	v := InV()

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, "__.inV()", v.String())
}

func TestAnonymousOutV(t *testing.T) {
	// GIVEN

	// WHEN
	v := OutV()

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, "__.outV()", v.String())
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

func TestAnonymousHas(t *testing.T) {
	// GIVEN

	// WHEN
	e := Has("name", "hans")

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, `__.has("name","hans")`, e.String())
}

func TestAnonymousHas_OnlyKey(t *testing.T) {
	// GIVEN

	// WHEN
	e := Has("name")

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, `__.has("name")`, e.String())
}
