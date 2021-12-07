package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVV(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)

	// WHEN
	result := v.HasLabel("Label").As("label").V()

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.V().hasLabel(\"Label\").as(\"label\").V()", graphName), result.String())
}

func TestCoalesceV(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	q1 := NewSimpleQB("g.V()")
	q2 := NewSimpleQB("g.V().count()")

	// WHEN
	result := v.Coalesce(q1, q2)

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.V().coalesce(%s,%s)", graphName, q1, q2), result.String())
}

func TestHasNextV(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)

	// WHEN
	result := v.HasNext()

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.V().hasNext()", graphName), result.String())
}

func TestFoldV(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)

	// WHEN
	result := v.Fold()

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.V().fold()", graphName), result.String())
}

func TestUnfoldV(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)

	// WHEN
	result := v.Unfold()

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.V().unfold()", graphName), result.String())
}

func TestSelectV(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	label1 := "l1"
	label2 := "l2"

	// WHEN
	result := v.Select(label1, label2)

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.V().select(\"%s\",\"%s\")", graphName, label1, label2), result.String())
}

func TestNotV(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	q1 := NewSimpleQB("g.V()")

	// WHEN
	result := v.Not(q1)

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.V().not(%s)", graphName, q1), result.String())
}

func TestWhereV(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	q1 := NewSimpleQB("g.V()")

	// WHEN
	result := v.Where(q1)

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.V().where(%s)", graphName, q1), result.String())
}

func TestNewVertexG(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)

	// WHEN
	v := NewVertexG(g)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, graphName, v.String())
}

func TestNewVertexE(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)

	// WHEN
	v := NewVertexE(e)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, graphName, v.String())
}

func TestHasCheck(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"

	// WHEN
	v = v.Has(key)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf(`%s.V().has("%s")`, graphName, key), v.String())
}

func TestHas(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"
	value := "value"

	// WHEN
	v = v.Has(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().has(\"%s\",\"%s\")", graphName, key, value), v.String())
}

func TestHasInt(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"
	value := 12345

	// WHEN
	v = v.Has(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().has(\"%s\",%d)", graphName, key, value), v.String())
}

func TestHasBool(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"
	value := true

	// WHEN
	v = v.Has(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().has(\"%s\",%t)", graphName, key, value), v.String())
}

func TestHasFloat(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"
	value := 12.34

	// WHEN
	v = v.Has(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().has(\"%s\",%f)", graphName, key, value), v.String())
}

func TestHasTime(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"
	value := time.Now()

	// WHEN
	v = v.Has(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().has(\"%s\",\"%s\")", graphName, key, value), v.String())
}

func TestHasMisc(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"
	value := myStructWithStringer{field1: "hello", field2: 12345}

	// WHEN
	v = v.Has(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().has(\"%s\",\"%s\")", graphName, key, value.String()), v.String())
}

func TestHasMiscFail(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"
	type myStruct struct {
		field1 string
		field2 int
	}
	value := myStruct{field1: "hello", field2: 12345}

	// WHEN + THEN
	assert.Panics(t, func() { v.Has(key, value) }, "The code did not panic")
}

func TestHasLabel(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	label := "label"

	// WHEN
	v = v.HasLabel(label)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().hasLabel(\"%s\")", graphName, label), v.String())
}

func TestHasLabelMulti(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	l1 := "label1"
	l2 := "label2"

	// WHEN
	v = v.HasLabel(l1, l2)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().hasLabel(\"%s\",\"%s\")", graphName, l1, l2), v.String())
}

func TestValuesBy(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	label := "label"

	// WHEN
	qb := v.ValuesBy(label)

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.V().values(\"%s\")", graphName, label), qb.String())
}

func TestValues(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)

	// WHEN
	qb := v.Values()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.V().values()", graphName), qb.String())
}

func TestValueMap(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)

	// WHEN
	qb := v.ValueMap()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.V().valueMap()", graphName), qb.String())
}

func TestProperties(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)

	// WHEN
	qb := v.Properties()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.V().properties()", graphName), qb.String())
}

func TestPropertiesWithKey(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)

	// WHEN
	qb := v.Properties("prop1", "prop2")

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf(`%s.V().properties("prop1","prop2")`, graphName), qb.String())
}

func TestPropertyStr(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"
	value := "value"

	// WHEN
	v = v.Property(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().property(\"%s\",\"%s\")", graphName, key, value), v.String())
}

func TestPropertyInt(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"
	value := 12345

	// WHEN
	v = v.Property(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().property(\"%s\",%d)", graphName, key, value), v.String())
}

func TestPropertyFloat(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"
	value := 23.02

	// WHEN
	v = v.Property(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().property(\"%s\",%f)", graphName, key, value), v.String())
}

