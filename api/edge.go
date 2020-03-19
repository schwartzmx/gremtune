package api

import (
	"github.com/supplyon/gremcos/interfaces"
)

type edge struct {
	builders []interfaces.QueryBuilder
}

func NewEdgeV(v interfaces.Vertex) interfaces.Edge {
	queryBuilders := make([]interfaces.QueryBuilder, 0)
	queryBuilders = append(queryBuilders, v)

	return &edge{
		builders: queryBuilders,
	}
}

func NewEdgeG(g interfaces.Graph) interfaces.Edge {
	queryBuilders := make([]interfaces.QueryBuilder, 0)
	queryBuilders = append(queryBuilders, g)

	return &edge{
		builders: queryBuilders,
	}
}

func (e *edge) String() string {
	queryString := ""
	for _, queryBuilder := range e.builders {
		queryString += queryBuilder.String()
	}
	return queryString
}

// Add can be used to add a custom QueryBuilder
// e.g. g.V().Add(NewSimpleQB(".myCustomCall('%s')",label))
func (e *edge) Add(builder interfaces.QueryBuilder) interfaces.Edge {
	e.builders = append(e.builders, builder)
	return e
}

// To adds .to(<vertex>), to the query. The query call will be the second step to add an edge
func (e *edge) To(v interfaces.Vertex) interfaces.Edge {
	return e.Add(NewSimpleQB(".to(%s)", v))
}

// From adds .from(<vertex>), to the query. The query call will be the second step to add an edge
func (e *edge) From(v interfaces.Vertex) interfaces.Edge {
	return e.Add(NewSimpleQB(".from(%s)", v))
}

// Drop adds .drop(), to the query. The query call will drop/ delete all referenced entities
func (e *edge) Drop() interfaces.QueryBuilder {
	return e.Add(NewSimpleQB(".drop()"))
}

// OutV adds .outV(), to the query. The query call will return the vertices on the outgoing side of this edge
func (e *edge) OutV() interfaces.Vertex {
	e.Add(NewSimpleQB(".outV()"))
	return NewVertexE(e)
}

// InV adds .inV(), to the query. The query call will return the vertices on the incoming side of this edge
func (e *edge) InV() interfaces.Vertex {
	e.Add(NewSimpleQB(".inV()"))
	return NewVertexE(e)
}

func (e *edge) Profile() interfaces.QueryBuilder {
	return e.Add(NewSimpleQB(".executionProfile()"))
}
