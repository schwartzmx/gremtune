package api

import (
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

func mapStructToValue(s map[string]interface{}) (TypedValue, error) {
	var result TypedValue

	config := &mapstructure.DecoderConfig{
		Result:           &result,
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return TypedValue{}, err
	}

	if err := decoder.Decode(s); err != nil {
		return TypedValue{}, err
	}
	return result, nil
}

func mapStructToProperty(s map[string]interface{}) (Property, error) {
	var result Property

	config := &mapstructure.DecoderConfig{
		Result:           &result,
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return Property{}, err
	}

	if err := decoder.Decode(s); err != nil {
		return Property{}, err
	}
	return result, nil
}

func mapStructToVertex(s map[string]interface{}) (Vertex, error) {
	var result Vertex

	config := &mapstructure.DecoderConfig{
		Result:           &result,
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return Vertex{}, err
	}

	if err := decoder.Decode(s); err != nil {
		return Vertex{}, err
	}
	return result, nil
}

func mapStructToEdge(s map[string]interface{}) (Edge, error) {
	var result Edge

	config := &mapstructure.DecoderConfig{
		Result:           &result,
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return Edge{}, err
	}

	if err := decoder.Decode(s); err != nil {
		return Edge{}, err
	}
	return result, nil
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

func toValues(input []interface{}) ([]TypedValue, error) {
	if input == nil {
		return nil, fmt.Errorf("Data is nil")
	}

	result := make([]TypedValue, 0, len(input))
	for _, element := range input {
		value, err := toValue(element)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}

	return result, nil
}

func ToPropertiesNew(input []byte) ([]Property, error) {
	if input == nil {
		return nil, fmt.Errorf("Data is nil")
	}

	parsedInput := make([]interface{}, 0)
	if err := json.Unmarshal(input, &parsedInput); err != nil {
		return nil, err
	}

	result := make([]Property, 0, len(parsedInput))
	for _, element := range parsedInput {
		value, err := toValue(element)
		if err != nil {
			return nil, err
		}

		prop := value.Value.(map[string]interface{})
		prop2, _ := mapStructToProperty(prop)
		result = append(result, prop2)
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
		value, err := toValue(element)
		if err != nil {
			return nil, err
		}
		// TODO: Check for @type
		prop := value.Value.(map[string]interface{})
		prop2, _ := mapStructToVertex(prop)
		result = append(result, prop2)
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
		value, err := toValue(element)
		if err != nil {
			return nil, err
		}
		// TODO: Check for @type
		prop := value.Value.(map[string]interface{})
		prop2, _ := mapStructToEdge(prop)
		result = append(result, prop2)
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

// toValue converts the given type to a TypedValue
func toValue(input interface{}) (TypedValue, error) {
	switch v := input.(type) {
	case string:
		return TypedValue{
			Type:  TypeString,
			Value: v,
		}, nil
	case bool:
		return TypedValue{
			Type:  TypeBool,
			Value: v,
		}, nil
	case map[string]interface{}:
		return mapStructToValue(v)
	default:
		return TypedValue{}, fmt.Errorf("Unknown type %T, can't process element: %v", v, v)
	}
}
