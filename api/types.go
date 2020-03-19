package api

import (
	"fmt"
)

type Property struct {
	ID    string     `mapstructure:"id"`
	Value TypedValue `mapstructure:"value,squash"`
	Label string     `mapstructure:"label"`
}

type Edge struct {
	ID        string `mapstructure:"id"`
	Label     string `mapstructure:"label"`
	Type      Type   `mapstructure:"type"`
	InVLabel  string `mapstructure:"inVLabel"`
	InV       string `mapstructure:"inV"`
	OutVLabel string `mapstructure:"outVLabel"`
	OutV      string `mapstructure:"outV"`
}

type Vertex struct {
	Type       Type              `mapstructure:"type"`
	ID         string            `mapstructure:"id"`
	Label      string            `mapstructure:"label"`
	Properties VertexPropertyMap `mapstructure:"properties"`
}

type ValueWithID struct {
	ID    string     `mapstructure:"id"`
	Value TypedValue `mapstructure:"value,squash"`
}

type VertexPropertyMap map[string][]ValueWithID

type VertexProperty struct {
	ValueWithID
	Label string `mapstructure:"label"`
}

func (p VertexProperty) String() string {
	return fmt.Sprintf("[%s] '%s':'%s'", p.ID, p.Label, p.Value)
}

func (v Vertex) String() string {
	return fmt.Sprintf("%s %s (props %v - type %s", v.ID, v.Label, v.Properties, v.Type)
}

func (e Edge) String() string {
	return fmt.Sprintf("%s (%s)-%s->%s (%s) - type %s", e.InVLabel, e.InV, e.Label, e.OutVLabel, e.OutV, e.Type)
}
