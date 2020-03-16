package interfaces

// QueryBuilder can be used to generate queries for the cosmos db
type QueryBuilder interface {
	String() string
}

// Graph represents a QueryBuilder that can be used to create
// queries on graph level
type Graph interface {
	QueryBuilder
	// V adds .V()
	V() Vertex
	// VBy adds .V(<id>), e.g. .V(123)
	VBy(id int) Vertex
	// AddV adds .addV('<label>'), e.g. .addV('user')
	AddV(label string) Vertex
}

// Vertex represents a QueryBuilder that can be used to create
// queries on vertex level
type Vertex interface {
	QueryBuilder
	// HasLabel adds .hasLabel('<label>'), e.g. .hasLabel('user')
	HasLabel(vertexLabel string) Vertex
	// Property adds .property('<key>','<value>'), e.g. .property('name','hans')
	Property(key, value string) Vertex
	// Properties adds .properties()
	Properties() QueryBuilder
	// Has adds .has('<key>','<value>'), e.g. .has('name','hans')
	Has(key, value string) Vertex
	// ValuesBy adds .values('<label>'), e.g. .values('user')
	ValuesBy(label string) QueryBuilder
	// Values adds .values()
	Values() QueryBuilder
	// ValueMap adds .valueMap()
	ValueMap() QueryBuilder
	// Add can be used to add a custom QueryBuilder
	// e.g. g.V().Add(NewSimpleQB(".myCustomCall('%s')",label))
	Add(builder QueryBuilder) Vertex
	// Id adds .id()
	Id() QueryBuilder
}
