package api

import (
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
	v := NewVertex(g)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, graphName, v.String())
}
