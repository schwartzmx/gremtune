package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/supplyon/gremcos/interfaces"
)

type vertex struct {
	builders []interfaces.QueryBuilder
}

func (v *vertex) String() string {

	queryString := ""
	for _, queryBuilder := range v.builders {
		queryString += queryBuilder.String()
	}

	return queryString
}

func NewVertexG(g interfaces.Graph) interfaces.Vertex {
	queryBuilders := make([]interfaces.QueryBuilder, 0)
	queryBuilders = append(queryBuilders, g)

	return &vertex{
		builders: queryBuilders,
	}
}

func NewVertexE(e interfaces.Edge) interfaces.Vertex {
	queryBuilders := make([]interfaces.QueryBuilder, 0)
	queryBuilders = append(queryBuilders, e)

	return &vertex{
		builders: queryBuilders,
	}
}

// V adds .V()
func (v *vertex) V() interfaces.Vertex {
	v.Add(NewSimpleQB(".V()"))
	return v
}

// Out adds .out([<edge_1>,<edge_2>,..,<edge_n>]), to the query. The query call returns all outgoing vertices of the named edge
func (v *vertex) Out(edgenames ...string) interfaces.Vertex {
	query := multiParamQuery(".out", edgenames...)
	v.Add(query)
	return v
}

// In adds .in([<edge_1>,<edge_2>,..,<edge_n>]), to the query. The query call returns all incoming vertices of the named edge
func (v *vertex) In(edgenames ...string) interfaces.Vertex {
	query := multiParamQuery(".in", edgenames...)
	v.Add(query)
	return v
}

// ByV adds .by([<traversal>]) to the query.
func (v *vertex) By(traversals ...interfaces.QueryBuilder) interfaces.Vertex {
	query := multitraversalQuery(".by", traversals...)
	return v.Add(query)
}

// Project adds .project([<label_1>,<label_2>,..,<label_n>])
func (v *vertex) Project(labels ...string) interfaces.Vertex {
	query := multiParamQuery(".project", labels...)
	return v.Add(query)
}

// Dedup adds .dedup() to the query.
func (v *vertex) Dedup() interfaces.Vertex {
	return v.Add(NewSimpleQB(".dedup()"))
}

// Order adds .order(), to the query.
func (v *vertex) Order() interfaces.Vertex {
	return v.Add(NewSimpleQB(".order()"))
}

// ByOrder adds .by('<name of the property>',[<sort-order>]), to the query.
// Sort order is ascending per default.
func (v *vertex) ByOrder(propertyName string, order ...interfaces.Order) interfaces.Vertex {
	if len(order) == 0 {
		return v.Add(NewSimpleQB(`.by("%s",%s)`, propertyName, toSortOrder(gUSE_COSMOS_DB_QUERY_LANGUAGE, interfaces.OrderAscending)))
	}

	return v.Add(NewSimpleQB(`.by("%s",%s)`, propertyName, toSortOrder(gUSE_COSMOS_DB_QUERY_LANGUAGE, order[0])))
}

// toSortOrder returns the sort order respecting the language differences between cosmos and tinkerpop gremlin dialect
func toSortOrder(useCosmosDialect bool, order ...interfaces.Order) string {
	sortOrder := interfaces.OrderAscending
	if len(order) > 0 {
		sortOrder = order[0]
	}

	if !useCosmosDialect {
		return sortOrder.String()
	}

	switch sortOrder {
	case interfaces.OrderAscending:
		return "incr"
	case interfaces.OrderDescending:
		return "decr"
	default:
		return "unknown"
	}
}

// Limit adds .limit(<num>), to the query. The query call will limit the results of the query to the given number.
func (v *vertex) Limit(maxElements int) interfaces.Vertex {
	return v.Add(NewSimpleQB(".limit(%d)", maxElements))
}

// As adds .as([<label_1>,<label_2>,..,<label_n>]), to the query to label that query step for later access.
func (v *vertex) As(labels ...string) interfaces.Vertex {
	query := multiParamQuery(".as", labels...)
	return v.Add(query)
}

// Aggregate adds .aggregate(<label>) step to the query. This is used to aggregate all the objects at a particular point of traversal into a Collection.
func (v *vertex) Aggregate(label string) interfaces.Vertex {
	return v.Add(NewSimpleQB(".aggregate(\"%s\")", label))
}

// Select adds .select([<label_1>,<label_2>,..,<label_n>]), to the query to select previous results using their label
func (v *vertex) Select(labels ...string) interfaces.Vertex {
	query := multiParamQuery(".select", labels...)
	return v.Add(query)
}

