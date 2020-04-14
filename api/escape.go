package api

import (
	"net/url"
	"regexp"
)

// Escape escapes all values that are not allowed to be stored directly into the cosmos-db.
func Escape(value string) string {
	if !ShouldEscape(value) {
		return value
	}

	// just use url query escaping to ensure that neither
	// $, \ or control characters lead to failing queries.
	return url.QueryEscape(value)
}

// UnEscape reverses potentially applied escaping. In case of an error the original input value will be returned
func UnEscape(value string) string {
	result, err := url.QueryUnescape(value)
	if err != nil {
		return value
	}
	return result
}

var regexpSpecialChars = regexp.MustCompile(`(\$|\\n|\\v|\\r|\\t|\\f|\\s|\\b|'|"|\\)`)

// ShouldEscape returns true in case the given string needs to be escaped.
func ShouldEscape(value string) bool {
	return regexpSpecialChars.MatchString(value)
}
