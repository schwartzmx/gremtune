package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToTypeArrayTypedValues(t *testing.T) {
	t.Parallel()
	// GIVEN
	var typedValues []TypedValue
	data := `["bla",true,{"@type":"g:Int32","@value":1287}]`
	err := toTypeArray([]byte(data), &typedValues)

	assert.NoError(t, err)
	assert.Len(t, typedValues, 3)
	assert.Equal(t, "bla", typedValues[0].AsString())
	assert.True(t, typedValues[1].AsBool())
	assert.Equal(t, int32(1287), typedValues[2].AsInt32())
}

func TestToTypeArrayProperties(t *testing.T) {
	t.Parallel()
	// GIVEN
	var properties []VertexProperty
	dataProperties := `[{
		"@type":"g:VertexProperty",
		"@value":{
		"id":{
			"@type":"g:Int64",
			"@value":30
		},
		"value":"prop value",
		"label":"prop key"
		}
	}]`

	var vertices []Vertex
	dataVertices := `[{
		"@type":"g:Vertex",
		"@value":{
		"id":{
			"@type":"g:Int64",
			"@value":30
		},
		"label":"vertex label"
		}
	}]`
	var edges []Edge
	dataEdges := `[{
		"@type":"g:Edge",
		"@value":{
			"id":{
				"@type":"g:Int64",
				"@value":38
			},
			"label":"knows",
			"inVLabel":"user1",
			"outVLabel":"user2",
			"inV":{
				"@type":"g:Int64",
				"@value":29
			},
			"outV":{
				"@type":"g:Int64",
				"@value":33
			}
		}
	}]`

	// WHEN
	errProperties := toTypeArray([]byte(dataProperties), &properties)
	errVertices := toTypeArray([]byte(dataVertices), &vertices)
	errEdges := toTypeArray([]byte(dataEdges), &edges)

	// THEN
	assert.NoError(t, errProperties)
	assert.Len(t, properties, 1)
	assert.Equal(t, "prop value", properties[0].Value)
	assert.Equal(t, "prop key", properties[0].Label)

	assert.NoError(t, errVertices)
	assert.Len(t, vertices, 1)
	assert.Equal(t, "vertex label", vertices[0].Label)

	assert.NoError(t, errEdges)
	assert.Len(t, edges, 1)
	assert.Equal(t, "knows", edges[0].Label)
	assert.Equal(t, "user1", edges[0].InVLabel)
	assert.Equal(t, "user2", edges[0].OutVLabel)
}

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