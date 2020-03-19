package api

import (
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
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

// untypedToType converts the given untyped value into a known type
func untypedToType(source interface{}, target interface{}) error {

	// extract the type information
	typedValue, err := toValue(source)
	if err != nil {
		return err
	}

	// verify the type
	if !isTypeMatching(target, typedValue.Type) {
		return fmt.Errorf("Expected type %T but got %s", target, typedValue.Type)
	}

	// if it is not a complex type we can stop here and return the TypedValue
	if !isComplexType(typedValue.Type) {
		targetAsTypedValue, ok := target.(*TypedValue)
		if !ok {
			return fmt.Errorf("%T is not %T", target, typedValue)
		}
		targetAsTypedValue.Value = typedValue.Value
		targetAsTypedValue.Type = typedValue.Type
		return nil
	}

	// cast the extracted typed value into a mapstruct
	mapStrct, ok := typedValue.Value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Failed to cast %v (%T) into map[string]interface{}", typedValue.Value, typedValue.Value)
	}

	// convert the mapstruct into the target type
	if err := mapStructToType(mapStrct, &target); err != nil {
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
				return err
			}
			*targetValue = append(*targetValue, property)
			return nil
		case *[]Edge:
			var edge Edge
			if err := mapStructToType(mapStrct, &edge); err != nil {
				return err
			}
			*targetValue = append(*targetValue, edge)
			return nil
		case *[]Vertex:
			var vertex Vertex
			if err := mapStructToType(mapStrct, &vertex); err != nil {
				return err
			}
			*targetValue = append(*targetValue, vertex)
			return nil
		default:
			return fmt.Errorf("Unexpected type %T", target)
		}
	}
	return nil
}

func ToValues(input []byte) ([]TypedValue, error) {
	var typedValues []TypedValue
	if err := toTypeArray(input, &typedValues); err != nil {
		return nil, err
	}
	return typedValues, nil
}

func ToProperties(input []byte) ([]Property, error) {
	var properties []Property
	if err := toTypeArray(input, &properties); err != nil {
		return nil, err
	}
	return properties, nil
}

func ToVertices(input []byte) ([]Vertex, error) {
	var vertices []Vertex
	if err := toTypeArray(input, &vertices); err != nil {
		return nil, err
	}
	return vertices, nil
}

func ToEdges(input []byte) ([]Edge, error) {
	var edges []Edge
	if err := toTypeArray(input, &edges); err != nil {
		return nil, err
	}
	return edges, nil
}

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
				return nil, err
			}

			if len(value) != 1 {
				return nil, fmt.Errorf("Unable to convert value map entry: %s %v", key, entry)
			}
			result[key] = value[0]
		}
	}

	return result, nil
}
