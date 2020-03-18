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

func ToValues(input []byte) ([]TypedValue, error) {
	if input == nil {
		return nil, fmt.Errorf("Data is nil")
	}

	parsedInput := make([]interface{}, 0)
	if err := json.Unmarshal(input, &parsedInput); err != nil {
		return nil, err
	}

	return toValues(parsedInput)
}

func ToProperties(input []byte) ([]VertexProperty, error) {
	if input == nil {
		return nil, fmt.Errorf("Input is nil")
	}

	parsedInput := make([]interface{}, 0)
	if err := json.Unmarshal(input, &parsedInput); err != nil {
		return nil, err
	}

	result := make([]VertexProperty, 0, len(parsedInput))
	for _, element := range parsedInput {
		var property VertexProperty
		if err := untypedToType(element, &property); err != nil {
			return nil, err
		}
		result = append(result, property)
	}

	return result, nil
}

func ToVertex(input []byte) ([]Vertex, error) {
	if input == nil {
		return nil, fmt.Errorf("Data is nil")
	}

	parsedInput := make([]interface{}, 0)
	if err := json.Unmarshal(input, &parsedInput); err != nil {
		return nil, err
	}

	result := make([]Vertex, 0, len(parsedInput))
	for _, element := range parsedInput {
		var vertex Vertex

		if err := untypedToType(element, &vertex); err != nil {
			return nil, err
		}

		result = append(result, vertex)
	}

	return result, nil
}

func ToEdge(input []byte) ([]Edge, error) {
	if input == nil {
		return nil, fmt.Errorf("Data is nil")
	}

	parsedInput := make([]interface{}, 0)
	if err := json.Unmarshal(input, &parsedInput); err != nil {
		return nil, err
	}

	result := make([]Edge, 0, len(parsedInput))
	for _, element := range parsedInput {
		var edge Edge
		if err := untypedToType(element, &edge); err != nil {
			return nil, err
		}
		result = append(result, edge)
	}

	return result, nil
}

// TODO: do the unmarshalling into the []map[string]interface{}/ map[string]interface{}
// outside of the concrete methods if possible
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
