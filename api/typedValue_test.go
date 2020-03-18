package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTypeMatching(t *testing.T) {
	t.Parallel()
	assert.True(t, isTypeMatching(&TypedValue{}, TypeString))
	assert.True(t, isTypeMatching(&TypedValue{}, TypeBool))
	assert.True(t, isTypeMatching(&TypedValue{}, TypeFloat64))
	assert.True(t, isTypeMatching(&TypedValue{}, TypeInt32))
	assert.True(t, isTypeMatching(&TypedValue{}, TypeInt64))
	assert.False(t, isTypeMatching(&TypedValue{}, TypeVertex))
	assert.False(t, isTypeMatching(&TypedValue{}, TypeVertexProperty))
	assert.False(t, isTypeMatching(&TypedValue{}, TypeEdge))
	assert.True(t, isTypeMatching(&Vertex{}, TypeVertex))
	assert.True(t, isTypeMatching(&VertexProperty{}, TypeVertexProperty))
	assert.True(t, isTypeMatching(&Edge{}, TypeEdge))
	assert.False(t, isTypeMatching(&Edge{}, ""))
}

func TestToValues(t *testing.T) {
	t.Parallel()

	// GIVEN
	inputStr := "hello"
	inputBool := true
	inputInt32 := map[string]interface{}{
		"@type":  TypeInt32,
		"@value": int32(11),
	}
	inputFloat64 := map[string]interface{}{
		"@type":  TypeFloat64,
		"@value": float64(22),
	}
	input := []interface{}{inputStr, inputBool, inputInt32, inputFloat64}

	// WHEN
	values, err := toValues(input)

	// THEN
	assert.NoError(t, err)
	assert.Len(t, values, 4)
}

func TestToValue(t *testing.T) {
	t.Parallel()

	// GIVEN
	inputStr := "hello"
	inputBool := true
	inputInt32 := map[string]interface{}{
		"@type":  TypeInt32,
		"@value": int32(11),
	}
	inputFloat64 := map[string]interface{}{
		"@type":  TypeFloat64,
		"@value": float64(22),
	}

	// WHEN
	valueStr, errStr := toValue(inputStr)
	valueBool, errBool := toValue(inputBool)
	valueInt32, errInt32 := toValue(inputInt32)
	valueFloat64, errFloat64 := toValue(inputFloat64)

	// THEN
	assert.NoError(t, errStr)
	assert.Equal(t, TypeString, valueStr.Type)
	assert.Equal(t, inputStr, valueStr.AsString())
	assert.NoError(t, errBool)
	assert.Equal(t, TypeBool, valueBool.Type)
	assert.Equal(t, inputBool, valueBool.AsBool())
	assert.NoError(t, errInt32)
	assert.Equal(t, TypeInt32, valueInt32.Type)
	assert.Equal(t, inputInt32["@value"], valueInt32.AsInt32())
	assert.NoError(t, errFloat64)
	assert.Equal(t, TypeFloat64, valueFloat64.Type)
	assert.Equal(t, inputFloat64["@value"], valueFloat64.AsFloat64())
}

func TestToValueFail(t *testing.T) {
	t.Parallel()

	// GIVEN
	invalidMapStructMissingTypeField := map[string]interface{}{
		"@value": float64(22),
	}
	invalidMapStructMissingValueField := map[string]interface{}{
		"@type": TypeFloat64,
	}

	invalidMapStructUnknownField := map[string]interface{}{
		"@value":       float64(22),
		"@type":        TypeFloat64,
		"unknownField": "xyz",
	}

	// WHEN
	_, errNil := toValue(nil)
	_, errWrongType := toValue(float64(23))
	_, errInvalidMapStructMissingTypeField := toValue(invalidMapStructMissingTypeField)
	_, errInvalidMapStructMissingValueField := toValue(invalidMapStructMissingValueField)
	_, errInvalidMapStructUnknownField := toValue(invalidMapStructUnknownField)

	// THEN
	assert.Error(t, errNil)
	assert.Error(t, errWrongType)
	assert.Error(t, errInvalidMapStructMissingTypeField)
	assert.Error(t, errInvalidMapStructMissingValueField)
	assert.Error(t, errInvalidMapStructUnknownField)
}
