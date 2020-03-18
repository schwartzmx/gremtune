package api

import "fmt"

type VertexProperty struct {
	ID    ID     `mapstructure:"id"`
	Value string `mapstructure:"value"`
	Label string `mapstructure:"label"`
}

type Edge struct {
	ID        ID     `mapstructure:"id"`
	Label     string `mapstructure:"label"`
	InVLabel  string `mapstructure:"inVLabel"`
	InV       ID     `mapstructure:"inV"`
	OutVLabel string `mapstructure:"outVLabel"`
	OutV      ID     `mapstructure:"outV"`
}

type Vertex struct {
	ID    ID     `mapstructure:"id"`
	Label string `mapstructure:"label"`
}

type ID struct {
	Value int  `mapstructure:"@value"`
	Type  Type `mapstructure:"@type"`
}

func (p VertexProperty) String() string {
	return fmt.Sprintf("[%s] '%s':'%s'", p.ID, p.Label, p.Value)
}

func (id ID) String() string {
	return fmt.Sprintf("%d", id.Value)
}

func (v Vertex) String() string {
	return fmt.Sprintf("%s %s", v.ID, v.Label)
}

func (e Edge) String() string {
	return fmt.Sprintf("%s (%s)-%s->%s (%s)", e.InVLabel, e.InV, e.Label, e.OutVLabel, e.OutV)
}
