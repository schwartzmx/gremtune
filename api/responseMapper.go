package api

import "github.com/supplyon/gremcos/interfaces"

// ResponseArray an array type for responses
type ResponseArray []interfaces.Response

// ToValues converts the given ResponseArray into an array of TypedValue type.
// The method will fail in case the data in the given ResponseArray does not contain primitive values.
func (responses ResponseArray) ToValues() ([]TypedValue, error) {
	result := make([]TypedValue, 0)
	for _, response := range responses {
		values, err := ToValues(response.Result.Data)
		if err != nil {
			return nil, err
		}
		result = append(result, values...)
	}
	return result, nil
}

// ToProperties converts the given ResponseArray into an array of Property type.
// The method will fail in case the data in the given ResponseArray does not contain values of type property.
func (responses ResponseArray) ToProperties() ([]Property, error) {
	result := make([]Property, 0)
	for _, response := range responses {
		properties, err := ToProperties(response.Result.Data)
		if err != nil {
			return nil, err
		}
		result = append(result, properties...)
	}
	return result, nil
}

// ToVertices converts the given ResponseArray into an array of Vertex type.
// The method will fail in case the data in the given ResponseArray does not contain values of type vertex.
func (responses ResponseArray) ToVertices() ([]Vertex, error) {
	result := make([]Vertex, 0)
	for _, response := range responses {
		vertices, err := ToVertices(response.Result.Data)
		if err != nil {
			return nil, err
		}
		result = append(result, vertices...)
	}
	return result, nil
}

// ToEdges converts the given ResponseArray into an array of Edge type.
// The method will fail in case the data in the given ResponseArray does not contain values of type edge.
func (responses ResponseArray) ToEdges() ([]Edge, error) {
	result := make([]Edge, 0)
	for _, response := range responses {
		edges, err := ToEdges(response.Result.Data)
		if err != nil {
			return nil, err
		}
		result = append(result, edges...)
	}
	return result, nil
}
