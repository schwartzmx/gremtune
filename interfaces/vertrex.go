package interfaces

type Vertex interface {
	QueryBuilder
	HasLabel(vertexLabel string) Vertex
	Property(key, value string) Vertex
	Properties() Vertex
	Has(key, value string) Vertex
	ValuesBy(label string) Vertex
	Values() Vertex
	ValueMap() Vertex
	Add(builder QueryBuilder) Vertex
}
