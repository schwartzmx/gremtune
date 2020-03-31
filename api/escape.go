package api

import (
	"net/url"
)

// Escape escapes all values that are not allowed to be stored directly into the cosmos-db.
func Escape(value string) string {
	// just use url query escaping to ensure that neither
	// $, \ or control characters lead to failing queries.
	return url.QueryEscape(value)
}

// UnEscape reverses potentially applied escaping
func UnEscape(value string) (string, error) {
	return url.QueryUnescape(value)
}
