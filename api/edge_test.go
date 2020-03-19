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
