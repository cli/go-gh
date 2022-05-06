package errors

import "errors"

// NoRepositoriesError is returned when go-gh is unable to determine current repository,
// when no git remotes configured for this repository.
type NoRepositoriesError struct {
	error
}

// NoRepositoryHostsError is returned when go-gh is unable to determine current repository,
// when none of the git remotes configured for this repository point to a known GitHub host.
type NoRepositoryHostsError struct {
	error
}

// NotFoundError is returned when there is no authentication token for the host.
type NotFoundError struct {
	error
}

var (
	ErrNoRepositories    = NoRepositoriesError{errors.New("unable to determine current repository, no git remotes configured for this repository")}
	ErrNoRepositoryHosts = NoRepositoryHostsError{errors.New("unable to determine current repository, none of the git remotes configured for this repository point to a known GitHub host")}
	ErrNotFound          = NotFoundError{errors.New("not found")}
)
