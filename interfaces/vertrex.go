package interfaces

type Vertex interface {
	HasLabel(vertexLabel string) Vertex
}
