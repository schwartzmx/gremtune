package api

import (
	"net/url"
	"regexp"
	"strings"
)

// matches all relevant control chars
var controlCharsRegexp = regexp.MustCompile("\n|\t|\f|\r|\v")

// matches $ that is not already escaped via \
var unescapedDollarRegexp = regexp.MustCompile(`[^\\]\$`)

// matches single \ that are not pre-/ succeeded by another \
var singleBackslashRegexp = regexp.MustCompile(`([^\\])\\([^\\\$])`)

// matches single \ that is at the start of the string and not succeeded by another backslash
var singleBackslashAtStartOfStringRegexp = regexp.MustCompile(`^\\([^\\\$])`)

// matches single \ that is at the end of the string and not preceded by another backslash
var singleBackslashAtEndOfStringRegexp = regexp.MustCompile(`([^\\])\\$`)

// matches single \ that is the only character in the string
var singleBackslashAsSoleCharRegexp = regexp.MustCompile(`^\\$`)

func shouldEncode(value string) bool {
	// check if value contains any control chars
	if controlCharsRegexp.MatchString(value) {
		return true
	}

	// check if value contains a dollar char that is not already escaped
	if unescapedDollarRegexp.MatchString(value) {
		return true
	}

	// check if value contains a dollar char that is not already escaped
	if strings.HasPrefix(value, "$") {
		return true
	}

	// check if there is a any unescaped backslash at the beginning of the string
	if singleBackslashAtStartOfStringRegexp.MatchString(value) {
		return true
	}

	// check if there is a any unescaped backslash at the end of the string
	if singleBackslashAtEndOfStringRegexp.MatchString(value) {
		return true
	}

	// check if there is a single \ that is the only character in the string
	if singleBackslashAsSoleCharRegexp.MatchString(value) {
		return true
	}

	// check if there is a any unescaped backslash
	if singleBackslashRegexp.MatchString(value) {
		return true
	}
	return false
}

func Escape(value string) string {
	//if !shouldEncode(value) {
	//	return value
	//}
	return url.QueryEscape(value)
}

//func UnEncode(value string) string {
//
//	return url.QueryUnescape(value)
//}

func EscapeGroovyChars(value string) string {

	// TODO: use urlencode

	return "value"
}

func UnEscapeGroovyChars(value string) string {
	// remove control chars
	return ""
}
