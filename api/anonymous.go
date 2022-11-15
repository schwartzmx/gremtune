package api

import (
	"github.com/pkg/errors"
	"github.com/supplyon/gremcos/interfaces"
	"reflect"
)

// WithinInt adds .within([<value_1>,<value_1>,..,<value_n>]), to the query. Where values are of type int.
func WithinInt(values ...int) interfaces.QueryBuilder {
	return multiParamQueryInt("within", values...)
}

// Within adds .within([<value_1>,<value_1>,..,<value_n>]), to the query. Where values are of type string.
func Within(values ...string) interfaces.QueryBuilder {
	return multiParamQuery("within", values...)
}

// Eq adds .eq(<T>) to the query. (equal)
func Eq[T any](v T) interfaces.QueryBuilder {
	if t := reflect.TypeOf(v).String(); t == "string" {
		return NewSimpleQB("eq(\"%v\")", v)
	}
	return NewSimpleQB("eq(%v)", v)
}

// Neq adds .neq(<T>) to the query. (not equal)
func Neq[T Ordered](v T) interfaces.QueryBuilder {
	if t := reflect.TypeOf(v).String(); t == "string" {
		return NewSimpleQB("neq(\"%v\")", v)
	}
	return NewSimpleQB("neq(%v)", v)
}

// Lt adds .lt(<T>) to the query. (less than)
func Lt[T Ordered](v T) interfaces.QueryBuilder {
	if t := reflect.TypeOf(v).String(); t == "string" {
		return NewSimpleQB("lt(\"%v\")", v)
	}
	return NewSimpleQB("lt(%v)", v)
}

// Lte adds .lte(<T>) to the query. (less than equal)
func Lte[T Ordered](v T) interfaces.QueryBuilder {
	if t := reflect.TypeOf(v).String(); t == "string" {
		return NewSimpleQB("lte(\"%v\")", v)
	}
	return NewSimpleQB("lte(%v)", v)
}

// Gt adds .gt(<T>) to the query. (greater than)
func Gt[T Ordered](v T) interfaces.QueryBuilder {
	if t := reflect.TypeOf(v).String(); t == "string" {
		return NewSimpleQB("gt(\"%v\")", v)
	}
	return NewSimpleQB("gt(%v)", v)
}

// Gte adds .gte(<T>) to the query. (greater than equal)
func Gte[T Ordered](v T) interfaces.QueryBuilder {
	if t := reflect.TypeOf(v).String(); t == "string" {
		return NewSimpleQB("gte(\"%v\")", v)
	}
	return NewSimpleQB("gte(%v)", v)
}

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

// OutV adds .outV(), to the query. The query call returns all outgoing vertex of the edge
func OutV() interfaces.Vertex {
	query := NewSimpleQB("__.outV()")
	return &vertex{
		builders: []interfaces.QueryBuilder{query},
	}
}

// InV adds .inV(), to the query. The query call returns all incoming vertex of the edge
func InV() interfaces.Vertex {
	query := NewSimpleQB("__.inV()")
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

// Has adds .has("<key>","<value>"), e.g. .has("name","hans") depending on the given type the quotes for the value are omitted.
// e.g. .has("temperature",23.02) or .has("available",true)
// The method can also be used to return vertices that have a certain property.
// Then .has("<prop name>") will be added to the query.
//
//	v.Has("prop1")
func Has(key string, value ...interface{}) interfaces.QueryBuilder {
	if len(value) == 0 {
		return NewSimpleQB("__.has(\"%s\")", key)
	}

	keyVal, err := toKeyValueString(key, value[0])
	if err != nil {
		panic(errors.Wrapf(err, "cast has value %T to string failed (You could either implement the Stringer interface for this type or cast it to string beforehand)", value))
	}

	return NewSimpleQB("__.has%s", keyVal)
}
