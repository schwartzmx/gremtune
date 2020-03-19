package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVertex(t *testing.T) {
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
