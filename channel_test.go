package gremcos

import (
	"testing"
)

func TestNewCloseOnceChannel(t *testing.T) {
	// GIVEN
	channel := make(chan error)
	// WHEN
	NewCloseOnceChannel(channel)
	// THEN
}
