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

	queryBuilders := make([]queryBuilder, 0)
	queryBuilders = append(queryBuilders, g)
	queryBuilders = append(queryBuilders, newSimpleQB(".V()"))

	return &vertex{
		builders: queryBuilders,
	}
}

func (g *graph) VBy(id int) interfaces.Vertex {

	queryBuilders := make([]queryBuilder, 0)
	queryBuilders = append(queryBuilders, g)
	queryBuilders = append(queryBuilders, newSimpleQB(".V('%d')", id))

	return &vertex{
		builders: queryBuilders,
	}
}

func (g *graph) String() string {
	return g.name
}
