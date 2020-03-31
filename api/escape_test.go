package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldEncodeTrue(t *testing.T) {
	// GIVEN
	value1 := "\n"
	value2 := "a\na"
	value3 := "a\n"
	value4 := "\na"
	value5 := "$"
	value6 := "a$a"
	value7 := "a$"
	value8 := "$a"
	value9 := `\`
	value10 := `a\a`
	value11 := `a\`
	value12 := `\a`
	value13 := "\t"
	value14 := "\f"
	value15 := "\r"
	value16 := "\v"

	// WHEN
	should1 := shouldEncode(value1)
	should2 := shouldEncode(value2)
	should3 := shouldEncode(value3)
	should4 := shouldEncode(value4)
	should5 := shouldEncode(value5)
	should6 := shouldEncode(value6)
	should7 := shouldEncode(value7)
	should8 := shouldEncode(value8)
	should9 := shouldEncode(value9)
	should10 := shouldEncode(value10)
	should11 := shouldEncode(value11)
	should12 := shouldEncode(value12)
	should13 := shouldEncode(value13)
	should14 := shouldEncode(value14)
	should15 := shouldEncode(value15)
	should16 := shouldEncode(value16)

	// THEN
	assert.True(t, should1)
	assert.True(t, should2)
	assert.True(t, should3)
	assert.True(t, should4)
	assert.True(t, should5)
	assert.True(t, should6)
	assert.True(t, should7)
	assert.True(t, should8)
	assert.True(t, should9)
	assert.True(t, should10)
	assert.True(t, should11)
	assert.True(t, should12)
	assert.True(t, should13)
	assert.True(t, should14)
	assert.True(t, should15)
	assert.True(t, should16)
}

func TestShouldEncodeFalse(t *testing.T) {
	// GIVEN
	value1 := `\$`
	value2 := `a\$a`
	value3 := `\$a`
	value4 := `a\$`
	value5 := `\\`
	value6 := `a\\a`
	value7 := `a\\`
	value8 := `\\a`
	value9 := `\\`

	// WHEN
	should1 := shouldEncode(value1)
	should2 := shouldEncode(value2)
	should3 := shouldEncode(value3)
	should4 := shouldEncode(value4)
	should5 := shouldEncode(value5)
	should6 := shouldEncode(value6)
	should7 := shouldEncode(value7)
	should8 := shouldEncode(value8)
	should9 := shouldEncode(value9)

	// THEN
	assert.False(t, should1)
	assert.False(t, should2)
	assert.False(t, should3)
	assert.False(t, should4)
	assert.False(t, should5)
	assert.False(t, should6)
	assert.False(t, should7)
	assert.False(t, should8)
	assert.False(t, should9)
}

func TestEscape(t *testing.T) {
	// GIVEN
	value1 := "A\n\t\f\r\vB"
	value2 := "A\\B"
	value3 := "A$B"
	//	value4 := "A\\$B"

	// WHEN
	escaped1 := Escape(value1)
	escaped2 := Escape(value2)
	escaped3 := Escape(value3)
	//	escaped4 := Escape(value4)

	// THEN
	assert.Equal(t, "A%0A%09%0C%0D%0BB", escaped1)
	assert.Equal(t, "A%5CB", escaped2)
	assert.Equal(t, "A%24B", escaped3)
	//	assert.Equal(t, "A\\$B", escaped4)
}

func TestUnEscape(t *testing.T) {
	// GIVEN
	value1 := "A\n\t\f\r\vB"
	value2 := "A\\B"
	value3 := "A$B"
	//	value4 := "A\\$B"

	// WHEN
	escaped1 := Escape(value1)
	escaped2 := Escape(value2)
	escaped3 := Escape(value3)
	//	escaped4 := Escape(value4)

	// THEN
	assert.Equal(t, "A%0A%09%0C%0D%0BB", escaped1)
	assert.Equal(t, "A%5CB", escaped2)
	assert.Equal(t, "A%24B", escaped3)
	//	assert.Equal(t, "A\\$B", escaped4)
}

func TestEscapeDollar(t *testing.T) {
	// GIVEN
	value1 := "$"
	value2 := "\\$"

	// WHEN
	escaped1 := EscapeGroovyChars(value1)
	escaped2 := EscapeGroovyChars(value2)

	// THEN
	assert.Equal(t, "\\$", escaped1)
	assert.Equal(t, "\\$", escaped2)
}

func TestUnEscapeDollar(t *testing.T) {
	// GIVEN
	value := "A$B\\$C\\"
	escaped := EscapeGroovyChars(value)
	// WHEN
	unescaped := UnEscapeGroovyChars(escaped)
	fmt.Printf("%s -> %s -> %s\n", value, escaped, unescaped)

	// THEN
	assert.Equal(t, value, unescaped)
	assert.Fail(t, "failureMessage")
}
