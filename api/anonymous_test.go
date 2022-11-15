package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnonymousWithin(t *testing.T) {
	// GIVEN

	// WHEN
	e1 := Within()
	e2 := Within("a")
	e3 := Within("a", "b", "c")

	// THEN
	assert.NotNil(t, e1)
	assert.Equal(t, `within()`, e1.String())

	assert.NotNil(t, e2)
	assert.Equal(t, `within("a")`, e2.String())

	assert.NotNil(t, e3)
	assert.Equal(t, `within("a","b","c")`, e3.String())
}

func TestAnonymousWithinInt(t *testing.T) {
	// GIVEN

	// WHEN
	e1 := WithinInt()
	e2 := WithinInt(1)
	e3 := WithinInt(1, 2, 3)

	// THEN
	assert.NotNil(t, e1)
	assert.Equal(t, `within()`, e1.String())

	assert.NotNil(t, e2)
	assert.Equal(t, `within(1)`, e2.String())

	assert.NotNil(t, e3)
	assert.Equal(t, `within(1,2,3)`, e3.String())
}

func TestAnonymousEq(t *testing.T) {
	intInput := Eq(123)
	assert.NotNil(t, intInput)
	assert.Equal(t, "eq(123)", intInput.String())

	floatInput := Eq(-0.6)
	assert.NotNil(t, floatInput)
	assert.Equal(t, "eq(-0.6)", floatInput.String())

	stringInput := Eq("abc")
	assert.NotNil(t, stringInput)
	assert.Equal(t, `eq("abc")`, stringInput.String())
}

func TestAnonymousNeq(t *testing.T) {
	intInput := Neq(123)
	assert.NotNil(t, intInput)
	assert.Equal(t, "neq(123)", intInput.String())

	floatInput := Neq(-0.6)
	assert.NotNil(t, floatInput)
	assert.Equal(t, "neq(-0.6)", floatInput.String())

	stringInput := Neq("abc")
	assert.NotNil(t, stringInput)
	assert.Equal(t, `neq("abc")`, stringInput.String())
}

func TestAnonymousLt(t *testing.T) {
	intInput := Lt(123)
	assert.NotNil(t, intInput)
	assert.Equal(t, "lt(123)", intInput.String())

	floatInput := Lt(-0.6)
	assert.NotNil(t, floatInput)
	assert.Equal(t, "lt(-0.6)", floatInput.String())

	stringInput := Lt("abc")
	assert.NotNil(t, stringInput)
	assert.Equal(t, `lt("abc")`, stringInput.String())
}

func TestAnonymousLte(t *testing.T) {
	intInput := Lte(123)
	assert.NotNil(t, intInput)
	assert.Equal(t, "lte(123)", intInput.String())

	floatInput := Lte(-0.6)
	assert.NotNil(t, floatInput)
	assert.Equal(t, "lte(-0.6)", floatInput.String())

	stringInput := Lte("abc")
	assert.NotNil(t, stringInput)
	assert.Equal(t, `lte("abc")`, stringInput.String())
}

func TestAnonymousGt(t *testing.T) {
	intInput := Gt(123)
	assert.NotNil(t, intInput)
	assert.Equal(t, "gt(123)", intInput.String())

	floatInput := Gt(-0.6)
	assert.NotNil(t, floatInput)
	assert.Equal(t, "gt(-0.6)", floatInput.String())

	stringInput := Gt("abc")
	assert.NotNil(t, stringInput)
	assert.Equal(t, `gt("abc")`, stringInput.String())
}

func TestAnonymousGte(t *testing.T) {
	intInput := Gte(123)
	assert.NotNil(t, intInput)
	assert.Equal(t, "gte(123)", intInput.String())

	floatInput := Gte(-0.6)
	assert.NotNil(t, floatInput)
	assert.Equal(t, "gte(-0.6)", floatInput.String())

	stringInput := Gte("abc")
	assert.NotNil(t, stringInput)
	assert.Equal(t, `gte("abc")`, stringInput.String())
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
