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
	// AddV adds .addV('<label>'), e.g. .addV('user'), to the query. The query call adds a vertex with the given label and returns that vertex.
	AddV(label string) Vertex
}

// Vertex represents a QueryBuilder that can be used to create
// queries on vertex level
type Vertex interface {
	QueryBuilder
	// HasLabel adds .hasLabel('<label>'), e.g. .hasLabel('user'), to the query. The query call returns all vertices with the given label.
	HasLabel(vertexLabel string) Vertex
	// Property adds .property('<key>','<value>'), e.g. .property('name','hans'), to the query. The query call will add the given property.
	Property(key, value string) Vertex
	// Properties adds .properties(), to the query. The query call returns all properties of the vertex.
	Properties() QueryBuilder
	// Has adds .has('<key>','<value>'), e.g. .has('name','hans'), to the query. The query call returns all vertices
	// with the property which is defined by the given key value pair.
	Has(key, value string) Vertex
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
}
