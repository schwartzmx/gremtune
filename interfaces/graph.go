package interfaces

type Graph interface {
	QueryBuilder
	V() Vertex
	VBy(id int) Vertex
	AddV(label string) Vertex
}