// AddV adds .addV("<label>"), e.g. .addV("user")
func (v *vertex) AddV(label string) interfaces.Vertex {
	return v.Add(NewSimpleQB(".addV(\"%s\")", label))
}

// Add can be used to add a custom QueryBuilder
// e.g. g.V().Add(NewSimpleQB(".myCustomCall("%s")",label))
func (v *vertex) Add(builder interfaces.QueryBuilder) interfaces.Vertex {
	v.builders = append(v.builders, builder)
	return v
}

// Coalesce adds .coalesce(<traversal>,<traversal>) to the query.
func (v *vertex) Coalesce(qb1 interfaces.QueryBuilder, qb2 interfaces.QueryBuilder) interfaces.Vertex {
	return v.Add(NewSimpleQB(".coalesce(%s,%s)", qb1, qb2))
}

// HasNext adds .hasNext() to the query. This part is commonly used to check for element existence (see: https://tinkerpop.apache.org/docs/current/recipes/#element-existence)
func (v *vertex) HasNext() interfaces.Vertex {
	return v.Add(NewSimpleQB(".hasNext()"))
}

// Unfold adds .unfold() to the query. An iterator, iterable, or map, then it is unrolled into a linear form. If not, then the object is simply emitted.
func (v *vertex) Unfold() interfaces.Vertex {
	return v.Add(NewSimpleQB(".unfold()"))
}

// Fold adds .fold() to the query.
func (v *vertex) Fold() interfaces.Vertex {
	return v.Add(NewSimpleQB(".fold()"))
}

// Or adds .or(<traversal_1>, <traversal_2>,...,<traversal_n>) to the query.
func (v *vertex) Or(traversals ...interfaces.QueryBuilder) interfaces.Vertex {
	query := multitraversalQuery(".or", traversals...)
	return v.Add(query)
}

// And adds .and(<traversal_1>, <traversal_2>,...,<traversal_n>) to the query.
func (v *vertex) And(traversals ...interfaces.QueryBuilder) interfaces.Vertex {
	query := multitraversalQuery(".and", traversals...)
	return v.Add(query)
}

//  Not adds .not(<traversal>) to the query.
func (v *vertex) Not(not interfaces.QueryBuilder) interfaces.Vertex {
	return v.Add(NewSimpleQB(".not(%s)", not))
}

// Where adds .where(<traversal>) to the query. The query call can be user to filter the results of a traversal
func (v *vertex) Where(where interfaces.QueryBuilder) interfaces.Vertex {
	return v.Add(NewSimpleQB(".where(%s)", where))
}

// Has adds .has("<key>","<value>"), e.g. .has("name","hans") depending on the given type the quotes for the value are omitted.
// e.g. .has("temperature",23.02) or .has("available",true)
// The method can also be used to return vertices that have a certain property.
// Then .has("<prop name>") will be added to the query.
//	v.Has("prop1")
func (v *vertex) Has(key string, value ...interface{}) interfaces.Vertex {
	if len(value) == 0 {
		return v.Add(NewSimpleQB(".has(\"%s\")", key))
	}

	keyVal, err := toKeyValueString(key, value[0])
	if err != nil {
		panic(errors.Wrapf(err, "cast has value %T to string failed (You could either implement the Stringer interface for this type or cast it to string beforehand)", value))
	}

	return v.Add(NewSimpleQB(".has%s", keyVal))
}

// HasLabel adds .hasLabel([<label_1>,<label_2>,..,<label_n>]), e.g. .hasLabel('user','name'), to the query. The query call returns all vertices with the given label.
func (v *vertex) HasLabel(vertexLabel ...string) interfaces.Vertex {
	query := multiParamQuery(".hasLabel", vertexLabel...)
	return v.Add(query)
}

// ValuesBy adds .values("<label>"), e.g. .values("user")
func (v *vertex) ValuesBy(label string) interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".values(\"%s\")", label))
}

// Values adds .values()
func (v *vertex) Values() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".values()"))
}

// ValueMap adds .valueMap()
func (v *vertex) ValueMap() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".valueMap()"))
}

// Properties adds .properties() or .properties("<prop1 name>","<prop2 name>",...)
func (v *vertex) Properties(keys ...string) interfaces.Property {

	query := NewSimpleQB(".properties()")
	if len(keys) > 0 {
		quotedKeys := make([]string, 0, len(keys))
		for _, key := range keys {
			quotedKeys = append(quotedKeys, fmt.Sprintf(`"%s"`, key))
		}
		keyList := strings.Join(quotedKeys, `,`)

		query = NewSimpleQB(".properties(%s)", keyList)
	}

	v.Add(query)
	return NewPropertyV(v)
}

