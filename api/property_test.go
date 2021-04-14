package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPropertyV(t *testing.T) {
	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := NewVertexG(g)
	require.NotNil(t, v)

	// WHEN
	p := NewPropertyV(v)

	// THEN
	assert.NotNil(t, p)
	assert.Equal(t, graphName, p.String())
}

func TestAddP(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := NewVertexG(g)
	require.NotNil(t, v)
	p := NewPropertyV(v)
	require.NotNil(t, p)

	// WHEN
	edge := p.Add(NewSimpleQB(".test()"))

	// THEN
	assert.NotNil(t, edge)
	assert.Equal(t, fmt.Sprintf("%s.test()", graphName), p.String())
}

func TestPropertyDrop(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := NewVertexG(g)
	require.NotNil(t, v)
	p := NewPropertyV(v)
	require.NotNil(t, p)

	// WHEN
	qb := p.Drop()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.drop()", graphName), p.String())
}

func TestPropertyProfile(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := NewVertexG(g)
	require.NotNil(t, v)
	p := NewPropertyV(v)
	require.NotNil(t, p)

	// WHEN
	qb := p.Profile()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.executionProfile()", graphName), p.String())
}

func TestPropertyProfile_GremlinDialect(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := NewVertexG(g)
	require.NotNil(t, v)
	p := NewPropertyV(v)
	require.NotNil(t, p)

	// WHEN
	SetQueryLanguageTo(QueryLanguageTinkerpopGremlin)
	qb := p.Profile()
	SetQueryLanguageTo(QueryLanguageCosmosDB)

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.profile()", graphName), p.String())
}

func TestPropertyCount(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := NewVertexG(g)
	require.NotNil(t, v)
	p := NewPropertyV(v)
	require.NotNil(t, p)

	// WHEN
	qb := p.Count()

	// THEN
	assert.NotNil(t, qb)
	assert.Equal(t, fmt.Sprintf("%s.count()", graphName), qb.String())
}

func TestPropertyLimit(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	require.NotNil(t, g)
	v := NewVertexG(g)
	require.NotNil(t, v)
	p := NewPropertyV(v)
	require.NotNil(t, p)

	limit := 1234

	// WHEN
	edge := p.Limit(limit)

	// THEN
	assert.NotNil(t, edge)
	assert.Equal(t, fmt.Sprintf(`%s.limit(%d)`, graphName, limit), p.String())
}
