package api

import "github.com/supplyon/gremcos/interfaces"

func NewGraph(name string) interfaces.Graph {
	return &graph{
		name: name,
	}
}

type graph struct {
	name string
}

func (g *graph) V() interfaces.Vertex {
	vertex := NewVertex(g)
	vertex.Add(NewSimpleQB(".V()"))
	return vertex
}

func (g *graph) VBy(id int) interfaces.Vertex {
	vertex := NewVertex(g)
	vertex.Add(NewSimpleQB(".V('%d')", id))
	return vertex
}

func (g *graph) AddV(label string) interfaces.Vertex {
	vertex := NewVertex(g)
	vertex.Add(NewSimpleQB(".addV('%s')", label))
	return vertex
}

func (g *graph) String() string {
	return g.name
}
