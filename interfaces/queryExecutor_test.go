package interfaces

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmpty(t *testing.T) {
	t.Parallel()
	// GIVEN
	response1 := Response{}
	response2 := Response{Result: Result{Data: []byte("")}}
	response3 := Response{Result: Result{Data: []byte("null")}}
	response4 := Response{Result: Result{Data: []byte("some data")}}

	// WHEN
	res1 := response1.IsEmpty()
	res2 := response2.IsEmpty()
	res3 := response3.IsEmpty()
	res4 := response4.IsEmpty()

	// THEN
	assert.True(t, res1)
	assert.True(t, res2)
	assert.True(t, res3)
	assert.False(t, res4)
}