// Id adds .id()
func (v *vertex) Id() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".id()"))
}

// Drop adds .drop(), to the query. The query call will drop/ delete all referenced entities
func (v *vertex) Drop() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".drop()"))
}

// AddE adds .addE(<label>), to the query. The query call will be the first step to add an edge
func (v *vertex) AddE(label string) interfaces.Edge {
	v.Add(NewSimpleQB(".addE(\"%s\")", label))
	return NewEdgeV(v)
}

// BothE adds .bothE(), to the query. The query call returns all edges of the Vertex
func (v *vertex) BothE() interfaces.Edge {
	v.Add(NewSimpleQB(".bothE()"))
	return NewEdgeV(v)
}

func (v *vertex) Profile() interfaces.QueryBuilder {
	if !gUSE_COSMOS_DB_QUERY_LANGUAGE {
		return v.Add(NewSimpleQB(".profile()"))
	}
	return v.Add(NewSimpleQB(".executionProfile()"))
}

// HasId adds .hasId('<id>'), e.g. .hasId('8aaaa410-dae1-4f33-8dd7-0217e69df10c'), to the query. The query call returns all vertices
// with the given id.
func (v *vertex) HasId(id string) interfaces.Vertex {
	return v.Add(NewSimpleQB(".hasId(\"%s\")", id))
}

// OutE adds .outE([<label_1>,<label_2>,..,<label_n>]), to the query. The query call returns all outgoing edges of the Vertex
func (v *vertex) OutE(labels ...string) interfaces.Edge {
	query := multiParamQuery(".outE", labels...)
	v.Add(query)
	return NewEdgeV(v)
}

// InE adds .inE([<label_1>,<label_2>,..,<label_n>]), to the query. The query call returns all incoming edges of the Vertex
func (v *vertex) InE(labels ...string) interfaces.Edge {
	query := multiParamQuery(".inE", labels...)
	v.Add(query)
	return NewEdgeV(v)
}

// Count adds .count(), to the query. The query call will return the number of entities found in the query.
func (v *vertex) Count() interfaces.QueryBuilder {
	return v.Add(NewSimpleQB(".count()"))
}

// PropertyList adds .property(list,"<key>","<value>"), e.g. .property(list, "name","hans"), to the query. The query call will add the given property.
func (v *vertex) PropertyList(key, value string) interfaces.Vertex {
	return v.Add(NewSimpleQB(".property(list,\"%s\",\"%s\")", key, Escape(value)))
}

// Property adds .property("<key>","<value>"), e.g. .property("name","hans") depending on the given type the quotes for the value are omitted.
// e.g. .property("temperature",23.02) or .property("available",true)
func (v *vertex) Property(key, value interface{}) interfaces.Vertex {
	keyVal, err := toKeyValueString(key, value)
	if err != nil {
		panic(errors.Wrapf(err, "cast property value %T to string failed (You could either implement the Stringer interface for this type or cast it to string beforehand)", value))
	}

	return v.Add(NewSimpleQB(".property%s", keyVal))
}

// toKeyValueString creates a string based on the given key and value as a key/value pair using the following format
//	(\"key\",\"value\")
// Depending on the given type of the value the quotes for the value are omitted.
// e.g. ("temperature",23.02) or ("available",true)
func toKeyValueString(key, value interface{}) (string, error) {
	switch casted := value.(type) {
	case string:
		return fmt.Sprintf("(\"%s\",\"%s\")", key, Escape(casted)), nil
	case bool:
		return fmt.Sprintf("(\"%s\",%t)", key, casted), nil
	case int, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("(\"%s\",%d)", key, casted), nil
	case float64:
		return fmt.Sprintf("(\"%s\",%f)", key, casted), nil
	case time.Time:
		return fmt.Sprintf("(\"%s\",\"%s\")", key, casted.String()), nil
	case fmt.Stringer:
		return fmt.Sprintf("(\"%s\",\"%s\")", key, Escape(casted.String())), nil
	default:
		fmt.Printf("[warn] Type %T is not supported in v.toKeyValueString() will try to cast to string\n", casted)
		asStr, err := cast.ToStringE(casted)
		if err != nil {
			return "", errors.Wrapf(err, "cast %T to string failed (You could either implement the Stringer interface for this type or cast it to string beforehand)", casted)
		}
		return fmt.Sprintf("(\"%s\",\"%s\")", key, Escape(asStr)), nil
	}
}
