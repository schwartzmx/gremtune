package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscape(t *testing.T) {
	// GIVEN
	value1 := "abcdefghijklmnopqrstufvxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	value2 := `^°!"§$%&/()=?*+'#~,.;:-_<>|@€²³¼½¬{[]}\`
	value3 := `^°!§%&/()=?*+#~,.;:-_<>|@€²³¼½¬{[]}` // no $, no ", no ' and no \

	// WHEN
	escaped1 := Escape(value1)
	escaped2 := Escape(value2)
	escaped3 := Escape(value3)

	// THEN
	assert.Equal(t, value1, escaped1)
	assert.Equal(t, "%5E%C2%B0%21%22%C2%A7%24%25%26%2F%28%29%3D%3F%2A%2B%27%23~%2C.%3B%3A-_%3C%3E%7C%40%E2%82%AC%C2%B2%C2%B3%C2%BC%C2%BD%C2%AC%7B%5B%5D%7D%5C", escaped2)
	assert.Equal(t, value3, escaped3)
}

func TestEscapeUnescape(t *testing.T) {
	// GIVEN
	value1 := "abcdefghijklmnopqrstufvxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	value2 := `^°!"§$%&/()=?*+'#~,.;:-_<>|@€²³¼½¬{[]}\`
	value3 := "Hello $'World"
	value4 := `^°!"§%&/()=?*+'#~,.;:-_<>|@€²³¼½¬{[]}` // no $ and no \

	// WHEN
	escaped1 := Escape(value1)
	escaped2 := Escape(value2)
	escaped3 := Escape(value3)
	escaped4 := Escape(value4)

	unescaped1 := UnEscape(escaped1)
	unescaped2 := UnEscape(escaped2)
	unescaped3 := UnEscape(escaped3)
	unescaped4 := UnEscape(escaped4)

	// THEN
	assert.Equal(t, value1, unescaped1)
	assert.Equal(t, value2, unescaped2)
	assert.Equal(t, value3, unescaped3)
	assert.Equal(t, value4, unescaped4)
}

func TestShouldEscape(t *testing.T) {
	// GIVEN
	value1 := "a$b"
	value2 := `a\b`
	value3 := `\n`
	value4 := `\s`
	value5 := `\b`
	value6 := `nothing to escape here +1-/()ß?`

	// WHEN
	shouldEscape1 := ShouldEscape(value1)
	shouldEscape2 := ShouldEscape(value2)
	shouldEscape3 := ShouldEscape(value3)
	shouldEscape4 := ShouldEscape(value4)
	shouldEscape5 := ShouldEscape(value5)
	shouldEscape6 := ShouldEscape(value6)

	// THEN
	assert.True(t, shouldEscape1)
	assert.True(t, shouldEscape2)
	assert.True(t, shouldEscape3)
	assert.True(t, shouldEscape4)
	assert.True(t, shouldEscape5)
	assert.False(t, shouldEscape6)
}
