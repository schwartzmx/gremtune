package interfaces

import "github.com/gofrs/uuid"

// QueryBuilder can be used to generate queries for the cosmos db
type QueryBuilder interface {
	String() string
}

// Graph represents a QueryBuilder that can be used to create
// queries on graph level
type Graph interface {
	QueryBuilder

	// V adds .V() to the query. The query call returns all vertices.
	V() Vertex
	// VBy adds .V(<id>), e.g. .V(123), to the query. The query call returns the vertex with the given id.
	VBy(id int) Vertex
	// VByUUID adds .V(<id>), e.g. .V('8fff9259-09e6-4ea5-aaf8-250b31cc7f44'), to the query. The query call returns the vertex with the given id.
	VByUUID(id uuid.UUID) Vertex
	// VByStr adds .V(<id>), e.g. .V("123a"), to the query.  The query call returns the vertex with the given id.
	VByStr(id string) Vertex
	// AddV adds .addV('<label>'), e.g. .addV('user'), to the query. The query call adds a vertex with the given label and returns that vertex.
	AddV(label string) Vertex
	// E adds .E() to the query. The query call returns all edges.
	E() Edge
}

// Vertex represents a QueryBuilder that can be used to create
// queries on vertex level
type Vertex interface {
	QueryBuilder
	Dropper
	Profiler
	Counter

	// HasLabel adds .hasLabel('<label>'), e.g. .hasLabel('user'), to the query. The query call returns all vertices with the given label.
	HasLabel(vertexLabel string) Vertex
	// Property adds .property('<key>','<value>'), e.g. .property('name','hans'), to the query. The query call will add the given property.
	Property(key, value string) Vertex
	// PropertyInt adds .property('<key>',<int value>), e.g. .property('age',55), to the query. The query call will add the given property.
	PropertyInt(key string, value int) Vertex
	// PropertyList adds .property(list,'<key>','<value>'), e.g. .property(list, 'name','hans'), to the query. The query call will add the given property.
	PropertyList(key, value string) Vertex
	// Properties adds .properties(), to the query. The query call returns all properties of the vertex.
	Properties() QueryBuilder
	// Has adds .has('<key>','<value>'), e.g. .has('name','hans'), to the query. The query call returns all vertices
	// with the property which is defined by the given key value pair.
	Has(key, value string) Vertex
	// Has adds .has('<key>',<int value>), e.g. .has('age',55), to the query. The query call returns all vertices
	// with the property which is defined by the given key value pair.
	HasInt(key string, value int) Vertex
	// ValuesBy adds .values('<label>'), e.g. .values('user'), to the query. The query call returns all values of the vertex.
	ValuesBy(label string) QueryBuilder
	// Values adds .values(), to the query. The query call returns all values with the given label of the vertex.
	Values() QueryBuilder
	// ValueMap adds .valueMap(), to the query. The query call returns all values as a map of the vertex.
	ValueMap() QueryBuilder
	// Add can be used to add a custom QueryBuilder
	// e.g. g.V().Add(NewSimpleQB(".myCustomCall('%s')",label))
	Add(builder QueryBuilder) Vertex
	// Id adds .id(), to the query. The query call returns the id of the vertex.
	Id() QueryBuilder

	// AddE adds .addE(<label>), to the query. The query call will be the first step to add an edge
	AddE(label string) Edge

	// OutE adds .outE(), to the query. The query call returns all outgoing edges of the Vertex
	OutE() Edge

	// InE adds .inE(), to the query. The query call returns all incoming edges of the Vertex
	InE() Edge
}

type Edge interface {
	QueryBuilder
	Dropper
	Profiler
	Counter

	// To adds .to(<vertex>), to the query. The query call will be the second step to add an edge
	To(v Vertex) Edge
	// From adds .from(<vertex>), to the query. The query call will be the second step to add an edge
	From(v Vertex) Edge

	// OutV adds .outV(), to the query. The query call will return the vertices on the outgoing side of this edge
	OutV() Vertex
	// InV adds .inV(), to the query. The query call will return the vertices on the incoming side of this edge
	InV() Vertex
	// Add can be used to add a custom QueryBuilder
	// e.g. g.V().Add(NewSimpleQB(".myCustomCall('%s')",label))
	Add(builder QueryBuilder) Edge

	// HasLabel adds .hasLabel('<label>'), e.g. .hasLabel('user'), to the query. The query call returns all edges with the given label.
	HasLabel(label string) Edge
}

type Dropper interface {
	// Drop adds .drop(), to the query. The query call will drop/ delete all referenced entities
	Drop() QueryBuilder
}

type Profiler interface {
	// Profile adds .executionProfile(), to the query. The query call will return profiling information of the executed query
	Profile() QueryBuilder
}

type Counter interface {
	// Count adds .count(), to the query. The query call will return the number of entities found in the query.
	Count() QueryBuilder
}
