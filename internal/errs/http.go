package errs

// HTTPError is a custom error for HTTP responses
type HTTPError struct {
	error // the original error
	code    int // the HTTP status code
	message string // user-friendly error message
}

/* NewHTTPError creates a new HTTPError

src: is the original error.
code: is the HTTP status code.
message: is the user-friendly error message.
*/
func NewHTTPError(src error, code int, message string) error {
	return HTTPError{error: src, code: code, message: message}
}

func (e HTTPError) Code() int {
	return e.code
}

func (e HTTPError) Message() string {
	return e.message
}

func (e HTTPError) Error() string {
	return e.message
}

func (e HTTPError) Unwrap() error {
	return e.error
}
