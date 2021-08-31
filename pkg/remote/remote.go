package remote

import "net/url"

type Remote interface {
	Host() string
	Name() string
	Owner() string
	PullURL() *url.URL
	PushURL() *url.URL
	Repo() string
}

type remote struct {
	host     string
	name     string
	owner    string
	fetchURL *url.URL
	pushURL  *url.URL
	repo     string
}

// Remotes gets the git remotes set for the current repo
// Sort them
// Used to determine baseRepo
func Remotes() ([]*Remote, error) {
	return nil, nil
}
