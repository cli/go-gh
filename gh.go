// Package gh is a library for CLI Go applications to help interface with the gh CLI tool,
// and the GitHub API.
//
// Note that the examples in this package assume gh and git are installed. They do not run in
// the Go Playground used by pkg.go.dev.
package gh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	iapi "github.com/cli/go-gh/internal/api"
	"github.com/cli/go-gh/internal/git"
	irepo "github.com/cli/go-gh/internal/repository"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/auth"
	"github.com/cli/go-gh/pkg/config"
	repo "github.com/cli/go-gh/pkg/repository"
	"github.com/cli/go-gh/pkg/ssh"
	"github.com/cli/safeexec"
)

// Exec invokes a gh command in a subprocess and captures the output and error streams.
func Exec(args ...string) (stdout, stderr bytes.Buffer, err error) {
	ghExe, err := ghLookPath()
	if err != nil {
		return
	}
	err = run(context.Background(), ghExe, nil, nil, &stdout, &stderr, args)
	return
}

// ExecContext invokes a gh command in a subprocess and captures the output and error streams.
func ExecContext(ctx context.Context, args ...string) (stdout, stderr bytes.Buffer, err error) {
	ghExe, err := ghLookPath()
	if err != nil {
		return
	}
	err = run(ctx, ghExe, nil, nil, &stdout, &stderr, args)
	return
}

// Exec invokes a gh command in a subprocess with its stdin, stdout, and stderr streams connected to
// those of the parent process. This is suitable for running gh commands with interactive prompts.
func ExecInteractive(ctx context.Context, args ...string) error {
	ghExe, err := ghLookPath()
	if err != nil {
		return err
	}
	return run(ctx, ghExe, nil, os.Stdin, os.Stdout, os.Stderr, args)
}

func ghLookPath() (string, error) {
	if ghExe := os.Getenv("GH_PATH"); ghExe != "" {
		return ghExe, nil
	}
	return safeexec.LookPath("gh")
}

func run(ctx context.Context, ghExe string, env []string, stdin io.Reader, stdout, stderr io.Writer, args []string) error {
	cmd := exec.CommandContext(ctx, ghExe, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if env != nil {
		cmd.Env = env
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh execution failed: %w", err)
	}
	return nil
}

// RESTClient builds a client to send requests to GitHub REST API endpoints.
// As part of the configuration a hostname, auth token, default set of headers,
// and unix domain socket are resolved from the gh environment configuration.
// These behaviors can be overridden using the opts argument.
func RESTClient(opts *api.ClientOptions) (api.RESTClient, error) {
	if opts == nil {
		opts = &api.ClientOptions{}
	}
	if optionsNeedResolution(opts) {
		err := resolveOptions(opts)
		if err != nil {
			return nil, err
		}
	}
	return iapi.NewRESTClient(opts.Host, opts), nil
}

// GQLClient builds a client to send requests to GitHub GraphQL API endpoints.
// As part of the configuration a hostname, auth token, default set of headers,
// and unix domain socket are resolved from the gh environment configuration.
// These behaviors can be overridden using the opts argument.
func GQLClient(opts *api.ClientOptions) (api.GQLClient, error) {
	if opts == nil {
		opts = &api.ClientOptions{}
	}
	if optionsNeedResolution(opts) {
		err := resolveOptions(opts)
		if err != nil {
			return nil, err
		}
	}
	return iapi.NewGQLClient(opts.Host, opts), nil
}

// HTTPClient builds a client that can be passed to another library.
// As part of the configuration a hostname, auth token, default set of headers,
// and unix domain socket are resolved from the gh environment configuration.
// These behaviors can be overridden using the opts argument. In this instance
// providing opts.Host will not change the destination of your request as it is
// the responsibility of the consumer to configure this. However, if opts.Host
// does not match the request host, the auth token will not be added to the headers.
// This is to protect against the case where tokens could be sent to an arbitrary
// host.
func HTTPClient(opts *api.ClientOptions) (*http.Client, error) {
	if opts == nil {
		opts = &api.ClientOptions{}
	}
	if optionsNeedResolution(opts) {
		err := resolveOptions(opts)
		if err != nil {
			return nil, err
		}
	}
	client := iapi.NewHTTPClient(opts)
	return &client, nil
}


func optionsNeedResolution(opts *api.ClientOptions) bool {
	if opts.Host == "" {
		return true
	}
	if opts.AuthToken == "" {
		return true
	}
	if opts.UnixDomainSocket == "" && opts.Transport == nil {
		return true
	}
	return false
}

func resolveOptions(opts *api.ClientOptions) error {
	cfg, _ := config.Read()
	if opts.Host == "" {
		opts.Host, _ = auth.DefaultHost()
	}
	if opts.AuthToken == "" {
		opts.AuthToken, _ = auth.TokenForHost(opts.Host)
		if opts.AuthToken == "" {
			return fmt.Errorf("authentication token not found for host %s", opts.Host)
		}
	}
	if opts.UnixDomainSocket == "" && cfg != nil {
		opts.UnixDomainSocket, _ = cfg.Get([]string{"http_unix_socket"})
	}
	return nil
}
