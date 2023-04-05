package gh

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/config"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GH_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if err := func(args []string) error {
		if args[len(args)-1] == "error" {
			return fmt.Errorf("process exited with error")
		}
		fmt.Fprintf(os.Stdout, "%v", args)
		return nil
	}(os.Args[3:]); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func TestHelperProcessLongRunning(t *testing.T) {
	if os.Getenv("GH_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args[3:]
	fmt.Fprintf(os.Stdout, "%v", args)
	fmt.Fprint(os.Stderr, "going to sleep...")
	time.Sleep(10 * time.Second)
	fmt.Fprint(os.Stderr, "...going to exit")
	os.Exit(0)
}

func TestRun(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run(context.TODO(), os.Args[0], []string{"GH_WANT_HELPER_PROCESS=1"}, nil, &stdout, &stderr,
		[]string{"-test.run=TestHelperProcess", "--", "gh", "issue", "list"})
	assert.NoError(t, err)
	assert.Equal(t, "[gh issue list]", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestRunError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run(context.TODO(), os.Args[0], []string{"GH_WANT_HELPER_PROCESS=1"}, nil, &stdout, &stderr,
		[]string{"-test.run=TestHelperProcess", "--", "gh", "error"})
	assert.EqualError(t, err, "gh execution failed: exit status 1")
	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "process exited with error", stderr.String())
}

func TestRunInteractiveContextCanceled(t *testing.T) {
	// pass current time to ensure that deadline has already passed
	ctx, cancel := context.WithDeadline(context.Background(), time.Now())
	cancel()
	err := run(ctx, os.Args[0], []string{"GH_WANT_HELPER_PROCESS=1"}, nil, nil, nil,
		[]string{"-test.run=TestHelperProcessLongRunning", "--", "gh", "issue", "list"})
	assert.EqualError(t, err, "gh execution failed: context deadline exceeded")
}

func TestRESTClient(t *testing.T) {
	stubConfig(t, testConfig())
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Get("/some/test/path").
		MatchHeader("Authorization", "token abc123").
		Reply(200).
		JSON(`{"message": "success"}`)

	client, err := RESTClient(nil)
	assert.NoError(t, err)

	res := struct{ Message string }{}
	err = client.Do("GET", "some/test/path", nil, &res)
	assert.NoError(t, err)
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
	assert.Equal(t, "success", res.Message)
}

func TestGQLClient(t *testing.T) {
	stubConfig(t, testConfig())
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Post("/graphql").
		MatchHeader("Authorization", "token abc123").
		BodyString(`{"query":"QUERY","variables":{"var":"test"}}`).
		Reply(200).
		JSON(`{"data":{"viewer":{"login":"hubot"}}}`)

	client, err := GQLClient(nil)
	assert.NoError(t, err)

	vars := map[string]interface{}{"var": "test"}
	res := struct{ Viewer struct{ Login string } }{}
	err = client.Do("QUERY", vars, &res)
	assert.NoError(t, err)
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
	assert.Equal(t, "hubot", res.Viewer.Login)
}

func TestGQLClientError(t *testing.T) {
	stubConfig(t, testConfig())
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Post("/graphql").
		MatchHeader("Authorization", "token abc123").
		BodyString(`{"query":"QUERY","variables":null}`).
		Reply(200).
		JSON(`{"errors":[{"type":"NOT_FOUND","path":["organization"],"message":"Could not resolve to an Organization with the login of 'cli'."}]}`)

	client, err := GQLClient(nil)
	assert.NoError(t, err)

	res := struct{ Organization struct{ Name string } }{}
	err = client.Do("QUERY", nil, &res)
	assert.EqualError(t, err, "GraphQL: Could not resolve to an Organization with the login of 'cli'. (organization)")
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
}

func TestHTTPClient(t *testing.T) {
	stubConfig(t, testConfig())
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Get("/some/test/path").
		MatchHeader("Authorization", "token abc123").
		Reply(200).
		JSON(`{"message": "success"}`)

	client, err := HTTPClient(nil)
	assert.NoError(t, err)

	res, err := client.Get("https://api.github.com/some/test/path")
	assert.NoError(t, err)
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
	assert.Equal(t, 200, res.StatusCode)
}

func TestResolveOptions(t *testing.T) {
	stubConfig(t, testConfigWithSocket())

	tests := []struct {
		name          string
		opts          *api.ClientOptions
		wantAuthToken string
		wantHost      string
		wantSocket    string
	}{
		{
			name: "honors consumer provided ClientOptions",
			opts: &api.ClientOptions{
				Host:             "test.com",
				AuthToken:        "token_from_opts",
				UnixDomainSocket: "socket_from_opts",
			},
			wantAuthToken: "token_from_opts",
			wantHost:      "test.com",
			wantSocket:    "socket_from_opts",
		},
		{
			name:          "uses config values if there are no consumer provided ClientOptions",
			opts:          &api.ClientOptions{},
			wantAuthToken: "token",
			wantHost:      "github.com",
			wantSocket:    "socket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := resolveOptions(tt.opts)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantHost, tt.opts.Host)
			assert.Equal(t, tt.wantAuthToken, tt.opts.AuthToken)
			assert.Equal(t, tt.wantSocket, tt.opts.UnixDomainSocket)
		})
	}
}

