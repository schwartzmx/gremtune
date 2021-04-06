package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
