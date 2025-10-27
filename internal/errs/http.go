// Package errs provides custom error types.
package errs

// HTTPError represents an error with an associated HTTP status code.
type HTTPError struct {
	error   // The original error.
	code    int    // The HTTP status code.
	message string // A user-friendly error message.
}

// NewHTTPError creates a new HTTPError.
//
// src is the original error.
// code is the HTTP status code.
// message is the user-friendly error message.
func NewHTTPError(src error, code int, message string) error {
	return HTTPError{error: src, code: code, message: message}
}

// Code returns the HTTP status code of the error.
func (e HTTPError) Code() int {
	return e.code
}

// Message returns the user-friendly error message.
func (e HTTPError) Message() string {
	return e.message
}

// Error returns the error message.
func (e HTTPError) Error() string {
	return e.message
}

// Unwrap returns the original error.
func (e HTTPError) Unwrap() error {
	return e.error
}
