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
	assert.Equal(t, "%5E%C2%B0%21%22%C2%A7$%25&%2F%28%29=%3F%2A+%27%23~%2C.%3B:-_%3C%3E%7C@%E2%82%AC%C2%B2%C2%B3%C2%BC%C2%BD%C2%AC%7B%5B%5D%7D%5C", escaped2)
	assert.Equal(t, value3, escaped3)
}

func TestEscapeUnescape(t *testing.T) {
	// GIVEN
	value1 := "abcdefghijklmnopqrstufvxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	value2 := `^°!"§$%&/()=?*+'#~,.;:-_<>|@€²³¼½¬{[]}\`
	value3 := "Hello $'World"
	value4 := `^°!"§%&/()=?*+'#~,.;:-_<>|@€²³¼½¬{[]}` // no $ and no \
	value5 := `invalid escape sequence %&`
	value6 := `2020-03-19 15:34:17.8242487 +0000 UTC`
	value7 := `2020-03-19 15:34:17.8242487 +0000 UTC`

	// WHEN
	escaped1 := Escape(value1)
	escaped2 := Escape(value2)
	escaped3 := Escape(value3)
	escaped4 := Escape(value4)
	escaped5 := Escape(value5)
	escaped6 := Escape(value6)

	unescaped1 := UnEscape(escaped1)
	unescaped2 := UnEscape(escaped2)
	unescaped3 := UnEscape(escaped3)
	unescaped4 := UnEscape(escaped4)
	unescaped5 := UnEscape(escaped5)
	unescaped6 := UnEscape(escaped6)
	unescaped7 := UnEscape(value7)

	// THEN
	assert.Equal(t, value1, unescaped1)
	assert.Equal(t, value2, unescaped2)
	assert.Equal(t, value3, unescaped3)
	assert.Equal(t, value4, unescaped4)
	assert.Equal(t, value5, unescaped5)
	assert.Equal(t, value6, unescaped6)
	assert.Equal(t, value7, unescaped7)
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

func TestShouldUnEscape(t *testing.T) {
	// GIVEN
	value1 := "%AD"
	value2 := `%ad`
	value3 := `%99`
	value4 := `%EF`
	value5 := `abc%23de`
	value6 := `%NO`
	value7 := `2020-03-19 15:34:17.8242487 +0000 UTC`

	// WHEN
	shouldUnEscape1 := ShouldUnescape(value1)
	shouldUnEscape2 := ShouldUnescape(value2)
	shouldUnEscape3 := ShouldUnescape(value3)
	shouldUnEscape4 := ShouldUnescape(value4)
	shouldUnEscape5 := ShouldUnescape(value5)
	shouldUnEscape6 := ShouldUnescape(value6)
	shouldUnEscape7 := ShouldUnescape(value7)

	// THEN
	assert.True(t, shouldUnEscape1)
	assert.True(t, shouldUnEscape2)
	assert.True(t, shouldUnEscape3)
	assert.True(t, shouldUnEscape4)
	assert.True(t, shouldUnEscape5)
	assert.False(t, shouldUnEscape6)
	assert.False(t, shouldUnEscape7)
}
