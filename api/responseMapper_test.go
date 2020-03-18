package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUntypedToComplexType(t *testing.T) {
	t.Parallel()
	// GIVEN
	label := "thelabel"
	id := 11
	inputVertex := map[string]interface{}{
		"@type": TypeVertex,
		"@value": map[string]interface{}{
			"id": map[string]interface{}{
				"@type":  TypeVertex,
				"@value": id,
			},
			"label": label,
		},
	}
	var vertex Vertex

	inputTValue1 := "hello"
	var tValue1 TypedValue

	// WHEN
	errVertex := untypedToComplexType(inputVertex, &vertex, TypeVertex)
	errTValue1 := untypedToComplexType(inputTValue1, &tValue1, TypeString)

	// THEN
	assert.NoError(t, errVertex)
	assert.Equal(t, id, vertex.ID.Value)
	assert.Equal(t, label, vertex.Label)

	assert.NoError(t, errTValue1)
	assert.Equal(t, "hello", tValue1.AsString())
}

func TestUntypedToComplexTypeFail(t *testing.T) {
	t.Parallel()
	// GIVEN
	inputInvalid1 := map[string]interface{}{
		"someting": "wrong",
	}
	inputInvalid2 := 1234
	inputInvalid3 := map[string]interface{}{
		"@type":  TypeVertex,
		"@value": 1234,
	}
	inputWrongType := map[string]interface{}{
		"@type": TypeString,
		"@value": map[string]interface{}{
			"id": map[string]interface{}{
				"@type":  TypeVertex,
				"@value": 11,
			},
			"label": "label",
			"value": "value",
		},
	}

	inputWrongTarget := map[string]interface{}{
		"@type": TypeString,
		"@value": map[string]interface{}{
			"id": map[string]interface{}{
				"@type":  TypeVertex,
				"@value": 11,
			},
			"label": "label",
			"value": "value",
		},
	}

	var vertex Vertex

	// WHEN
	errInputInvalid1 := untypedToComplexType(inputInvalid1, &vertex, TypeVertex)
	errInputInvalid2 := untypedToComplexType(inputInvalid2, &vertex, TypeVertex)
	errInputInvalid3 := untypedToComplexType(inputInvalid3, &vertex, TypeVertex)
	errInputWrongType := untypedToComplexType(inputWrongType, &vertex, TypeVertex)
	errInputWrongTarget := untypedToComplexType(inputWrongTarget, &vertex, TypeString)

	// THEN
	assert.Error(t, errInputInvalid1)
	assert.Error(t, errInputInvalid2)
	assert.Error(t, errInputInvalid3)
	assert.Error(t, errInputWrongType)
	assert.Error(t, errInputWrongTarget)
}
