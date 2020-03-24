package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	assert.Equal(t, fmt.Sprintf("%s.V().has('%s','%s')", graphName, key, value), v.String())
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
	v = v.HasInt(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().has('%s',%d)", graphName, key, value), v.String())
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
	v = v.PropertyInt(key, value)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V().property('%s',%d)", graphName, key, value), v.String())
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
	assert.Equal(t, fmt.Sprintf("%s.V().hasLabel('%s')", graphName, label), v.String())
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
	assert.Equal(t, fmt.Sprintf("%s.V().values('%s')", graphName, label), qb.String())
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

func TestProperty(t *testing.T) {
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
	assert.Equal(t, fmt.Sprintf("%s.V().property('%s','%s')", graphName, key, value), v.String())
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
	assert.Equal(t, fmt.Sprintf("%s.V().addE('%s')", graphName, label), qb.String())
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
	assert.Equal(t, fmt.Sprintf("%s.addV('%s').property('%s','%s').property('%s','%s').properties()", graphName, vertrexlabel, key1, value1, key2, value2), qb.String())
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
	assert.Equal(t, fmt.Sprintf("%s.property(list,'%s','%s')", graphName, key, value), v.String())
}
