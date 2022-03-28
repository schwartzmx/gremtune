package interfaces

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderString(t *testing.T) {
	assert.Equal(t, "asc", OrderAscending.String())
	assert.Equal(t, "desc", OrderDescending.String())
}
