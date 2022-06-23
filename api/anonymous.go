package api

import (
	"github.com/pkg/errors"
	"github.com/supplyon/gremcos/interfaces"
)

// WithinInt adds .within([<value_1>,<value_1>,..,<value_n>]), to the query. Where values are of type int.
func WithinInt(values ...int) interfaces.QueryBuilder {
	query := multiParamQueryInt("within", values...)
	return &edge{
		builders: []interfaces.QueryBuilder{query},
	}
}

// Within adds .within([<value_1>,<value_1>,..,<value_n>]), to the query. Where values are of type string.
func Within(values ...string) interfaces.QueryBuilder {
	query := multiParamQuery("within", values...)
	return &edge{
		builders: []interfaces.QueryBuilder{query},
	}
}

// Eq adds .eq(<int>) to the query. (equal)
func Eq(v int) interfaces.QueryBuilder {
	return NewSimpleQB("eq(%d)", v)
}

// Neq adds .neq(<int>) to the query. (not equal)
func Neq(v int) interfaces.QueryBuilder {
	return NewSimpleQB("neq(%d)", v)
}

// Lt adds .lt(<int>) to the query. (less than)
func Lt(v int) interfaces.QueryBuilder {
	return NewSimpleQB("lt(%d)", v)
}

// Lte adds .lte(<int>) to the query. (less than equal)
func Lte(v int) interfaces.QueryBuilder {
	return NewSimpleQB("lte(%d)", v)
}

// Gt adds .gt(<int>) to the query. (greater than)
func Gt(v int) interfaces.QueryBuilder {
	return NewSimpleQB("gt(%d)", v)
}

// Gte adds .gte(<int>) to the query. (greater than equal)
func Gte(v int) interfaces.QueryBuilder {
	return NewSimpleQB("gte(%d)", v)
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
