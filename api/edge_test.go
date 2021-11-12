package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoalesceE(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	q1 := NewSimpleQB("g.V()")
	q2 := NewSimpleQB("g.V().count()")

	// WHEN
	result := e.Coalesce(q1, q2)

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.coalesce(%s,%s)", graphName, q1, q2), result.String())
}

func TestHasNextE(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	// WHEN
	result := e.HasNext()

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.hasNext()", graphName), result.String())
}

func TestFoldE(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	// WHEN
	result := e.Fold()

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.fold()", graphName), result.String())
}

func TestSelectE(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	label1 := "l1"
	label2 := "l2"

	// WHEN
	result := e.Select(label1, label2)

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.select(\"%s\",\"%s\")", graphName, label1, label2), result.String())
}

func TestNotE(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	q1 := NewSimpleQB("g.V()")

	// WHEN
	result := e.Not(q1)

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.not(%s)", graphName, q1), result.String())
}

func TestWhereE(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	q1 := NewSimpleQB("g.V()")

	// WHEN
	result := e.Where(q1)

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.where(%s)", graphName, q1), result.String())
}

func TestToLblE(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	label := "l1"

	// WHEN
	result := e.ToLbl(label)

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.to(\"%s\")", graphName, label), result.String())
}

func TestFromLblE(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	label := "l1"

	// WHEN
	result := e.FromLbl(label)

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.from(\"%s\")", graphName, label), result.String())
}

func TestPropertyE(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	key := "temperature"
	value := 23.4

	// WHEN
	result := e.Property(key, value)

	// THEN
	assert.NotNil(t, result)
	assert.Equal(t, fmt.Sprintf("%s.property(\"%s\",%f)", graphName, key, value), result.String())
}
func TestNewEdgeG(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)

	// WHEN
	e := NewEdgeG(g)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, graphName, e.String())
}

func TestNewEdgeV(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := NewVertexG(g)
	require.NotNil(t, v)

	// WHEN
	e := NewEdgeV(v)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, graphName, e.String())
}

func TestAdd(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)

	// WHEN
	edge := e.Add(NewSimpleQB(".test()"))

	// THEN
	assert.NotNil(t, edge)
	assert.Equal(t, fmt.Sprintf("%s.test()", graphName), e.String())
}

func TestTo(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)

	// WHEN
	edge := e.To(g.V())

	// THEN
	assert.NotNil(t, edge)
	assert.Equal(t, fmt.Sprintf("%s.to(%s.V())", graphName, graphName), e.String())
}

func TestFrom(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)

	// WHEN
	edge := e.From(g.V())

	// THEN
	assert.NotNil(t, edge)
	assert.Equal(t, fmt.Sprintf("%s.from(%s.V())", graphName, graphName), e.String())
}

func TestEdgeDrop(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)

	// WHEN
	qb := e.Drop()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.drop()", graphName), e.String())
}

func TestEdgeProfile(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)

	// WHEN
	qb := e.Profile()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.executionProfile()", graphName), e.String())
}

func TestEdgeProfile_GremlinDialect(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)

	// WHEN
	SetQueryLanguageTo(QueryLanguageTinkerpopGremlin)
	qb := e.Profile()
	SetQueryLanguageTo(QueryLanguageCosmosDB)

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.profile()", graphName), e.String())
}

func TestInV(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)

	// WHEN
	v := e.InV()

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.inV()", graphName), e.String())
}

func TestOutV(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)

	// WHEN
	v := e.OutV()

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.outV()", graphName), e.String())
}

func TestEdgeHasLabel(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	label := "label"

	// WHEN
	e = e.HasLabel(label)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, fmt.Sprintf("%s.hasLabel(\"%s\")", graphName, label), e.String())
}

func TestEdgeHasLabelMulti(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	l1 := "label1"
	l2 := "label2"

	// WHEN
	e = e.HasLabel(l1, l2)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, fmt.Sprintf("%s.hasLabel(\"%s\",\"%s\")", graphName, l1, l2), e.String())
}

func TestEdgeCount(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)

	// WHEN
	qb := e.Count()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.count()", graphName), qb.String())
}

func TestEdgeHasId(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	id := "my-id"

	// WHEN
	e = e.HasId(id)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, fmt.Sprintf("%s.hasId(\"%s\")", graphName, id), e.String())
}

func TestEdgeId(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)

	// WHEN
	qb := e.Id()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.id()", graphName), qb.String())
}

func TestEdgeLimit(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	limit := 1234

	// WHEN
	edge := e.Limit(limit)

	// THEN
	assert.NotNil(t, edge)
	assert.Equal(t, fmt.Sprintf(`%s.limit(%d)`, graphName, limit), e.String())
}

func TestEdgeAs(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	label := "label"

	// WHEN
	e = e.As(label)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, fmt.Sprintf("%s.as(\"%s\")", graphName, label), e.String())
}

func TestEdgeAsMulti(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	e := NewEdgeG(g)
	require.NotNil(t, e)
	l1 := "label1"
	l2 := "label2"

	// WHEN
	e = e.As(l1, l2)

	// THEN
	assert.NotNil(t, e)
	assert.Equal(t, fmt.Sprintf("%s.as(\"%s\",\"%s\")", graphName, l1, l2), e.String())
}
