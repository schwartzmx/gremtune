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
	errVertex := untypedToType(inputVertex, &vertex)
	errTValue1 := untypedToType(inputTValue1, &tValue1)

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
	errInputInvalid1 := untypedToType(inputInvalid1, &vertex)
	errInputInvalid2 := untypedToType(inputInvalid2, &vertex)
	errInputInvalid3 := untypedToType(inputInvalid3, &vertex)
	errInputWrongType := untypedToType(inputWrongType, &vertex)
	errInputWrongTarget := untypedToType(inputWrongTarget, &vertex)

	// THEN
	assert.Error(t, errInputInvalid1)
	assert.Error(t, errInputInvalid2)
	assert.Error(t, errInputInvalid3)
	assert.Error(t, errInputWrongType)
	assert.Error(t, errInputWrongTarget)
}
