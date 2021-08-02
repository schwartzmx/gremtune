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
// e.g. g.V().Add(NewSimpleQB(".myCustomCall("%s")",label))
func (e *edge) Add(builder interfaces.QueryBuilder) interfaces.Edge {
	e.builders = append(e.builders, builder)
	return e
}

// As adds .as([<label_1>,<label_2>,..,<label_n>]), to the query to label that query step for later access.
func (e *edge) As(labels ...string) interfaces.Edge {
	query := multiParamQuery(".as", labels...)
	return e.Add(query)
}

// Limit adds .limit(<num>), to the query. The query call will limit the results of the query to the given number.
func (e *edge) Limit(maxElements int) interfaces.Edge {
	return e.Add(NewSimpleQB(".limit(%d)", maxElements))
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

// Profile adds ..executionProfile(), to the query. The query call will return profiling information of the executed query
func (e *edge) Profile() interfaces.QueryBuilder {
	if !gUSE_COSMOS_DB_QUERY_LANGUAGE {
		return e.Add(NewSimpleQB(".profile()"))
	}
	return e.Add(NewSimpleQB(".executionProfile()"))
}

// HasLabel adds .hasLabel([<label_1>,<label_2>,..,<label_n>]), e.g. .hasLabel('user','name'), to the query. The query call returns all edges with the given label.
func (e *edge) HasLabel(labels ...string) interfaces.Edge {
	query := multiParamQuery(".hasLabel", labels...)
	return e.Add(query)
}

// Id adds .id()
func (e *edge) Id() interfaces.QueryBuilder {
	return e.Add(NewSimpleQB(".id()"))
}

// HasId adds .hasId('<id>'), e.g. .hasId('8aaaa410-dae1-4f33-8dd7-0217e69df10c'), to the query. The query call returns all edges
// with the given id.
func (e *edge) HasId(id string) interfaces.Edge {
	return e.Add(NewSimpleQB(".hasId(\"%s\")", id))
}

// Count adds .count(), to the query. The query call will return the number of entities found in the query.
func (e *edge) Count() interfaces.QueryBuilder {
	return e.Add(NewSimpleQB(".count()"))
}
