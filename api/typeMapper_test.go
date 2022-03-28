package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dataVertices = `[{
	"id": "example",
	"label": "device",
	"type": "vertex",
	"properties": {
	  "pk": [
		{
		  "id": "example",
		  "value": "example"
		}
	  ],
	  "deviceIPAddress": [
		{
		  "id": "example",
		  "value": ".36"
		}
	  ],
	  "critical": [
		{
		  "id": "example",
		  "value": "No"
		}
	  ],
	  "placeInNetwork": [
		{
		  "id": "example-f57f0a21dfde",
		  "value": "example_bldg-example"
		}
	  ],
	  "vendor": [
		{
		  "id": "example-example-example-example-eb02386bacd3",
		  "value": "example"
		}
	  ],
	  "deviceModel": [
		{
		  "id": "example-example-example-b150-976553721ca4",
		  "value": "example 10506e"
		}
	  ],
	  "deviceTypeName": [
		{
		  "id": "example-example-example-9c7e-50bc3ab37313",
		  "value": "Switch"
		}
	  ],
	  "region": [
		{
		  "id": "example-example-4698-example-1ee85c528c66",
		  "value": "A"
		}
	  ],
	  "snmpDomain": [
		{
		  "id": "example-20a2-example-example-4eefa4e22898",
		  "value": "example"
		}
	  ],
	  "sysObjID": [
		{
		  "id": "example-example-4b5f-example-fde71caa478b",
		  "value": "1.2.3.4.5.6.7"
		}
	  ],
	  "deviceFamily": [
		{
		  "id": "example-example-example-9a91-7b8ff62e9378",
		  "value": "example OS"
		}
	  ],
	  "userName": [
		{
		  "id": "example-example-4171-example-fb8d25a49de5",
		  "value": "example"
		}
	  ],
	  "assetTag": [
		{
		  "id": "example-2958-example-example-99b18ee05fa0",
		  "value": "12345678"
		}
	  ],
	  "createdAt": [
		{
		  "id": "example-example-example-example-543efdc271fe",
		  "value": "2020-08-20T04:16:08.7220998Z"
		}
	  ]
	}
  }
]`

const dataProperties = `[
	{
	  "id": "example",
	  "value": "example",
	  "label": "pk"
	},
	{
	  "id": "example",
	  "value": ".36",
	  "label": "deviceIPAddress"
	},
	{
	  "id": "example3",
	  "value": "No",
	  "label": "critical"
	},
	{
	  "id": "example-f57f0a21dfde",
	  "value": "example_bldg-example",
	  "label": "placeInNetwork"
	},
	{
	  "id": "example-example-example-example-eb02386bacd3",
	  "value": "example",
	  "label": "vendor"
	},
	{
	  "id": "example-example-example-b150-976553721ca4",
	  "value": "example 10506e",
	  "label": "deviceModel"
	},
	{
	  "id": "example-example-example-9c7e-50bc3ab37313",
	  "value": "Switch",
	  "label": "deviceTypeName"
	},
	{
	  "id": "example-example-4698-example-1ee85c528c66",
	  "value": "A",
	  "label": "region"
	},
	{
	  "id": "example-20a2-example-example-4eefa4e22898",
	  "value": "example",
	  "label": "snmpDomain"
	},
	{
	  "id": "example-example-4b5f-example-fde71caa478b",
	  "value": "1.2.3.4.5.6.7",
	  "label": "sysObjID"
	},
	{
	  "id": "example-example-example-9a91-7b8ff62e9378",
	  "value": "example OS",
	  "label": "deviceFamily"
	},
	{
	  "id": "example-20a2-example-example-4eefa4e22898",
	  "value": "example",
	  "label": "userName"
	},
	{
	  "id": "example-2958-example-example-99b18ee05fa0",
	  "value": "12345678",
	  "label": "assetTag"
	},
	{
	  "id": "example-example-example-example-543efdc271fe",
	  "value": "2020-08-20T04:16:08.7220998Z",
	  "label": "createdAt"
	}
  ]`

const dataValues = `[
	"example",
	".36",
	"No",
	"example_bldg-example",
	"example",
	"example 10506e",
	"Switch",
	"A",
	"example",
	"1.2.3.4.5.6.7",
	"example OS",
	"example",
	"12345678",
	"2020-08-20T04:16:08.7220998Z"
  ]`

const dataValueMap = `[{
	"pk" : ["example"],
	"deviceIPAddress" : [".36"],
	"critical" : ["No"],
	"placeInNetwork" : ["example_bldg-example"],
	"vendor" : ["example"],
	"deviceModel" : ["example 10506e"],
	"deviceTypeName" : ["Switch"],
	"region" : ["A"],
	"snmpDomain" : ["example"],
	"sysObjID" : ["1.2.3.4.5.6.7"],
	"deviceFamily" : ["example OS"],
	"userName" : ["example"],
	"assetTag" : ["12345678"],
	"createdAt" : ["2020-08-20T04:16:08.7220998Z"]
  }]`

