package api

import (
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

func mapStructToType(s map[string]interface{}, target interface{}) error {

	config := &mapstructure.DecoderConfig{
		Result:           target,
		WeaklyTypedInput: true,
		ErrorUnused:      true,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	if err := decoder.Decode(s); err != nil {
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

func ToProperties(input []byte) ([]Property, error) {
	if input == nil {
		return nil, fmt.Errorf("Input is nil")
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

		if value.Type != TypeVertexProperty {
			return nil, fmt.Errorf("Expected type %s but got %s", TypeVertexProperty, value.Type)
		}

		prop := value.Value.(map[string]interface{})
		var property Property
		if err := mapStructToType(prop, &property); err != nil {
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
		value, err := toValue(element)
		if err != nil {
			return nil, err
		}

		if value.Type != TypeVertex {
			return nil, fmt.Errorf("Expected type %s but got %s", TypeVertex, value.Type)
		}

		vert := value.Value.(map[string]interface{})
		var vertex Vertex
		if err := mapStructToType(vert, &vertex); err != nil {
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
		value, err := toValue(element)
		if err != nil {
			return nil, err
		}

		if value.Type != TypeEdge {
			return nil, fmt.Errorf("Expected type %s but got %s", TypeEdge, value.Type)
		}

		ed := value.Value.(map[string]interface{})
		var edge Edge
		if err := mapStructToType(ed, &edge); err != nil {
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
