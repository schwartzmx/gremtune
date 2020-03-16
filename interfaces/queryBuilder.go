package interfaces

type QueryBuilder interface {
	String() string
}

type Graph interface {
	QueryBuilder
	V() Vertex
	VBy(id int) Vertex
	AddV(label string) Vertex
}

type Vertex interface {
	QueryBuilder
	HasLabel(vertexLabel string) Vertex
	Property(key, value string) Vertex
	Properties() QueryBuilder
	Has(key, value string) Vertex
	ValuesBy(label string) QueryBuilder
	Values() QueryBuilder
	ValueMap() QueryBuilder
	Add(builder QueryBuilder) Vertex
	Id() QueryBuilder
}
