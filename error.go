package gremcos

import (
	"errors"
	"fmt"
	"net"
)

type ErrConnectivity struct {
	Wrapped error
}

func (e ErrConnectivity) Error() string {
	return fmt.Sprintf("connectivity error: %v", e.Wrapped)
}

var ErrNoConnection = fmt.Errorf("Can't write - no connection")

// IsNetworkErr determines whether the given error is related to any network issues (timeout, connectivity,..)
func IsNetworkErr(err error) bool {
	if errors.Is(err, ErrNoConnection) {
		return true
	}

	errConn := ErrConnectivity{}
	if errors.As(err, &errConn) {
		return true
	}

	if isNetError(err) {
		return true
	}

	return false
}

// isNetError checks if the given error is (or is a wrapped) net.Error
func isNetError(err error) bool {
	if err == nil {
		return false
	}

	// call unwrap and try to cast to net.Error as long as possible
	for {
		if _, ok := err.(net.Error); ok {
			return true
		}

		if err = errors.Unwrap(err); err == nil {
			return false
		}
	}
	return false
}
