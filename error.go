package gremcos

import (
	"errors"
	"fmt"
	"net"

	"github.com/supplyon/gremcos/interfaces"
)

type ErrorCategory string

const (
	ErrorCategoryGeneral      ErrorCategory = "GeneralErr"
	ErrorCategoryConnectivity ErrorCategory = "ConnectivityErr"
	ErrorCategoryAuth         ErrorCategory = "AuthErr"
	ErrorCategoryClient       ErrorCategory = "ClientErr"
	ErrorCategoryServer       ErrorCategory = "ServerErr"
)

type Error struct {
	Wrapped  error
	Category ErrorCategory
}

func (e Error) Error() string {
	return fmt.Sprintf("[%s] %v", e.Category, e.Wrapped)
}

var ErrNoConnection = Error{Wrapped: fmt.Errorf("no connection"), Category: ErrorCategoryConnectivity}

// IsNetworkErr determines whether the given error is related to any network issues (timeout, connectivity,..)
func IsNetworkErr(err error) bool {
	if errors.Is(err, ErrNoConnection) {
		return true
	}

	errConn := Error{}
	if errors.As(err, &errConn) {
		if errConn.Category != ErrorCategoryConnectivity {
			return false
		}
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
}

// DetectError detects any possible errors in responses from Gremlin Server and generates an error for each code
func extractError(r interfaces.Response) error {
	switch r.Status.Code {
	case interfaces.StatusSuccess, interfaces.StatusNoContent, interfaces.StatusPartialContent:
		return nil
	case interfaces.StatusUnauthorized:
		return Error{Wrapped: fmt.Errorf("unauthorized: %s", r.Status.Message), Category: ErrorCategoryAuth}
	case interfaces.StatusAuthenticate:
		return Error{Wrapped: fmt.Errorf("not authenticated: %s", r.Status.Message), Category: ErrorCategoryAuth}
	case interfaces.StatusMalformedRequest:
		return Error{Wrapped: fmt.Errorf("malformed request: %s", r.Status.Message), Category: ErrorCategoryClient}
	case interfaces.StatusInvalidRequestArguments:
		return Error{Wrapped: fmt.Errorf("invalid request arguments: %s", r.Status.Message), Category: ErrorCategoryClient}
	case interfaces.StatusServerError:
		return Error{Wrapped: fmt.Errorf("server error: %s", r.Status.Message), Category: ErrorCategoryServer}
	case interfaces.StatusScriptEvaluationError:
		return Error{Wrapped: fmt.Errorf("script evaluation failed: %s", r.Status.Message), Category: ErrorCategoryClient}
	case interfaces.StatusServerTimeout:
		return Error{Wrapped: fmt.Errorf("server timeout: %s", r.Status.Message), Category: ErrorCategoryServer}
	case interfaces.StatusServerSerializationError:
		return Error{Wrapped: fmt.Errorf("script evaluation failed: %s", r.Status.Message), Category: ErrorCategoryClient}
	default:
		return Error{Wrapped: fmt.Errorf("unknown error: %s", r.Status.Message), Category: ErrorCategoryGeneral}
	}
}
