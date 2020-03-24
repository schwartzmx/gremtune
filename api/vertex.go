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

func NewVertexG(g interfaces.Graph) interfaces.Vertex {
	queryBuilders := make([]interfaces.QueryBuilder, 0)
	queryBuilders = append(queryBuilders, g)

	return &vertex{
		builders: queryBuilders,
	}
}

func NewVertexE(e interfaces.Edge) interfaces.Vertex {
	queryBuilders := make([]interfaces.QueryBuilder, 0)
	queryBuilders = append(queryBuilders, e)

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

// Drop adds .drop(), to the query. The query call will drop/ delete all referenced entities
func (v *vertex) Drop() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".drop()"))
}

// AddE adds .addE(<label>), to the query. The query call will be the first step to add an edge
func (v *vertex) AddE(label string) interfaces.Edge {
	v.Add(NewSimpleQB(".addE('%s')", label))
	return NewEdgeV(v)
}

func (v *vertex) Profile() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".executionProfile()"))
}

func (v *vertex) HasInt(key string, value int) interfaces.Vertex {
	return v.Add(NewSimpleQB(".has('%s',%d)", key, value))
}

func (v *vertex) PropertyInt(key string, value int) interfaces.Vertex {
	return v.Add(NewSimpleQB(".property('%s',%d)", key, value))
}

// OutE adds .outE(), to the query. The query call returns all outgoing edges of the Vertex
func (v *vertex) OutE() interfaces.Edge {
	v.Add(NewSimpleQB(".outE()"))
	return NewEdgeV(v)
}

// InE adds .inE(), to the query. The query call returns all incoming edges of the Vertex
func (v *vertex) InE() interfaces.Edge {
	v.Add(NewSimpleQB(".inE()"))
	return NewEdgeV(v)
}

// Count adds .count(), to the query. The query call will return the number of entities found in the query.
func (v *vertex) Count() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".count()"))
}

// PropertyList adds .property(list,'<key>','<value>'), e.g. .property(list, 'name','hans'), to the query. The query call will add the given property.
func (v *vertex) PropertyList(key, value string) interfaces.Vertex {
	return v.Add(NewSimpleQB(".property(list,'%s','%s')", key, value))
}
