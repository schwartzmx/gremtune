package interfaces

type Vertex interface {
	QueryBuilder
	HasLabel(vertexLabel string) Vertex
}
