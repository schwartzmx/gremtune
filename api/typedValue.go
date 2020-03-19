package api

import (
	"fmt"
	"reflect"

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

func isComplexType(t Type) bool {
	_, ok := complexTypes[t]
	return ok
}

func isTypeMatching(source interface{}, expectedType Type) bool {
	switch expectedType {
	case TypeBool, TypeString, TypeFloat64, TypeInt32, TypeInt64:
		return reflect.TypeOf(source) == reflect.TypeOf(&TypedValue{})
	case TypeVertex:
		_, ok := source.(*Vertex)
		if !ok {
			return false
		}
		return true
	case TypeVertexProperty:
		_, ok := source.(*VertexProperty)
		if !ok {
			return false
		}
		return true
	case TypeEdge:
		_, ok := source.(*Edge)
		if !ok {
			return false
		}
		return true
	default:
		return false
	}
}

// TypedValue is a value with a cosmos db type
type TypedValue struct {
	Value interface{}
	Type  Type
}

// toValue converts the given input to a TypedValue
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
	case float64:
		return TypedValue{
			Type:  TypeFloat64,
			Value: v,
		}, nil
	case int32:
		return TypedValue{
			Type:  TypeInt32,
			Value: v,
		}, nil
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
	case TypeFloat64:
		return fmt.Sprintf("%f", tv.AsFloat64())
	case TypeString:
		return tv.AsString()
	default:
		return fmt.Sprintf("Unknown type=%T/%s, value=%v", tv.Value, tv.Type, tv.Value)
	}
}
