package api

import (
	"fmt"

	"github.com/supplyon/gremcos/interfaces"
)

type queryBuilder interface {
	String() string
}

type simpleQB struct {
	value string
}

func newSimpleQB(format string, a ...interface{}) queryBuilder {
	return &simpleQB{
		value: fmt.Sprintf(format, a...),
	}
}

func (sqb *simpleQB) String() string {
	return sqb.value
}

type vertex struct {
	builders []queryBuilder
}

func (v *vertex) String() string {

	queryString := ""
	for _, queryBuilder := range v.builders {
		queryString += fmt.Sprintf("%s", queryBuilder)
	}

	return queryString
}

func (v *vertex) has(key, value string) *vertex {
	v.builders = append(v.builders, newSimpleQB(".has('%s','%s')", key, value))
	return v
}

func (v *vertex) HasLabel(vertexLabel string) interfaces.Vertex {
	v.builders = append(v.builders, newSimpleQB(".hasLabel('%s')", vertexLabel))
	return v
}

func (v *vertex) valuesBy(label string) *vertex {
	v.builders = append(v.builders, newSimpleQB(".values('%s')", label))
	return v
}

func (v *vertex) values() *vertex {
	v.builders = append(v.builders, newSimpleQB(".values()"))
	return v
}

func (v *vertex) valueMap() *vertex {
	v.builders = append(v.builders, newSimpleQB(".valueMap()"))
	return v
}

func (v *vertex) properties() *vertex {
	v.builders = append(v.builders, newSimpleQB(".properties()"))
	return v
}

func (v *vertex) property(key, value string) *vertex {
	v.builders = append(v.builders, newSimpleQB(".property('%s','%s')", key, value))
	return v
}
