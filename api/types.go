package api

import (
	"fmt"

	"github.com/spf13/cast"
)

// Ordered is a constraint that permits any ordered type: any type
// that supports the operators < <= >= >.
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~string
}

// Property represents the cosmos db type for a property.
// As it would be returned by a call to .properties().
type Property struct {
	ID    string     `mapstructure:"id"`
	Value TypedValue `mapstructure:"value,squash"`
	Label string     `mapstructure:"label"`
}

// Edge represents the cosmos DB type for an edge.
// As it would be returned by a call to g.E().
type Edge struct {
	ID        string `mapstructure:"id"`
	Label     string `mapstructure:"label"`
	Type      Type   `mapstructure:"type"`
	InVLabel  string `mapstructure:"inVLabel"`
	InV       string `mapstructure:"inV"`
	OutVLabel string `mapstructure:"outVLabel"`
	OutV      string `mapstructure:"outV"`
}

// Vertex represents the cosmos DB type for an vertex.
// As it would be returned by a call to g.V().
type Vertex struct {
	Type       Type              `mapstructure:"type"`
	ID         string            `mapstructure:"id"`
	Label      string            `mapstructure:"label"`
	Properties VertexPropertyMap `mapstructure:"properties"`
}

// ValueWithID represents the cosmos DB type for a value in case
// it is used/ attached to a complex type.
type ValueWithID struct {
	ID    string     `mapstructure:"id"`
	Value TypedValue `mapstructure:"value,squash"`
}

type VertexPropertyMap map[string][]ValueWithID

// Type defines the cosmos db complex types
type Type string

const (
	TypeVertex Type = "vertex"
	TypeEdge   Type = "edge"
)

// TypedValue represents the cosmos DB type for a value in case
// it is not used/ attached to a complex type.
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

func (tv TypedValue) AsInt64E() (int64, error) {
	return cast.ToInt64E(tv.Value)
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
	value, err := cast.ToStringE(tv.Value)
	if err != nil {
		return "", err
	}
	return UnEscape(value), nil
}

func (tv TypedValue) AsString() string {
	return UnEscape(cast.ToString(tv.Value))
}

func (tv TypedValue) String() string {
	return fmt.Sprintf("%v", tv.Value)
}

func (v Vertex) String() string {
	return fmt.Sprintf("%s %s (props %v - type %s", v.ID, v.Label, v.Properties, v.Type)
}

func (e Edge) String() string {
	return fmt.Sprintf("%s (%s)-%s->%s (%s) - type %s", e.InVLabel, e.InV, e.Label, e.OutVLabel, e.OutV, e.Type)
}

// Value returns the first value of the properties for this key
// the others are ignored. Anyway it is not possible to store multiple
// values for one property key.
func (vpm VertexPropertyMap) Value(key string) (ValueWithID, bool) {
	value, ok := vpm[key]
	if !ok {
		return ValueWithID{}, false
	}
	if len(value) == 0 {
		return ValueWithID{}, false
	}
	return value[0], true
}

func (vpm VertexPropertyMap) AsString(key string) (string, error) {
	value, ok := vpm.Value(key)
	if !ok {
		return "", fmt.Errorf("%s does not exist", key)
	}

	return value.Value.AsStringE()
}

func (vpm VertexPropertyMap) AsInt32(key string) (int32, error) {
	value, ok := vpm.Value(key)
	if !ok {
		return 0, fmt.Errorf("%s does not exist", key)
	}

	return value.Value.AsInt32E()
}
