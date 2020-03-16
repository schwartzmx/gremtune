package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGraph(t *testing.T) {
	// GIVEN
	graphName := "mygraph"

	// WHEN
	g := NewGraph(graphName)

	// THEN
	assert.NotNil(t, g)
	assert.Equal(t, graphName, g.String())
}

func TestV(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)

	// WHEN
	v := g.V()

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V()", graphName), v.String())
}

func TestVBy(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	id := 1

	// WHEN
	v := g.VBy(id)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V('%d')", graphName, id), v.String())
}

func TestAddV(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	label := "user"

	// WHEN
	v := g.AddV(label)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.addV('%s')", graphName, label), v.String())
}
