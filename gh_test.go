package gh

import (
	"fmt"
	"os"
	"testing"

	"github.com/cli/go-gh/internal/httpmock"
	"github.com/cli/go-gh/pkg/api"
	"github.com/stretchr/testify/assert"
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
	tempDir := t.TempDir()
	orig_GH_CONFIG_DIR := os.Getenv("GH_CONFIG_DIR")
	orig_GH_TOKEN := os.Getenv("GH_TOKEN")
	t.Cleanup(func() {
		os.Setenv("GH_CONFIG_DIR", orig_GH_CONFIG_DIR)
		os.Setenv("GH_TOKEN", orig_GH_TOKEN)
	})
	os.Setenv("GH_CONFIG_DIR", tempDir)
	os.Setenv("GH_TOKEN", "GH_TOKEN")

	http := httpmock.NewRegistry(t)
	http.Register(
		httpmock.REST("GET", "some/test/path"),
		httpmock.StatusStringResponse(200, `{"message": "success"}`),
	)

	client, err := RESTClient(&api.ClientOptions{Transport: http})
	assert.NoError(t, err)

	res := struct{ Message string }{}
	err = client.Do("GET", "some/test/path", nil, &res)
	assert.NoError(t, err)
	assert.Equal(t, "success", res.Message)
	assert.Equal(t, "api.github.com", http.Requests[0].URL.Hostname())
	assert.Equal(t, "token GH_TOKEN", http.Requests[0].Header.Get("Authorization"))
}

func TestGQLClient(t *testing.T) {
	tempDir := t.TempDir()
	orig_GH_CONFIG_DIR := os.Getenv("GH_CONFIG_DIR")
	orig_GH_TOKEN := os.Getenv("GH_TOKEN")
	t.Cleanup(func() {
		os.Setenv("GH_CONFIG_DIR", orig_GH_CONFIG_DIR)
		os.Setenv("GH_TOKEN", orig_GH_TOKEN)
	})
	os.Setenv("GH_CONFIG_DIR", tempDir)
	os.Setenv("GH_TOKEN", "GH_TOKEN")

	http := httpmock.NewRegistry(t)
	http.Register(
		httpmock.GQL("QUERY"),
		httpmock.StringResponse(`{"data":{"viewer":{"login":"hubot"}}}`),
	)

	client, err := GQLClient(&api.ClientOptions{Transport: http})
	assert.NoError(t, err)

	vars := map[string]interface{}{"var": "test"}
	res := struct{ Viewer struct{ Login string } }{}
	err = client.Do("QUERY", vars, &res)
	assert.NoError(t, err)
	assert.Equal(t, "hubot", res.Viewer.Login)
	assert.Equal(t, "api.github.com", http.Requests[0].URL.Hostname())
	assert.Equal(t, "token GH_TOKEN", http.Requests[0].Header.Get("Authorization"))
}

func TestHTTPClient(t *testing.T) {
	tempDir := t.TempDir()
	orig_GH_CONFIG_DIR := os.Getenv("GH_CONFIG_DIR")
	orig_GH_TOKEN := os.Getenv("GH_TOKEN")
	t.Cleanup(func() {
		os.Setenv("GH_CONFIG_DIR", orig_GH_CONFIG_DIR)
		os.Setenv("GH_TOKEN", orig_GH_TOKEN)
	})
	os.Setenv("GH_CONFIG_DIR", tempDir)
	os.Setenv("GH_TOKEN", "GH_TOKEN")

	http := httpmock.NewRegistry(t)
	http.Register(
		httpmock.REST("GET", "some/test/path"),
		httpmock.StatusStringResponse(200, `{"message": "success"}`),
	)

	client, err := HTTPClient(&api.ClientOptions{Transport: http})
	assert.NoError(t, err)

	res, err := client.Get("https://api.github.com/some/test/path")
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "api.github.com", http.Requests[0].URL.Hostname())
	assert.Equal(t, "token GH_TOKEN", http.Requests[0].Header.Get("Authorization"))
}
