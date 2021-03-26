package api

import (
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/supplyon/gremcos/interfaces"
)

type QueryLanguage string

const (
	QueryLanguageCosmosDB         QueryLanguage = "cosmos"
	QueryLanguageTinkerpopGremlin QueryLanguage = "tinkerpop"
)

var gUSE_COSMOS_DB_QUERY_LANGUAGE = true

// SetQueryLanguageTo sets the query language that shall be used.
// Per default QueryLanguageCosmosDB is in use.
func SetQueryLanguageTo(ql QueryLanguage) {
	gUSE_COSMOS_DB_QUERY_LANGUAGE = (ql == QueryLanguageCosmosDB)
}

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

// VByStr adds .V(<id>), e.g. .V("123a"), to the query.  The query call returns the vertex with the given id.
func (g *graph) VByStr(id string) interfaces.Vertex {
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

// multiParamQuery creates a query based on the given (optional) parameters.
// The query is the name of the query method that supports 0..* parameters.
// Examples:
//    q1:=multiParamQuery(".out","label1","label2") ==> generates ".out('label1','label2')"
//    q2:=multiParamQuery(".out") ==> generates ".out()"
func multiParamQuery(query string, params ...string) interfaces.QueryBuilder {
	if len(params) == 0 {
		return NewSimpleQB(fmt.Sprintf("%s()", query))
	}

	qStr := strings.Join(params, "\",\"")
	qStr = fmt.Sprintf("%s(\"%s\")", query, qStr)
	return NewSimpleQB(qStr)
}
