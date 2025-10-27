// Package errs provides custom error types.
package errs

// RepoError is a custom error for repository-related errors.
type RepoError struct {
	error
}

// NewRepoError creates a new RepoError.
func NewRepoError(err error) error {
	return RepoError{err}
}