func TestToPropertiesMulti(t *testing.T) {
	t.Parallel()
	// GIVEN
	// the expected props is a map <label,value>. Its values have to match the given input for the test
	// which is 'dataProperties'
	expectedProps := map[string]string{
		"pk":              "example",
		"deviceIPAddress": ".36",
		"critical":        "No",
		"placeInNetwork":  "example_bldg-example",
		"vendor":          "example",
		"deviceModel":     "example 10506e",
		"deviceTypeName":  "Switch",
		"region":          "A",
		"snmpDomain":      "example",
		"sysObjID":        "1.2.3.4.5.6.7",
		"deviceFamily":    "example OS",
		"userName":        "example",
		"assetTag":        "12345678",
		"createdAt":       "2020-08-20T04:16:08.7220998Z",
	}

	// WHEN
	properties, err := ToProperties([]byte(dataProperties))

	// THEN
	assert.NoError(t, err)
	assert.Len(t, properties, 14)
	for _, prop := range properties {
		assert.Equal(t, expectedProps[prop.Label], prop.Value.AsString())
	}
}

func TestToValuesMulti(t *testing.T) {
	t.Parallel()
	// GIVEN
	// the expected props is a list of values.
	// Its values have to match the given input for the test which is 'dataValues'
	expectedProps := []string{
		"example",
		".36",
		"No",
		"example_bldg-example",
		"example",
		"example 10506e",
		"Switch",
		"A",
		"example",
		"1.2.3.4.5.6.7",
		"example OS",
		"example",
		"12345678",
		"2020-08-20T04:16:08.7220998Z",
	}

	// WHEN
	values, err := ToValues([]byte(dataValues))

	// THEN
	assert.NoError(t, err)
	assert.Len(t, values, 14)
	for _, val := range values {
		assert.Contains(t, expectedProps, val.AsString())
	}
}

func TestToValueMapMulti(t *testing.T) {
	t.Parallel()
	// GIVEN
	// the expected valueMap is a map <label,value>. Its values have to match the given input for the test
	// which is 'dataValueMap'
	expectedValueMap := map[string]string{
		"pk":              "example",
		"deviceIPAddress": ".36",
		"critical":        "No",
		"placeInNetwork":  "example_bldg-example",
		"vendor":          "example",
		"deviceModel":     "example 10506e",
		"deviceTypeName":  "Switch",
		"region":          "A",
		"snmpDomain":      "example",
		"sysObjID":        "1.2.3.4.5.6.7",
		"deviceFamily":    "example OS",
		"userName":        "example",
		"assetTag":        "12345678",
		"createdAt":       "2020-08-20T04:16:08.7220998Z",
	}
	// WHEN
	valueMap, err := ToValueMap([]byte(dataValueMap))

	// THEN
	assert.NoError(t, err)
	assert.Len(t, valueMap, 14)
	for key, value := range valueMap {
		assert.Equal(t, expectedValueMap[key], value.AsString())
	}
}

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
	assert.Equal(t, int32(1234), valueMap["a number"].AsInt32())
	assert.Equal(t, "max.mustermann@example.com", valueMap["email"].AsString())
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
	assert.Equal(t, int32(1234), values[1].AsInt32())
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

func TestToPropertiesFailMapProperties(t *testing.T) {
	t.Parallel()
	// GIVEN

	// WHEN
	properties, err := ToProperties([]byte(dataVertices))

	// THEN
	assert.Error(t, err)
	assert.Empty(t, properties)
}

func TestToPropertiesFailMapEdges(t *testing.T) {
	t.Parallel()
	// GIVEN

	// WHEN
	edges, err := ToEdges([]byte(dataVertices))

	// THEN
	assert.Error(t, err)
	assert.Empty(t, edges)
}

func TestToPropertiesFailMapValues(t *testing.T) {
	t.Parallel()
	// GIVEN
	data := "invalid"

	// WHEN
	values, err := ToValues([]byte(data))

	// THEN
	assert.Error(t, err)
	assert.Empty(t, values)
}

func TestToPropertiesFailMapValueMap(t *testing.T) {
	t.Parallel()
	// GIVEN

	// WHEN
	valuemap, err := ToValueMap([]byte(dataVertices))

	// THEN
	assert.Error(t, err)
	assert.Empty(t, valuemap)
}

func TestToPropertiesFailMapVertices(t *testing.T) {
	t.Parallel()
	// GIVEN
	dataValues := `[{
		"id":"8fff9259-09e6-4ea5-aaf8-250b31cc7f44|pk",
		"value":"prop value",
		"label":"prop key"
	}]`

	// WHEN
	vertices, err := ToVertices([]byte(dataValues))

	// THEN
	assert.Error(t, err)
	assert.Empty(t, vertices)
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

func TestToVertex(t *testing.T) {
	data := map[string]interface{}{
		"type":  "vertex",
		"id":    "the id",
		"label": "a vertex",
		"properties": map[string][]ValueWithID{
			"prop1": []ValueWithID{
				ValueWithID{
					ID:    "1234",
					Value: TypedValue{Value: "hello"},
				},
			},
		},
	}

	vertex, err := ToVertex(data)
	require.NoError(t, err)
	assert.Equal(t, Type("vertex"), vertex.Type)
	assert.Equal(t, "the id", vertex.ID)
	assert.Equal(t, "a vertex", vertex.Label)
	require.Len(t, vertex.Properties, 1)
	require.Len(t, vertex.Properties["prop1"], 1)
	assert.Equal(t, "hello", vertex.Properties["prop1"][0].Value.AsString())
}

func TestToVertex_Fail(t *testing.T) {
	_, err := ToVertex("not the right input type")
	assert.Error(t, err)

	data := map[string]interface{}{
		"missing": "vertex properties",
	}
	_, err = ToVertex(data)
	assert.Error(t, err)
}
