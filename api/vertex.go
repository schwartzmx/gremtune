package api

import (
	"github.com/supplyon/gremcos/interfaces"
)

type vertex struct {
	builders []interfaces.QueryBuilder
}

func (v *vertex) String() string {

	queryString := ""
	for _, queryBuilder := range v.builders {
		queryString += queryBuilder.String()
	}

	return queryString
}

func NewVertex(g interfaces.Graph) interfaces.Vertex {
	queryBuilders := make([]interfaces.QueryBuilder, 0)
	queryBuilders = append(queryBuilders, g)

	return &vertex{
		builders: queryBuilders,
	}
}

func (v *vertex) Add(builder interfaces.QueryBuilder) interfaces.Vertex {
	v.builders = append(v.builders, builder)
	return v
}

func (v *vertex) Has(key, value string) interfaces.Vertex {
	return v.Add(NewSimpleQB(".has('%s','%s')", key, value))
}

func (v *vertex) HasLabel(vertexLabel string) interfaces.Vertex {
	return v.Add(NewSimpleQB(".hasLabel('%s')", vertexLabel))
}

func (v *vertex) ValuesBy(label string) interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".values('%s')", label))
}

func (v *vertex) Values() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".values()"))
}

func (v *vertex) ValueMap() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".valueMap()"))
}

func (v *vertex) Properties() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".properties()"))
}

func (v *vertex) Property(key, value string) interfaces.Vertex {
	return v.Add(NewSimpleQB(".property('%s','%s')", key, value))
}

func (v *vertex) Id() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".id()"))
}
