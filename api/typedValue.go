package api

import (
	"fmt"

	"github.com/spf13/cast"
)

type Type string

const (
	TypeVertex Type = "g:Vertex"
	TypeInt64  Type = "g:Int64"
	TypeInt32  Type = "g:Int32"
	TypeString Type = "g:string"
	TypeBool   Type = "g:bool"
)

type TypedValue struct {
	Value interface{} `mapstructure:"@value"`
	Type  Type        `mapstructure:"@type"`
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
		return fmt.Sprintf("Unknown type=%T, value=%v", tv.Value, tv.Value)
	}
}
