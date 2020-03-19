package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToValues(t *testing.T) {
	t.Parallel()

	// GIVEN
	inputStr := "hello"
	inputBool := true
	inputFloat64 := 12345
	input := []interface{}{inputStr, inputBool, inputFloat64}

	// WHEN
	values, err := toValues(input)

	// THEN
	assert.NoError(t, err)
	assert.Len(t, values, 3)
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
