package interfaces

import "github.com/gofrs/uuid"

type Order string

const (
	OrderAscending  Order = "asc"
	OrderDescending Order = "desc"
)

func (order Order) String() string {
	if order == OrderAscending {
		return "asc"
	}
	return "desc"
}

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

	// HasLabel adds .hasLabel([<label_1>,<label_2>,..,<label_n>]), e.g. .hasLabel('user','name'), to the query. The query call returns all vertices with the given label.
	HasLabel(vertexLabel ...string) Vertex

	// Property adds .property("<key>","<value>"), e.g. .property("name","hans") depending on the given type the quotes for the value are omitted.
	// e.g. .property("temperature",23.02) or .property("available",true)
	Property(key, value interface{}) Vertex

	// PropertyList adds .property(list,'<key>','<value>'), e.g. .property(list, 'name','hans'), to the query. The query call will add the given property.
	PropertyList(key, value string) Vertex

	// Properties adds .properties(), to the query. The query call returns all properties of the vertex.
	// The method can also be used to return only specific properties identified by their name.
	// Then .properties("<prop1 name>","<prop2 name>",...) will be added to the query.
	//	v.Properties("prop1","prop2")
	Properties(key ...string) Property

	// Has adds .has("<key>","<value>"), e.g. .has("name","hans") depending on the given type the quotes for the value are omitted.
	// e.g. .has("temperature",23.02) or .has("available",true)
	// The method can also be used to return vertices that have a certain property.
	// Then .has("<prop name>") will be added to the query.
	//	v.Has("prop1")
	Has(key string, value ...interface{}) Vertex

	// HasId adds .hasId('<id>'), e.g. .hasId('8aaaa410-dae1-4f33-8dd7-0217e69df10c'), to the query. The query call returns all vertices
	// with the given id.
	HasId(id string) Vertex

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

	// OutE adds .outE([<label_1>,<label_2>,..,<label_n>]), to the query. The query call returns all outgoing edges of the Vertex
	OutE(labels ...string) Edge

	// InE adds .inE([<label_1>,<label_2>,..,<label_n>]), to the query. The query call returns all incoming edges of the Vertex
	InE(labels ...string) Edge

	// Out adds .out([<edge_1>,<edge_2>,..,<edge_n>]), to the query. The query call returns all outgoing vertices of the named edge
	Out(edgenames ...string) Vertex

	// In adds .in([<edge_1>,<edge_2>,..,<edge_n>]), to the query. The query call returns all incoming vertices of the named edge
	In(edgenames ...string) Vertex

	// Limit adds .limit(<num>), to the query. The query call will limit the results of the query to the given number.
	Limit(maxElements int) Vertex

	// As adds .as([<label_1>,<label_2>,..,<label_n>]), to the query to label that query step for later access.
	As(labels ...string) Vertex

	// Aggregate adds .aggregate(<label>) step to the query. This is used to aggregate all the objects at a particular point of traversal into a Collection.
	Aggregate(label string) Vertex

	// Select adds .select([<label_1>,<label_2>,..,<label_n>]), to the query to select previous results using their label
	Select(labels ...string) Vertex

	// Not adds .not(<traversal>) to the query.
	Not(builder QueryBuilder) Vertex

	// Or adds .or(<traversal_1>, <traversal_2>,...,<traversal_n>) to the query.
	Or(builder ...QueryBuilder) Vertex

	// And adds .and(<traversal_1>, <traversal_2>,...,<traversal_n>) to the query.
	And(builder ...QueryBuilder) Vertex

	// Fold adds .fold() to the query.
	Fold() Vertex

	// Coalesce adds .coalesce(<traversal>,<traversal>) to the query.
	Coalesce(qb1 QueryBuilder, qb2 QueryBuilder) Vertex

	// AddV adds .addV('<label>'), e.g. .addV('user'), to the query. The query call adds a vertex with the given label and returns that vertex.
	AddV(label string) Vertex

	// V adds .V() to the query. The query call returns all vertices.
	V() Vertex

	// Where adds .where(<traversal>) to the query. The query call can be user to filter the results of a traversal
	Where(qb QueryBuilder) Vertex

	// HasNext adds .hasNext() to the query. This part is commonly used to check for element existence (see: https://tinkerpop.apache.org/docs/current/recipes/#element-existence)
	HasNext() Vertex

	// Unfold adds .unfold() to the query. An iterator, iterable, or map, then it is unrolled into a linear form. If not, then the object is simply emitted.
	Unfold() Vertex

	// BothE adds .bothE(), to the query. The query call returns all edges of the Vertex
	BothE() Edge

	// Order adds .order(), to the query.
	Order() Vertex

	// ByOrder adds .by('<name of the property>',[<sort-order>]), to the query.
	// Sort order is ascending per default.
	ByOrder(propertyName string, order ...Order) Vertex

	// By adds .by(<traversal_1>, <traversal_2>,...,<traversal_n>) to the query.
	By(builder ...QueryBuilder) Vertex

	// Project adds .project([<label_1>,<label_2>,..,<label_n>])
	Project(labels ...string) Vertex

	// Dedup adds .dedup() to the query.
	Dedup() Vertex
}

