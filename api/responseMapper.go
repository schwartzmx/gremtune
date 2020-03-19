package api

import "github.com/supplyon/gremcos/interfaces"

type ResponseArray []interfaces.Response

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
