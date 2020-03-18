package api

import (
	"fmt"

	"github.com/spf13/cast"
)

// Type defines the cosmos db types
type Type string

const (
	TypeInt64          Type = "g:Int64"
	TypeInt32          Type = "g:Int32"
	TypeFloat64        Type = "g:Float64"
	TypeString         Type = "g:string"
	TypeBool           Type = "g:bool"
	TypeVertex         Type = "g:Vertex"
	TypeVertexProperty Type = "g:VertexProperty"
	TypeEdge           Type = "g:Edge"
)

var complexTypes = map[Type]struct{}{
	TypeVertex:         {},
	TypeVertexProperty: {},
	TypeEdge:           {},
}

func IsComplexType(t Type) bool {
	_, ok := complexTypes[t]
	return ok
}

// TypedValue is a value with a cosmos db type
type TypedValue struct {
	Value interface{} `mapstructure:"@value"`
	Type  Type        `mapstructure:"@type"`
}

// toValue converts the given input to a TypedValue
// Supported values are:
//
// * string: toValue("hello")
//
// * bool: toValue(true)
//
// * int32: toValue(map[string]interface{}{
//			"@type":  TypeInt32,
//			"@value": int32(11),
//		})
//
// * float64: toValue(map[string]interface{}{
//			"@type":  TypeFloat64,
//			"@value": float64(11),
//		})
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
		var value TypedValue
		if err := mapStructToType(v, &value); err != nil {
			return TypedValue{}, err
		}

		if len(value.Type) == 0 {
			return TypedValue{}, fmt.Errorf("Failed to decode type, expected field @type is missing")
		}

		if value.Value == nil {
			return TypedValue{}, fmt.Errorf("Failed to decode type, expected field @value is missing")
		}

		return value, nil
	default:
		return TypedValue{}, fmt.Errorf("Unknown type %T, can't process element: %v", v, v)
	}
}

// converts a list of values to TypedValue
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

func (tv TypedValue) AsFloat64E() (float64, error) {
	return cast.ToFloat64E(tv.Value)
}

func (tv TypedValue) AsFloat64() float64 {
	return cast.ToFloat64(tv.Value)
}

func (tv TypedValue) AsInt32E() (int32, error) {
	return cast.ToInt32E(tv.Value)
}

func (tv TypedValue) AsInt32() int32 {
	return cast.ToInt32(tv.Value)
}

func (tv TypedValue) AsBoolE() (bool, error) {
	return cast.ToBoolE(tv.Value)
}

func (tv TypedValue) AsBool() bool {
	return cast.ToBool(tv.Value)
}

func (tv TypedValue) AsStringE() (string, error) {
	return cast.ToStringE(tv.Value)
}

func (tv TypedValue) AsString() string {
	return cast.ToString(tv.Value)
}

func (tv TypedValue) String() string {
	switch tv.Type {
	case TypeInt32:
		return fmt.Sprintf("%d", tv.AsInt32())
	case TypeBool:
		return fmt.Sprintf("%t", tv.AsBool())
	case TypeString:
		return tv.AsString()
	default:
		return fmt.Sprintf("Unknown type=%T/%s, value=%v", tv.Value, tv.Type, tv.Value)
	}
}