type Edge interface {
	QueryBuilder
	Dropper
	Profiler
	Counter

	// Property adds .property("<key>","<value>"), e.g. .property("name","hans") depending on the given type the quotes for the value are omitted.
	// e.g. .property("temperature",23.02) or .property("available",true)
	Property(key, value interface{}) Edge

	// To adds .to(<vertex>), to the query. The query call will be the second step to add an edge
	To(v Vertex) Edge
	// From adds .from(<vertex>), to the query. The query call will be the second step to add an edge
	From(v Vertex) Edge

	// ToLbl adds .to(<label>), to the query. The query call will be the second step to add an edge
	ToLbl(label string) Edge
	// From adds .from(<label>), to the query. The query call will be the second step to add an edge
	FromLbl(label string) Edge

	// OutV adds .outV(), to the query. The query call will return the vertices on the outgoing side of this edge
	OutV() Vertex

	// InV adds .inV(), to the query. The query call will return the vertices on the incoming side of this edge
	InV() Vertex

	// Add can be used to add a custom QueryBuilder
	// e.g. g.V().Add(NewSimpleQB(".myCustomCall('%s')",label))
	Add(builder QueryBuilder) Edge

	// HasLabel adds .hasLabel([<label_1>,<label_2>,..,<label_n>]), e.g. .hasLabel('user','name'), to the query. The query call returns all edges with the given label.
	HasLabel(label ...string) Edge

	// Id adds .id(), to the query. The query call returns the id of the edge.
	Id() QueryBuilder

	// HasId adds .hasId('<id>'), e.g. .hasId('8aaaa410-dae1-4f33-8dd7-0217e69df10c'), to the query. The query call returns all edges
	// with the given id.
	HasId(id string) Edge

	// Limit adds .limit(<num>), to the query. The query call will limit the results of the query to the given number.
	Limit(maxElements int) Edge

	// As adds .as([<label_1>,<label_2>,..,<label_n>]), to the query to label that query step for later access.
	As(labels ...string) Edge

	// Aggregate adds .aggregate(<label>) step to the query. This is used to aggregate all the objects at a particular point of traversal into a Collection.
	Aggregate(label string) Edge

	// Select adds .select([<label_1>,<label_2>,..,<label_n>]), to the query to select previous results using their label
	Select(labels ...string) Vertex

	// Not adds .not(<traversal>) to the query.
	Not(builder QueryBuilder) Edge

	// Or adds .or(<traversal_1>, <traversal_2>,...,<traversal_n>) to the query.
	Or(builder ...QueryBuilder) Edge

	// And adds .and(<traversal_1>, <traversal_2>,...,<traversal_n>) to the query.
	And(builder ...QueryBuilder) Edge

	// Fold adds .fold() to the query.
	Fold() Edge

	// Coalesce adds .coalesce(<traversal>,<traversal>) to the query.
	Coalesce(qb1 QueryBuilder, qb2 QueryBuilder) Edge

	// Where adds .where(<traversal>) to the query. The query call can be user to filter the results of a traversal
	Where(qb QueryBuilder) Edge

	// HasNext adds .hasNext() to the query. This part is commonly used to check for element existence (see: https://tinkerpop.apache.org/docs/current/recipes/#element-existence)
	HasNext() Edge

	// Unfold adds .unfold() to the query. An iterator, iterable, or map, then it is unrolled into a linear form. If not, then the object is simply emitted.
	Unfold() Edge

	// Order adds .order(), to the query.
	Order() Edge

	// ByOrder adds .by('<name of the property>',[<sort-order>]), to the query.
	// Sort order is ascending per default.
	ByOrder(propertyName string, order ...Order) Edge

	// By adds .by(<traversal_1>, <traversal_2>,...,<traversal_n>) to the query.
	By(builder ...QueryBuilder) Edge

	// Project adds .project([<label_1>,<label_2>,..,<label_n>])
	Project(labels ...string) Edge

	// Dedup adds .dedup() to the query.
	Dedup() Edge
}

type Property interface {
	QueryBuilder
	Dropper
	Profiler
	Counter

	// Add can be used to add a custom QueryBuilder
	// e.g. g.V().properties("prop1").Add(NewSimpleQB(".myCustomCall('%s')",label))
	Add(builder QueryBuilder) Property

	// Limit adds .limit(<num>), to the query. The query call will limit the results of the query to the given number.
	Limit(maxElements int) Property

	// As adds .as([<label_1>,<label_2>,..,<label_n>]), to the query to label that query step for later access.
	As(labels ...string) Property
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
