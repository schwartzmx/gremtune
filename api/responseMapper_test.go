package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/supplyon/gremcos/interfaces"
)

func createTestResponse(data string) ResponseArray {
	input := []byte(data)
	testResponse := make(ResponseArray, 0)
	response := interfaces.Response{
		Result: interfaces.Result{Data: input},
	}

	testResponse = append(testResponse, response)
	return testResponse
}

func TestResponseToValues(t *testing.T) {
	t.Parallel()
	// GIVEN
	data := `["max.mustermann@example.com",1234,true]`
	responses := createTestResponse(data)

	// WHEN
	values, err := responses.ToValues()

	// THEN
	assert.NoError(t, err)
	assert.Len(t, values, 3)
	assert.Equal(t, "max.mustermann@example.com", values[0].AsString())
	assert.Equal(t, int32(1234), values[1].AsInt32())
	assert.True(t, values[2].AsBool())
}

func TestResponseToValues_Null(t *testing.T) {
	t.Parallel()
	// GIVEN
	data := "null"
	responses := createTestResponse(data)

	// WHEN
	values, err := responses.ToValues()

	// THEN
	assert.NoError(t, err)
	assert.Empty(t, values)
}

func TestResponseToProperties(t *testing.T) {
	t.Parallel()
	// GIVEN
	data := `[{
		"id":"8fff9259-09e6-4ea5-aaf8-250b31cc7f44|pk",
		"value":"prop value",
		"label":"prop key"
	}]`
	responses := createTestResponse(data)

	// WHEN
	properties, err := responses.ToProperties()

	// THEN
	assert.NoError(t, err)
	assert.Len(t, properties, 1)
	assert.Equal(t, "8fff9259-09e6-4ea5-aaf8-250b31cc7f44|pk", properties[0].ID)
	assert.Equal(t, "prop value", properties[0].Value.AsString())
	assert.Equal(t, "prop key", properties[0].Label)
}

func TestResponseToProperties_Null(t *testing.T) {
	t.Parallel()
	// GIVEN
	data := "null"
	responses := createTestResponse(data)

	// WHEN
	values, err := responses.ToProperties()

	// THEN
	assert.NoError(t, err)
	assert.Empty(t, values)
}

func TestResponseToVertices(t *testing.T) {
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
	responses := createTestResponse(data)

	// WHEN
	vertices, err := responses.ToVertices()

	// THEN
	assert.NoError(t, err)
	assert.Len(t, vertices, 1)
	assert.Equal(t, "8fff9259-09e6-4ea5-aaf8-250b31cc7f44", vertices[0].ID)
	assert.Equal(t, "vert label", vertices[0].Label)
	assert.Len(t, vertices[0].Properties, 3)
}

func TestResponseToVertices_Null(t *testing.T) {
	t.Parallel()
	// GIVEN
	data := "null"
	responses := createTestResponse(data)

	// WHEN
	values, err := responses.ToVertices()

	// THEN
	assert.NoError(t, err)
	assert.Empty(t, values)
}

func TestResponseToEdges(t *testing.T) {
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
	responses := createTestResponse(data)

	// WHEN
	edges, err := responses.ToEdges()

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

func TestResponseToEdges_Null(t *testing.T) {
	t.Parallel()
	// GIVEN
	data := "null"
	responses := createTestResponse(data)

	// WHEN
	values, err := responses.ToEdges()

	// THEN
	assert.NoError(t, err)
	assert.Empty(t, values)
}
