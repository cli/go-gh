// Package gh is a library for CLI Go applications to help interface with the gh CLI tool,
// and the GitHub API.
//
// Note that the examples in this package assume gh and git are installed. They do not run in
// the Go Playground used by pkg.go.dev.
package gh

import (
	"bytes"
	"fmt"
	"os/exec"

	iapi "github.com/cli/go-gh/internal/api"
	"github.com/cli/go-gh/internal/config"
	"github.com/cli/go-gh/internal/git"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/safeexec"
)

// Exec gh command with provided arguments.
func Exec(args ...string) (stdOut, stdErr bytes.Buffer, err error) {
	path, err := path()
	if err != nil {
		err = fmt.Errorf("could not find gh executable in PATH. error: %w", err)
		return
	}
	return run(path, nil, args...)
}

func path() (string, error) {
	return safeexec.LookPath("gh")
}

func run(path string, env []string, args ...string) (stdOut, stdErr bytes.Buffer, err error) {
	cmd := exec.Command(path, args...)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	if env != nil {
		cmd.Env = env
	}
	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("failed to run gh: %s. error: %w", stdErr.String(), err)
		return
	}
	return
}

// RESTClient builds a client to send requests to GitHub REST API endpoints.
// As part of the configuration a hostname, auth token, and default set of headers are resolved
// from the gh environment configuration. These behaviors can be overridden using the opts argument.
func RESTClient(opts *api.ClientOptions) (api.RESTClient, error) {
	var cfg config.Config
	var token string
	var err error
	if opts == nil {
		opts = &api.ClientOptions{}
	}
	if opts.Host == "" || opts.AuthToken == "" {
		cfg, err = config.Load()
		if err != nil {
			return nil, err
		}
	}
	if opts.Host == "" {
		opts.Host = cfg.Host()
	}
	if opts.AuthToken == "" {
		token, err = cfg.AuthToken(opts.Host)
		if err != nil {
			return nil, err
		}
		opts.AuthToken = token
	}
	return iapi.NewRESTClient(opts.Host, opts), nil
}

// GQLClient builds a client to send requests to GitHub GraphQL API endpoints.
// As part of the configuration a hostname, auth token, and default set of headers are resolved
// from the gh environment configuration. These behaviors can be overridden using the opts argument.
func GQLClient(opts *api.ClientOptions) (api.GQLClient, error) {
	var cfg config.Config
	var token string
	var err error
	if opts == nil {
		opts = &api.ClientOptions{}
	}
	if opts.Host == "" || opts.AuthToken == "" {
		cfg, err = config.Load()
		if err != nil {
			return nil, err
		}
	}
	if opts.Host == "" {
		opts.Host = cfg.Host()
	}
	if opts.AuthToken == "" {
		token, err = cfg.AuthToken(opts.Host)
		if err != nil {
			return nil, err
		}
		opts.AuthToken = token
	}
	return iapi.NewGQLClient(opts.Host, opts), nil
}

// CurrentRepository uses git remotes to determine the GitHub repository
// the current directory is tracking.
func CurrentRepository() (Repository, error) {
	remotes, err := git.Remotes()
	if err != nil {
		return nil, err
	}
	if len(remotes) == 0 {
		return nil, fmt.Errorf("unable to determine current repository")
	}
	r := remotes[0]
	return repo{host: r.Host, name: r.Repo, owner: r.Owner}, nil
}

// Repository is the interface that wraps repository information methods.
type Repository interface {
	Host() string
	Name() string
	Owner() string
}

type repo struct {
	host  string
	name  string
	owner string
}

func (r repo) Host() string {
	return r.host
}

func (r repo) Name() string {
	return r.name
}

func (r repo) Owner() string {
	return r.owner
}