func TestOptionsNeedResolution(t *testing.T) {
	tests := []struct {
		name string
		opts *api.ClientOptions
		out  bool
	}{
		{
			name: "Host, AuthToken, and UnixDomainSocket specified",
			opts: &api.ClientOptions{
				Host:             "test.com",
				AuthToken:        "token",
				UnixDomainSocket: "socket",
			},
			out: false,
		},
		{
			name: "Host, AuthToken, and Transport specified",
			opts: &api.ClientOptions{
				Host:      "test.com",
				AuthToken: "token",
				Transport: http.DefaultTransport,
			},
			out: false,
		},
		{
			name: "Host, and AuthToken specified",
			opts: &api.ClientOptions{
				Host:      "test.com",
				AuthToken: "token",
			},
			out: true,
		},
		{
			name: "Host, and UnixDomainSocket specified",
			opts: &api.ClientOptions{
				Host:             "test.com",
				UnixDomainSocket: "socket",
			},
			out: true,
		},
		{
			name: "Host, and Transport specified",
			opts: &api.ClientOptions{
				Host:      "test.com",
				Transport: http.DefaultTransport,
			},
			out: true,
		},
		{
			name: "AuthToken, and UnixDomainSocket specified",
			opts: &api.ClientOptions{
				AuthToken:        "token",
				UnixDomainSocket: "socket",
			},
			out: true,
		},
		{
			name: "AuthToken, and Transport specified",
			opts: &api.ClientOptions{
				AuthToken: "token",
				Transport: http.DefaultTransport,
			},
			out: true,
		},
		{
			name: "Host specified",
			opts: &api.ClientOptions{
				Host: "test.com",
			},
			out: true,
		},
		{
			name: "AuthToken specified",
			opts: &api.ClientOptions{
				AuthToken: "token",
			},
			out: true,
		},
		{
			name: "UnixDomainSocket specified",
			opts: &api.ClientOptions{
				UnixDomainSocket: "socket",
			},
			out: true,
		},
		{
			name: "Transport specified",
			opts: &api.ClientOptions{
				Transport: http.DefaultTransport,
			},
			out: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.out, optionsNeedResolution(tt.opts))
		})
	}
}

func printPendingMocks(mocks []gock.Mock) string {
	paths := []string{}
	for _, mock := range mocks {
		paths = append(paths, mock.Request().URLStruct.String())
	}
	return fmt.Sprintf("%d unmatched mocks: %s", len(paths), strings.Join(paths, ", "))
}

func stubConfig(t *testing.T, cfgStr string) {
	t.Helper()
	old := config.Read
	config.Read = func() (*config.Config, error) {
		return config.ReadFromString(cfgStr), nil
	}
	t.Cleanup(func() {
		config.Read = old
	})
}

func testConfig() string {
	return `
hosts:
  github.com:
    user: user1
    oauth_token: abc123
    git_protocol: ssh
`
}

func testConfigWithSocket() string {
	return `
http_unix_socket: socket
hosts:
  github.com:
    user: user1
    oauth_token: token
    git_protocol: ssh
`
}
