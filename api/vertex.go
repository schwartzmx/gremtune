package api

import (
	"fmt"

	"github.com/supplyon/gremcos/interfaces"
)

type vertex struct {
	builders []interfaces.QueryBuilder
}

func (v *vertex) String() string {

	queryString := ""
	for _, queryBuilder := range v.builders {
		queryString += fmt.Sprintf("%s", queryBuilder)
	}

	return queryString
}

func (v *vertex) has(key, value string) *vertex {
	v.builders = append(v.builders, NewSimpleQB(".has('%s','%s')", key, value))
	return v
}

func (v *vertex) HasLabel(vertexLabel string) interfaces.Vertex {
	v.builders = append(v.builders, NewSimpleQB(".hasLabel('%s')", vertexLabel))
	return v
}

func (v *vertex) valuesBy(label string) *vertex {
	v.builders = append(v.builders, NewSimpleQB(".values('%s')", label))
	return v
}

func (v *vertex) values() *vertex {
	v.builders = append(v.builders, NewSimpleQB(".values()"))
	return v
}

func (v *vertex) valueMap() *vertex {
	v.builders = append(v.builders, NewSimpleQB(".valueMap()"))
	return v
}

func (v *vertex) properties() *vertex {
	v.builders = append(v.builders, NewSimpleQB(".properties()"))
	return v
}

func (v *vertex) property(key, value string) *vertex {
	v.builders = append(v.builders, NewSimpleQB(".property('%s','%s')", key, value))
	return v
}
