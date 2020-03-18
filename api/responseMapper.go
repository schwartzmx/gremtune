package api

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/supplyon/gremcos/interfaces"
)

func ToProperties(responses []interfaces.Response) (Properties, error) {
	if len(responses) == 0 {
		return Properties{}, nil
	}
	return oneToProperties(responses[0])
}

type Properties struct {
}

func decodeHook(source reflect.Type, target reflect.Type, input interface{}) (interface{}, error) {
	fmt.Printf("HOOK source %v target %v input %v\n", source, target, input)
	return "nil", nil
}

func parseValue(s map[string]interface{}) (TypedValue, error) {
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

func ParseResponse(input []byte) ([]TypedValue, error) {
	if input == nil {
		return nil, fmt.Errorf("Data is nil")
	}

	parsedInput := make([]interface{}, 0)
	if err := json.Unmarshal(input, &parsedInput); err != nil {
		return nil, err
	}

	result := make([]TypedValue, 0, len(parsedInput))
	for _, element := range parsedInput {
		value, err := parseElement(element)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}

	return result, nil
}

func parseElement(input interface{}) (TypedValue, error) {
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
		return parseValue(v)
	default:
		return TypedValue{}, fmt.Errorf("Unknown type %T, can't process element: %v", v, v)
	}
}

func oneToProperties(response interfaces.Response) (Properties, error) {

	data := make([]interface{}, 0)
	if err := json.Unmarshal(response.Result.Data, &data); err != nil {
		return Properties{}, err
	}

	fmt.Printf("SUCCESS %s\n\n\n", data)
	for _, d := range data {
		//fmt.Printf("[%d] %s - %T\n", i, d, d)

		switch v := d.(type) {
		case string:
			fmt.Printf("STRING %s\n", v)
		case bool:
			fmt.Printf("BOOL %t\n", v)
		case map[string]interface{}:
			res, _ := parseValue(v)
			fmt.Printf("RESULT is %s\n", res)
		default:
			fmt.Printf("UNKNOWN %T %v\n", v, v)
		}

	}

	return Properties{}, nil
}

//
//func oneToVertex(response interfaces.Response) (ResponseVertexArray, error) {
//	fmt.Printf("%s\n", response.Result.Data)
//
//	data := make([]interface{}, 0)
//	if err := json.Unmarshal(response.Result.Data, &data); err != nil {
//		return ResponseVertexArray{}, err
//	}
//	fmt.Printf("SUCCESS %s\n", data)
//
//	if response.Result.Data == nil {
//		return ResponseVertexArray{}, fmt.Errorf("Given response.result.data is nil")
//	}
//
//	result := ResponseVertexArray{}
//	if err := json.Unmarshal(response.Result.Data, &result); err != nil {
//		return ResponseVertexArray{}, err
//	}
//
//	return result, nil
//}
//
