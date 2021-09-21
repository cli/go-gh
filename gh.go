package gh

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/cli/go-gh/internal/config"
	"github.com/cli/go-gh/internal/git"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/safeexec"
)

// Execute gh command with provided arguments.
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

func DefaultRESTClient(opts api.ClientOptions) (api.RESTClient, error) {
	var cfg config.Config
	var token string
	var err error
	if opts.Host == "" {
		cfg, err = config.Load()
		if err != nil {
			return nil, err
		}
		opts.Host = cfg.Host()
	}
	if opts.AuthToken == "" {
		if cfg == nil {
			cfg, err = config.Load()
			if err != nil {
				return nil, err
			}
		}
		token, err = cfg.Token(opts.Host)
		if err != nil {
			return nil, err
		}
		opts.AuthToken = token
	}
	return api.NewRESTClient(opts.Host, opts), nil
}

func DefaultGQLClient(opts api.ClientOptions) (api.GQLClient, error) {
	var cfg config.Config
	var token string
	var err error
	if opts.Host == "" {
		cfg, err = config.Load()
		if err != nil {
			return nil, err
		}
		opts.Host = cfg.Host()
	}
	if opts.AuthToken == "" {
		if cfg == nil {
			cfg, err = config.Load()
			if err != nil {
				return nil, err
			}
		}
		token, err = cfg.Token(opts.Host)
		if err != nil {
			return nil, err
		}
		opts.AuthToken = token
	}
	return api.NewGQLClient(opts.Host, opts), nil
}

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
