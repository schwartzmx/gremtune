package api

import "github.com/supplyon/gremcos/interfaces"

func InE(labels ...string) interfaces.Edge {
	query := multiParamQuery("__.inE", labels...)
	return &edge{
		builders: []interfaces.QueryBuilder{query},
	}
}

func OutE(labels ...string) interfaces.Edge {
	query := multiParamQuery("__.outE", labels...)
	return &edge{
		builders: []interfaces.QueryBuilder{query},
	}
}

func Unfold() interfaces.QueryBuilder {
	query := NewSimpleQB("__.unfold()")
	return &edge{
		builders: []interfaces.QueryBuilder{query},
	}
}

func AddV(label string) interfaces.Vertex {
	query := NewSimpleQB(".addV(\"%s\")", label)
	return &vertex{
		builders: []interfaces.QueryBuilder{query},
	}
}

func Unfold() interfaces.QueryBuilder {
	query := NewSimpleQB("__.unfold()")
	return &edge{
		builders: []interfaces.QueryBuilder{query},
	}
}
