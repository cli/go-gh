package gh

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cli/go-gh/internal/config"
	"github.com/cli/go-gh/pkg/api"
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

func TestRun(t *testing.T) {
	stdOut, stdErr, err := run(os.Args[0],
		[]string{"GH_WANT_HELPER_PROCESS=1"},
		"-test.run=TestHelperProcess", "--", "gh", "issue", "list")
	assert.NoError(t, err)
	assert.Equal(t, "[gh issue list]", stdOut.String())
	assert.Equal(t, "", stdErr.String())
}

func TestRunError(t *testing.T) {
	stdOut, stdErr, err := run(os.Args[0],
		[]string{"GH_WANT_HELPER_PROCESS=1"},
		"-test.run=TestHelperProcess", "--", "gh", "issue", "list", "error")
	assert.EqualError(t, err, "failed to run gh: process exited with error. error: exit status 1")
	assert.Equal(t, "", stdOut.String())
	assert.Equal(t, "process exited with error", stdErr.String())
}

func TestRESTClient(t *testing.T) {
	t.Cleanup(gock.Off)
	tempDir := t.TempDir()
	orig_GH_CONFIG_DIR := os.Getenv("GH_CONFIG_DIR")
	orig_GH_TOKEN := os.Getenv("GH_TOKEN")
	t.Cleanup(func() {
		os.Setenv("GH_CONFIG_DIR", orig_GH_CONFIG_DIR)
		os.Setenv("GH_TOKEN", orig_GH_TOKEN)
	})
	os.Setenv("GH_CONFIG_DIR", tempDir)
	os.Setenv("GH_TOKEN", "GH_TOKEN")

	gock.New("https://api.github.com").
		Get("/some/test/path").
		MatchHeader("Authorization", "token GH_TOKEN").
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
	t.Cleanup(gock.Off)
	tempDir := t.TempDir()
	orig_GH_CONFIG_DIR := os.Getenv("GH_CONFIG_DIR")
	orig_GH_TOKEN := os.Getenv("GH_TOKEN")
	t.Cleanup(func() {
		os.Setenv("GH_CONFIG_DIR", orig_GH_CONFIG_DIR)
		os.Setenv("GH_TOKEN", orig_GH_TOKEN)
	})
	os.Setenv("GH_CONFIG_DIR", tempDir)
	os.Setenv("GH_TOKEN", "GH_TOKEN")

	gock.New("https://api.github.com").
		Post("/graphql").
		MatchHeader("Authorization", "token GH_TOKEN").
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

func TestHTTPClient(t *testing.T) {
	t.Cleanup(gock.Off)
	tempDir := t.TempDir()
	orig_GH_CONFIG_DIR := os.Getenv("GH_CONFIG_DIR")
	orig_GH_TOKEN := os.Getenv("GH_TOKEN")
	t.Cleanup(func() {
		os.Setenv("GH_CONFIG_DIR", orig_GH_CONFIG_DIR)
		os.Setenv("GH_TOKEN", orig_GH_TOKEN)
	})
	os.Setenv("GH_CONFIG_DIR", tempDir)
	os.Setenv("GH_TOKEN", "GH_TOKEN")

	gock.New("https://api.github.com").
		Get("/some/test/path").
		MatchHeader("Authorization", "token GH_TOKEN").
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
	cfg := testConfig()

	tests := []struct {
		name          string
		opts          *api.ClientOptions
		wantAuthToken string
		wantHost      string
	}{
		{
			name: "honors consumer provided ClientOptions",
			opts: &api.ClientOptions{
				Host:      "test.com",
				AuthToken: "token_from_opts",
			},
			wantAuthToken: "token_from_opts",
			wantHost:      "test.com",
		},
		{
			name:          "uses config values if there are no consumer provided ClientOptions",
			opts:          &api.ClientOptions{},
			wantAuthToken: "token",
			wantHost:      "github.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := resolveOptions(tt.opts, cfg)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantHost, tt.opts.Host)
			assert.Equal(t, tt.wantAuthToken, tt.opts.AuthToken)
		})
	}
}

func testConfig() config.Config {
	var data = `
hosts:
  github.com:
    user: user1
    oauth_token: token
    git_protocol: ssh
`
	cfg, _ := config.FromString(data)
	return cfg
}

func printPendingMocks(mocks []gock.Mock) string {
	paths := []string{}
	for _, mock := range mocks {
		paths = append(paths, mock.Request().URLStruct.String())
	}
	return fmt.Sprintf("%d unmatched mocks: %s", len(paths), strings.Join(paths, ", "))
}
