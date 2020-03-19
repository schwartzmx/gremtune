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

// Add can be used to add a custom QueryBuilder
// e.g. g.V().Add(NewSimpleQB(".myCustomCall('%s')",label))
func (v *vertex) Add(builder interfaces.QueryBuilder) interfaces.Vertex {
	v.builders = append(v.builders, builder)
	return v
}

// Has adds .has('<key>','<value>'), e.g. .has('name','hans')
func (v *vertex) Has(key, value string) interfaces.Vertex {
	return v.Add(NewSimpleQB(".has('%s','%s')", key, value))
}

// HasLabel adds .hasLabel('<label>'), e.g. .hasLabel('user')
func (v *vertex) HasLabel(vertexLabel string) interfaces.Vertex {
	return v.Add(NewSimpleQB(".hasLabel('%s')", vertexLabel))
}

// ValuesBy adds .values('<label>'), e.g. .values('user')
func (v *vertex) ValuesBy(label string) interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".values('%s')", label))
}

// Values adds .values()
func (v *vertex) Values() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".values()"))
}

// ValueMap adds .valueMap()
func (v *vertex) ValueMap() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".valueMap()"))
}

// Properties adds .properties()
func (v *vertex) Properties() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".properties()"))
}

// Property adds .property('<key>','<value>'), e.g. .property('name','hans')
func (v *vertex) Property(key, value string) interfaces.Vertex {
	return v.Add(NewSimpleQB(".property('%s','%s')", key, value))
}

// Id adds .id()
func (v *vertex) Id() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".id()"))
}