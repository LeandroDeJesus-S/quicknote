package errs

type RepoError struct {
	error
}

func NewRepoError(err error) error {
	return RepoError{err}
}
