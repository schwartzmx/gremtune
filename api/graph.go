package api

import (
	"github.com/gofrs/uuid"
	"github.com/supplyon/gremcos/interfaces"
)

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
	vertex := NewVertexG(g)
	vertex.Add(NewSimpleQB(".V()"))
	return vertex
}

// VBy adds .V(<id>), e.g. .V(123)
func (g *graph) VBy(id int) interfaces.Vertex {
	vertex := NewVertexG(g)
	vertex.Add(NewSimpleQB(".V(\"%d\")", id))
	return vertex
}

// VByUUID adds .V(<id>), e.g. .V("8fff9259-09e6-4ea5-aaf8-250b31cc7f44"), to the query. The query call returns the vertex with the given id.
func (g *graph) VByUUID(id uuid.UUID) interfaces.Vertex {
	vertex := NewVertexG(g)
	vertex.Add(NewSimpleQB(".V(\"%s\")", id))
	return vertex
}

// AddV adds .addV("<label>"), e.g. .addV("user")
func (g *graph) AddV(label string) interfaces.Vertex {
	vertex := NewVertexG(g)
	vertex.Add(NewSimpleQB(".addV(\"%s\")", label))
	return vertex
}

// E adds .E()
func (g *graph) E() interfaces.Edge {
	edge := NewEdgeG(g)
	edge.Add(NewSimpleQB(".E()"))
	return edge
}

func (g *graph) String() string {
	return g.name
}
