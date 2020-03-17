package api

import (
	"encoding/json"
	"fmt"

	"github.com/supplyon/gremcos/interfaces"
)

type Type string

const (
	TypeVertex Type = "g:Vertex"
	TypeInt64  Type = "g:Int64"
)

type ResponseVertexArray []ResponseVertex

type ResponseVertex struct {
	Plain string `json:",omitempty"`
	//Type  Type   `json:"@type,omitempty"`
	//Value Value `json:"@value,omitempty"`
}

type Value struct {
	Id    Id     `json:"id,omitempty"`
	Label string `json:"label,omitempty"`
}

type Id struct {
	Type  Type `json:"@type,omitempty"`
	Value int  `json:"@value,omitempty"`
}

//func oneToValues(responses []interfaces.Response) (ResponseVertexArray, error) {
//
//}

func ToVertex(responses []interfaces.Response) (ResponseVertexArray, error) {
	vertexArray := make(ResponseVertexArray, 0)
	for _, response := range responses {
		vertexArrayPart, err := oneToVertex(response)
		if err != nil {
			return ResponseVertexArray{}, err
		}

		vertexArray = append(vertexArray, vertexArrayPart...)
	}
	return vertexArray, nil
}

func oneToVertex(response interfaces.Response) (ResponseVertexArray, error) {
	//fmt.Printf("%v", response)

	if response.Result.Data == nil {
		return ResponseVertexArray{}, fmt.Errorf("Given response.result.data is nil")
	}

	result := ResponseVertexArray{}
	if err := json.Unmarshal(response.Result.Data, &result); err != nil {
		return ResponseVertexArray{}, err
	}

	return result, nil
}
