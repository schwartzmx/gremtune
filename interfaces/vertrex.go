package interfaces

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
