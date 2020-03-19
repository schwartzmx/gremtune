package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_toValues(t *testing.T) {
	t.Parallel()

	// GIVEN
	inputStr := "hello"
	inputBool := true
	inputFloat64 := 123.45
	input := []interface{}{inputStr, inputBool, inputFloat64}

	// WHEN
	values, err := toValues(input)

	// THEN
	assert.NoError(t, err)
	assert.Len(t, values, 3)
	assert.Equal(t, inputStr, values[0].AsString())
	assert.True(t, values[1].AsBool())
	assert.Equal(t, inputFloat64, values[2].AsFloat64())
}

func TestToValue(t *testing.T) {
	t.Parallel()

	// GIVEN
	inputStr := "hello"
	inputBool := true
	inputFloat64 := float64(12345)

	// WHEN
	valueStr, errStr := toValue(inputStr)
	valueBool, errBool := toValue(inputBool)
	valueFloat64, errFloat64 := toValue(inputFloat64)

	// THEN
	assert.NoError(t, errStr)
	assert.Equal(t, inputStr, valueStr.AsString())
	assert.NoError(t, errBool)
	assert.Equal(t, inputBool, valueBool.AsBool())
	assert.NoError(t, errFloat64)
	assert.Equal(t, inputFloat64, valueFloat64.AsFloat64())
}

func TestVertexPropertyValue(t *testing.T) {
	t.Parallel()

	// GIVEN
	key := "myprop"
	value := TypedValue{Value: "some value"}
	valueWithIDInput := ValueWithID{
		ID:    "123",
		Value: value,
	}

	props := VertexPropertyMap{key: valueWithIDInput}

	// WHEN
	valueWithIDExtracted, exists := props.Value(key)

	// THEN
	assert.True(t, exists)
	assert.Equal(t, value, valueWithIDExtracted.Value)
}

func TestVertexPropertyValueMissing(t *testing.T) {
	t.Parallel()

	// GIVEN
	key := "myprop"
	props := VertexPropertyMap{}

	// WHEN
	_, exists := props.Value(key)

	// THEN
	assert.False(t, exists)
}

func TestVertexPropertyValueEmpty(t *testing.T) {
	t.Parallel()

	// GIVEN
	key := "myprop"
	valueWithIDInput := ValueWithID{}
	props := VertexPropertyMap{key: valueWithIDInput}

	// WHEN
	_, exists := props.Value(key)

	// THEN
	assert.False(t, exists)
}

func TestVertexPropertyAsString(t *testing.T) {
	t.Parallel()

	// GIVEN
	key := "myprop"
	value := "some value"
	valueWithIDInput := ValueWithID{
		ID:    "123",
		Value: TypedValue{Value: value},
	}

	props := VertexPropertyMap{key: valueWithIDInput}

	// WHEN
	valueWithIDExtracted, err := props.AsString(key)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, value, valueWithIDExtracted)
}

func TestVertexPropertyAsInt32(t *testing.T) {
	t.Parallel()

	// GIVEN
	key := "myprop"
	value := int32(12345)
	valueWithIDInput := ValueWithID{
		ID:    "123",
		Value: TypedValue{Value: value},
	}

	props := VertexPropertyMap{key: valueWithIDInput}

	// WHEN
	valueWithIDExtracted, err := props.AsInt32(key)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, value, valueWithIDExtracted)
}
