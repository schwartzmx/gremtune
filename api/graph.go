package api

import "github.com/supplyon/gremcos/interfaces"

// NewGraph creates a new graph query with the given name
// Hint: The actual graph has to exist on the server in order to execute the
// query that will be generated with this query builder
func NewGraph(name string) interfaces.Graph {
	return &graph{
		name: name,
	}
}

type graph struct {
	name string
}

// V adds .V()
func (g *graph) V() interfaces.Vertex {
	vertex := NewVertex(g)
	vertex.Add(NewSimpleQB(".V()"))
	return vertex
}

// VBy adds .V(<id>), e.g. .V(123)
func (g *graph) VBy(id int) interfaces.Vertex {
	vertex := NewVertex(g)
	vertex.Add(NewSimpleQB(".V('%d')", id))
	return vertex
}

// AddV adds .addV('<label>'), e.g. .addV('user')
func (g *graph) AddV(label string) interfaces.Vertex {
	vertex := NewVertex(g)
	vertex.Add(NewSimpleQB(".addV('%s')", label))
	return vertex
}

func (g *graph) String() string {
	return g.name
}
