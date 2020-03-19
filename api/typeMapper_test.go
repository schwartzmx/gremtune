package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToValueMap(t *testing.T) {
	t.Parallel()
	// GIVEN
	data := `[{
		"email":["max.mustermann@example.com"],
		"a number":[1234],
		"bool value":[true]
	}]`

	// WHEN
	valueMap, err := ToValueMap([]byte(data))

	// THEN
	assert.NoError(t, err)
	assert.Len(t, valueMap, 3)
	assert.True(t, valueMap["bool value"].AsBool())
}

func TestToValues(t *testing.T) {
	t.Parallel()
	// GIVEN
	data := `["max.mustermann@example.com",1234,true]`

	// WHEN
	values, err := ToValues([]byte(data))

	// THEN
	assert.NoError(t, err)
	assert.Len(t, values, 3)
	assert.Equal(t, "max.mustermann@example.com", values[0].AsString())
	assert.True(t, values[2].AsBool())
}

func TestToProperties(t *testing.T) {
	t.Parallel()
	// GIVEN
	data := `[{
		"id":"8fff9259-09e6-4ea5-aaf8-250b31cc7f44|pk",
		"value":"prop value",
		"label":"prop key"
	}]`

	// WHEN
	properties, err := ToProperties([]byte(data))

	// THEN
	assert.NoError(t, err)
	assert.Len(t, properties, 1)
	assert.Equal(t, "8fff9259-09e6-4ea5-aaf8-250b31cc7f44|pk", properties[0].ID)
	assert.Equal(t, "prop value", properties[0].Value.AsString())
	assert.Equal(t, "prop key", properties[0].Label)
}

func TestToVertices(t *testing.T) {
	t.Parallel()
	// GIVEN
	data := `[{
		"type":"vertex",
		"id":"8fff9259-09e6-4ea5-aaf8-250b31cc7f44",
		"label":"vert label",
		"properties":{
			"pk":[{
				"id":"8fff9259-09e6-4ea5-aaf8-250b31cc7f44|pk",
				"value":"test"
			}]
			,"email":[{
				"id":"80c0dfb2-b422-4005-829e-9c79acf4f642",
				"value":"max.mustermann@example.com"
			}]
			,"abcd":[{
				"id":"4f5a5962-c6a2-4eab-81cf-5b530393b54e",
				"value":true
			}]
		}}]`

	// WHEN
	vertices, err := ToVertices([]byte(data))

	// THEN
	assert.NoError(t, err)
	assert.Len(t, vertices, 1)
	assert.Equal(t, "8fff9259-09e6-4ea5-aaf8-250b31cc7f44", vertices[0].ID)
	assert.Equal(t, "vert label", vertices[0].Label)
	assert.Len(t, vertices[0].Properties, 3)
}

func TestToEdges(t *testing.T) {
	t.Parallel()
	// GIVEN
	data := `[{
		"id":"623709d5-fe22-4377-bc5b-9cb150fff124",
		"label":"edge label",
		"type":"edge",
		"inVLabel":"user1",
		"outVLabel":"user2",
		"inV":"7404ba4e-be30-486e-88e1-b2f5937a9001",
		"outV":"1111ba4e-be30-486e-88e1-b2f5937a9001"
	}]`

	// WHEN
	edges, err := ToEdges([]byte(data))

	// THEN
	assert.NoError(t, err)
	assert.Len(t, edges, 1)
	assert.Equal(t, "623709d5-fe22-4377-bc5b-9cb150fff124", edges[0].ID)
	assert.Equal(t, "edge label", edges[0].Label)
	assert.Equal(t, "user1", edges[0].InVLabel)
	assert.Equal(t, "user2", edges[0].OutVLabel)
	assert.Equal(t, "7404ba4e-be30-486e-88e1-b2f5937a9001", edges[0].InV)
	assert.Equal(t, "1111ba4e-be30-486e-88e1-b2f5937a9001", edges[0].OutV)
}