func TestPropertyBool(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"
	value := true

	// WHEN
	v = v.Property(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().property(\"%s\",%t)", graphName, key, value), v.String())
}

func TestPropertyTime(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"
	value := time.Now()

	// WHEN
	v = v.Property(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().property(\"%s\",\"%s\")", graphName, key, value), v.String())
}

func TestPropertyMiscFail(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"
	type myStruct struct {
		field1 string
		field2 int
	}
	value := myStruct{field1: "hello", field2: 12345}

	// WHEN + THEN
	assert.Panics(t, func() { v.Property(key, value) }, "The code did not panic")
}

type myStructWithStringer struct {
	field1 string
	field2 int
}

func (ms myStructWithStringer) String() string {
	return fmt.Sprintf("%s,%d", ms.field1, ms.field2)
}

func TestPropertyMisc(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	key := "key"
	value := myStructWithStringer{field1: "hello", field2: 12345}

	// WHEN
	v = v.Property(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().property(\"%s\",\"%s\")", graphName, key, value.String()), v.String())
}

func TestId(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)

	// WHEN
	qb := v.Id()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.V().id()", graphName), qb.String())
}

func TestProfile(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)

	// WHEN
	qb := v.Profile()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.V().executionProfile()", graphName), qb.String())
}

func TestProfile_GremlinDialect(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)

	// WHEN
	SetQueryLanguageTo(QueryLanguageTinkerpopGremlin)
	qb := v.Profile()
	SetQueryLanguageTo(QueryLanguageCosmosDB)

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.V().profile()", graphName), qb.String())
}

func TestDrop(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)

	// WHEN
	qb := v.Drop()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.V().drop()", graphName), qb.String())
}

func TestAddE(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	label := "mylabel"

	// WHEN
	qb := v.AddE(label)

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.V().addE(\"%s\")", graphName, label), qb.String())
}

func TestChain(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	vertrexlabel := "vertrexlabel"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.AddV(vertrexlabel)
	require.NotNil(t, v)
	key1 := "key1"
	value1 := "value1"
	key2 := "key2"
	value2 := "value2"

	// WHEN
	qb := v.Property(key1, value1).Property(key2, value2).Properties()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.addV(\"%s\").property(\"%s\",\"%s\").property(\"%s\",\"%s\").properties()", graphName, vertrexlabel, key1, value1, key2, value2), qb.String())
}

func TestVertexCount(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := NewVertexG(g)
	require.NotNil(t, v)

	// WHEN
	qb := v.Count()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.count()", graphName), qb.String())
}

func TestOutE(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := NewVertexG(g)
	require.NotNil(t, v)

	// WHEN
	e := v.OutE()

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, fmt.Sprintf("%s.outE()", graphName), e.String())
}

func TestOutEMulti(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := NewVertexG(g)
	require.NotNil(t, v)
	l1 := "label1"
	l2 := "label2"

	// WHEN
	e := v.OutE(l1, l2)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, fmt.Sprintf("%s.outE(\"label1\",\"label2\")", graphName), e.String())
}

func TestInE(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := NewVertexG(g)
	require.NotNil(t, v)

	// WHEN
	e := v.InE()

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, fmt.Sprintf("%s.inE()", graphName), e.String())
}

func TestInEMulti(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := NewVertexG(g)
	require.NotNil(t, v)
	l1 := "label1"
	l2 := "label2"

	// WHEN
	e := v.InE(l1, l2)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, fmt.Sprintf("%s.inE(\"label1\",\"label2\")", graphName), e.String())
}

func TestPropertyList(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := NewVertexG(g)
	require.NotNil(t, v)
	key := "key"
	value := "value"

	// WHEN
	v = v.PropertyList(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.property(list,\"%s\",\"%s\")", graphName, key, value), v.String())
}

func TestHasId(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	id := "my-id"

	// WHEN
	v = v.HasId(id)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().hasId(\"%s\")", graphName, id), v.String())
}

func TestVertexLimit(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)

	limit := 1234

	// WHEN
	edge := v.Limit(limit)

	// THEN
	assert.NotNil(t, edge)
	assert.Equal(t, fmt.Sprintf(`%s.V().limit(%d)`, graphName, limit), v.String())
}

func TestVertexAs(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	label := "label1"

	// WHEN
	v = v.As(label)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().as(\"%s\")", graphName, label), v.String())
}

func TestVertexAggregate(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	label := "label1"

	// WHEN
	v = v.Aggregate(label)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().aggregate(\"%s\")", graphName, label), v.String())
}
func TestVertexAsMulti(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := g.V()
	require.NotNil(t, v)
	l1 := "label1"
	l2 := "label2"

	// WHEN
	v = v.As(l1, l2)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().as(\"%s\",\"%s\")", graphName, l1, l2), v.String())
}
