package api

import "github.com/supplyon/gremcos/interfaces"

// InE adds .inE([<label_1>,<label_2>,..,<label_n>]), to the query. The query call returns all incoming edges of the Vertex
func InE(labels ...string) interfaces.Edge {
	query := multiParamQuery("__.inE", labels...)
	return &edge{
		builders: []interfaces.QueryBuilder{query},
	}
}

// OutE adds .outE([<label_1>,<label_2>,..,<label_n>]), to the query. The query call returns all outgoing edges of the Vertex
func OutE(labels ...string) interfaces.Edge {
	query := multiParamQuery("__.outE", labels...)
	return &edge{
		builders: []interfaces.QueryBuilder{query},
	}
}

// OutV adds .outv(), to the query. The query call returns all outgoing vertex of the edge
func OutV() interfaces.Vertex {
	query := NewSimpleQB("__.outV()")
	return &vertex{
		builders: []interfaces.QueryBuilder{query},
	}
}

// Unfold adds .unfold() to the query.
func Unfold() interfaces.QueryBuilder {
	query := NewSimpleQB("__.unfold()")
	return &edge{
		builders: []interfaces.QueryBuilder{query},
	}
}

// AddV adds .addV('<label>'), e.g. .addV('user'), to the query. The query call adds a vertex with the given label and returns that vertex.
func AddV(label string) interfaces.Vertex {
	query := NewSimpleQB("__.addV(\"%s\")", label)
	return &vertex{
		builders: []interfaces.QueryBuilder{query},
	}
}

// Constant adds .constant() to the query.
func Constant(c string) interfaces.QueryBuilder {
	return NewSimpleQB("__.constant(\"%s\")", c)
}