func TestToTypeArrayTypedValues(t *testing.T) {
	t.Parallel()
	// GIVEN
	var typedValues []TypedValue
	data := `["bla",true,1287]`
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
	var properties []Property
	dataProperties := `[{
		"id":"8fff9259-09e6-4ea5-aaf8-250b31cc7f44|pk",
		"value":"prop value",
		"label":"prop key"
	}]`

	var vertices []Vertex
	dataVertices := `[{
		"type":"vertex",
		"id":"8fff9259-09e6-4ea5-aaf8-250b31cc7f44",
		"label":"vertex label",
		"properties":{
			"pk":[{
				"id":"8fff9259-09e6-4ea5-aaf8-250b31cc7f44|pk",
				"value":"test"
			}]
			,"email":[{
				"id":"80c0dfb2-b422-4005-829e-9c79acf4f642",
				"value":"max.mustermann@example.com"
			}]
			,"abcd":[{
				"id":"4f5a5962-c6a2-4eab-81cf-5b530393b54e",
				"value":true
			}]
		}}]`

	var edges []Edge
	dataEdges := `[{
		"id":"623709d5-fe22-4377-bc5b-9cb150fff124",
		"label":"knows",
		"type":"edge",
		"inVLabel":"user1",
		"outVLabel":"user2",
		"inV":"7404ba4e-be30-486e-88e1-b2f5937a9001",
		"outV":"7404ba4e-be30-486e-88e1-b2f5937a9001"
	}]`

	// WHEN
	errProperties := toTypeArray([]byte(dataProperties), &properties)
	errVertices := toTypeArray([]byte(dataVertices), &vertices)
	errEdges := toTypeArray([]byte(dataEdges), &edges)

	// THEN
	assert.NoError(t, errProperties)
	assert.Len(t, properties, 1)
	assert.Equal(t, "prop value", properties[0].Value.AsString())
	assert.Equal(t, "prop key", properties[0].Label)

	assert.NoError(t, errVertices)
	assert.Len(t, vertices, 1)
	assert.Equal(t, "vertex label", vertices[0].Label)
	assert.Len(t, vertices[0].Properties, 3)

	assert.NoError(t, errEdges)
	assert.Len(t, edges, 1)
	assert.Equal(t, "knows", edges[0].Label)
	assert.Equal(t, "user1", edges[0].InVLabel)
	assert.Equal(t, "user2", edges[0].OutVLabel)
}

//func TestUntypedToComplexType(t *testing.T) {
//	t.Parallel()
//	// GIVEN
//	label := "thelabel"
//	id := 11
//	inputVertex := map[string]interface{}{
//		"@type": TypeVertex,
//		"@value": map[string]interface{}{
//			"id": map[string]interface{}{
//				"@type":  TypeVertex,
//				"@value": id,
//			},
//			"label": label,
//		},
//	}
//	var vertex Vertex
//
//	inputTValue1 := "hello"
//	var tValue1 TypedValue
//
//	// WHEN
//	errVertex := untypedToType(inputVertex, &vertex)
//	errTValue1 := untypedToType(inputTValue1, &tValue1)
//
//	// THEN
//	assert.NoError(t, errVertex)
//	assert.Equal(t, id, vertex.ID.Value)
//	assert.Equal(t, label, vertex.Label)
//
//	assert.NoError(t, errTValue1)
//	assert.Equal(t, "hello", tValue1.AsString())
//}
//
//func TestUntypedToComplexTypeFail(t *testing.T) {
//	t.Parallel()
//	// GIVEN
//	inputInvalid1 := map[string]interface{}{
//		"someting": "wrong",
//	}
//	inputInvalid2 := 1234
//	inputInvalid3 := map[string]interface{}{
//		"@type":  TypeVertex,
//		"@value": 1234,
//	}
//	inputWrongType := map[string]interface{}{
//		"@type": TypeString,
//		"@value": map[string]interface{}{
//			"id": map[string]interface{}{
//				"@type":  TypeVertex,
//				"@value": 11,
//			},
//			"label": "label",
//			"value": "value",
//		},
//	}
//
//	inputWrongTarget := map[string]interface{}{
//		"@type": TypeString,
//		"@value": map[string]interface{}{
//			"id": map[string]interface{}{
//				"@type":  TypeVertex,
//				"@value": 11,
//			},
//			"label": "label",
//			"value": "value",
//		},
//	}
//
//	var vertex Vertex
//
//	// WHEN
//	errInputInvalid1 := untypedToType(inputInvalid1, &vertex)
//	errInputInvalid2 := untypedToType(inputInvalid2, &vertex)
//	errInputInvalid3 := untypedToType(inputInvalid3, &vertex)
//	errInputWrongType := untypedToType(inputWrongType, &vertex)
//	errInputWrongTarget := untypedToType(inputWrongTarget, &vertex)
//
//	// THEN
//	assert.Error(t, errInputInvalid1)
//	assert.Error(t, errInputInvalid2)
//	assert.Error(t, errInputInvalid3)
//	assert.Error(t, errInputWrongType)
//	assert.Error(t, errInputWrongTarget)
//}
//
