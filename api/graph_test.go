package api

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, fmt.Sprintf("%s.V(\"%d\")", graphName, id), v.String())
}

func TestVByUUID(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	id, err := uuid.NewV4()
	require.NoError(t, err)

	// WHEN
	v := g.VByUUID(id)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V(\"%s\")", graphName, id), v.String())
}

func TestVByStr(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)
	id := "1234ABCD"

	// WHEN
	v := g.VByStr(id)

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.V(\"%s\")", graphName, id), v.String())
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
	assert.Equal(t, fmt.Sprintf("%s.addV(\"%s\")", graphName, label), v.String())
}

func TestE(t *testing.T) {

	// GIVEN
	graphName := "mygraph"
	g := NewGraph(graphName)

	// WHEN
	v := g.E()

	// THEN
	assert.NotNil(t, v)
	assert.Equal(t, fmt.Sprintf("%s.E()", graphName), v.String())
}
