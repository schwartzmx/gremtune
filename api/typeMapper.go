package api

import (
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// mapStructToType converts the given map struct into the desired target type
// The target type has to be annotated with 'mapstructure' tags
// Example:
//
// type MyTargetType struct {
//  Field1 int `mapstructure:"field1"`
//  Field2 string `mapstructure:"another_field"`
// }
func mapStructToType(source map[string]interface{}, target interface{}) error {

	config := &mapstructure.DecoderConfig{
		Result:           target,
		WeaklyTypedInput: true,
		ErrorUnused:      true,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	if err := decoder.Decode(source); err != nil {
		return err
	}

	return nil
}

// toTypeArray converts a given byte slice into the provided slice of one type
// Example
//
// var vertices []Vertex
//
// err := toTypeArray(data, &vertices)
func toTypeArray(input []byte, target interface{}) error {
	if input == nil {
		return fmt.Errorf("Data is nil")
	}

	parsedInput := make([]interface{}, 0)
	if err := json.Unmarshal(input, &parsedInput); err != nil {
		return err
	}

	// just for the primitive type
	switch targetValue := target.(type) {
	case *[]TypedValue:
		values, err := toValues(parsedInput)
		if err != nil {
			return err
		}
		*targetValue = append(*targetValue, values...)
		return nil
	}

	// handling of complex types
	for _, element := range parsedInput {
		mapStrct, ok := element.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Failed to cast %v (%T) into map[string]interface{}", element, element)
		}

		switch targetValue := target.(type) {
		case *[]Property:
			var property Property
			if err := mapStructToType(mapStrct, &property); err != nil {
				return errors.Wrap(err, "Mapping of response to Property failed. Please ensure that the response contains only properties.")
			}
			*targetValue = append(*targetValue, property)
			return nil
		case *[]Edge:
			var edge Edge
			if err := mapStructToType(mapStrct, &edge); err != nil {
				return errors.Wrap(err, "Mapping of response to Edge failed. Please ensure that the response contains only edges.")
			}
			*targetValue = append(*targetValue, edge)
			return nil
		case *[]Vertex:
			var vertex Vertex
			if err := mapStructToType(mapStrct, &vertex); err != nil {
				return errors.Wrap(err, "Mapping of response to Vertex failed. Please ensure that the response contains only vertices.")
			}
			*targetValue = append(*targetValue, vertex)
			return nil
		default:
			return fmt.Errorf("Unexpected type %T", target)
		}
	}
	return nil
}

// ToValues converts the given input byte array into an array of TypedValue type.
// The method will fail in case the data in the given byte array does not contain primitive values.
func ToValues(input []byte) ([]TypedValue, error) {
	var typedValues []TypedValue
	if err := toTypeArray(input, &typedValues); err != nil {
		return nil, errors.Wrap(err, "Mapping of response to TypedValue failed. Please ensure that the response contains only primitive types.")
	}
	return typedValues, nil
}

// ToProperties converts the given input byte array into an array of Property type.
// The method will fail in case the data in the given byte array does not contain values of type property.
func ToProperties(input []byte) ([]Property, error) {
	var properties []Property
	if err := toTypeArray(input, &properties); err != nil {
		return nil, err
	}
	return properties, nil
}

// ToVertices converts the given input byte array into an array of Vertex type.
// The method will fail in case the data in the given byte array does not contain values of type vertex.
func ToVertices(input []byte) ([]Vertex, error) {
	var vertices []Vertex
	if err := toTypeArray(input, &vertices); err != nil {
		return nil, err
	}
	return vertices, nil
}

// ToEdges converts the given input byte array into an array of Edge type.
// The method will fail in case the data in the given byte array does not contain values of type edge.
func ToEdges(input []byte) ([]Edge, error) {
	var edges []Edge
	if err := toTypeArray(input, &edges); err != nil {
		return nil, err
	}
	return edges, nil
}

// ToValueMap converts the given input byte array into a map of TypedValue's.
// The method will fail in case the data in the given byte array does not consist of key value pairs where these values are primitive types.
func ToValueMap(input []byte) (map[string]TypedValue, error) {
	if input == nil {
		return nil, fmt.Errorf("Data is nil")
	}

	parsedInput := make([]map[string][]interface{}, 0)
	if err := json.Unmarshal(input, &parsedInput); err != nil {
		return nil, err
	}

	result := make(map[string]TypedValue)

	// a value map is usually an array of map[string][]interface{}
	for _, arrayElement := range parsedInput {
		for key, entry := range arrayElement {

			value, err := toValues(entry)
			if err != nil {
				return nil, errors.Wrap(err, "Mapping of response to map of TypedValue failed. Please ensure that the response is a map of primitive types.")
			}

			if len(value) != 1 {
				return nil, fmt.Errorf("Unable to convert value map entry: %s %v", key, entry)
			}
			result[key] = value[0]
		}
	}

	return result, nil
}
