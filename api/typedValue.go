package api

import (
	"fmt"

	"github.com/spf13/cast"
)

// Type defines the cosmos db types
type Type string

const (
	TypeVertex Type = "vertex"
	TypeEdge   Type = "edge"
)

var complexTypes = map[Type]struct{}{
	TypeVertex: {},
	TypeEdge:   {},
}

// TypedValue is a value with a cosmos db type
type TypedValue struct {
	Value interface{}
}

// toValue converts the given input to a TypedValue
func toValue(input interface{}) (TypedValue, error) {
	return TypedValue{Value: input}, nil
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
	return fmt.Sprintf("%v", tv.Value)
}
