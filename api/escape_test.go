package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscape(t *testing.T) {
	// GIVEN
	value1 := "abcdefghijklmnopqrstufvxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	value2 := `^°!"§$%&/()=?*+'#~,.;:-_<>|@€²³¼½¬{[]}\`

	// WHEN
	escaped1 := Escape(value1)
	escaped2 := Escape(value2)

	// THEN
	assert.Equal(t, value1, escaped1)
	assert.Equal(t, "%5E%C2%B0%21%22%C2%A7%24%25%26%2F%28%29%3D%3F%2A%2B%27%23~%2C.%3B%3A-_%3C%3E%7C%40%E2%82%AC%C2%B2%C2%B3%C2%BC%C2%BD%C2%AC%7B%5B%5D%7D%5C", escaped2)
}

func TestEscapeUnescape(t *testing.T) {
	// GIVEN
	value1 := "abcdefghijklmnopqrstufvxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	value2 := `^°!"§$%&/()=?*+'#~,.;:-_<>|@€²³¼½¬{[]}\`
	value3 := "Hello $'World"

	// WHEN
	escaped1 := Escape(value1)
	escaped2 := Escape(value2)
	escaped3 := Escape(value3)

	unescaped1, err := UnEscape(escaped1)
	assert.NoError(t, err)
	unescaped2, err := UnEscape(escaped2)
	assert.NoError(t, err)
	unescaped3, err := UnEscape(escaped3)
	assert.NoError(t, err)

	// THEN
	assert.Equal(t, value1, unescaped1)
	assert.Equal(t, value2, unescaped2)
	assert.Equal(t, value3, unescaped3)
}
