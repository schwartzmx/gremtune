package interfaces

type Graph interface {
	V() Vertex
	VBy(id int) Vertex
}
