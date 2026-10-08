package repository

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/internal/git"
	"github.com/cli/go-gh/v2/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	testutils.StubConfig(t, "")

	tests := []struct {
		name         string
		input        string
		hostOverride string
		wantOwner    string
		wantName     string
		wantHost     string
		wantErr      string
	}{
		{
			name:      "OWNER/REPO combo",
			input:     "OWNER/REPO",
			wantHost:  "github.com",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:    "too few elements",
			input:   "OWNER",
			wantErr: `expected the "[HOST/]OWNER/REPO" format, got "OWNER"`,
		},
		{
			name:    "too many elements",
			input:   "a/b/c/d",
			wantErr: `expected the "[HOST/]OWNER/REPO" format, got "a/b/c/d"`,
		},
		{
			name:    "blank value",
			input:   "a/",
			wantErr: `expected the "[HOST/]OWNER/REPO" format, got "a/"`,
		},
		{
			name:      "with hostname",
			input:     "example.org/OWNER/REPO",
			wantHost:  "example.org",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:      "full URL",
			input:     "https://example.org/OWNER/REPO.git",
			wantHost:  "example.org",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:      "SSH URL",
			input:     "git@example.org:OWNER/REPO.git",
			wantHost:  "example.org",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:         "OWNER/REPO with default host override",
			input:        "OWNER/REPO",
			hostOverride: "override.com",
			wantHost:     "override.com",
			wantOwner:    "OWNER",
			wantName:     "REPO",
		},
		{
			name:         "HOST/OWNER/REPO with default host override",
			input:        "example.com/OWNER/REPO",
			hostOverride: "override.com",
			wantHost:     "example.com",
			wantOwner:    "OWNER",
			wantName:     "REPO",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GH_CONFIG_DIR", "nonexistant")
			if tt.hostOverride != "" {
				t.Setenv("GH_HOST", tt.hostOverride)
			}
			r, err := Parse(tt.input)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantHost, r.Host)
			assert.Equal(t, tt.wantOwner, r.Owner)
			assert.Equal(t, tt.wantName, r.Name)
		})
	}
}

func TestParse_hostFromConfig(t *testing.T) {
	var cfgStr = `
hosts:
  enterprise.com:
    user: user2
    oauth_token: yyyyyyyyyyyyyyyyyyyy
    git_protocol: https
`
	testutils.StubConfig(t, cfgStr)
	r, err := Parse("OWNER/REPO")
	assert.NoError(t, err)
	assert.Equal(t, "enterprise.com", r.Host)
	assert.Equal(t, "OWNER", r.Owner)
	assert.Equal(t, "REPO", r.Name)
}

func TestParseWithHost(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		host      string
		wantOwner string
		wantName  string
		wantHost  string
		wantErr   string
	}{
		{
			name:      "OWNER/REPO combo",
			input:     "OWNER/REPO",
			host:      "github.com",
			wantHost:  "github.com",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:    "too few elements",
			input:   "OWNER",
			host:    "github.com",
			wantErr: `expected the "[HOST/]OWNER/REPO" format, got "OWNER"`,
		},
		{
			name:    "too many elements",
			input:   "a/b/c/d",
			host:    "github.com",
			wantErr: `expected the "[HOST/]OWNER/REPO" format, got "a/b/c/d"`,
		},
		{
			name:    "blank value",
			input:   "a/",
			host:    "github.com",
			wantErr: `expected the "[HOST/]OWNER/REPO" format, got "a/"`,
		},
		{
			name:      "with hostname",
			input:     "example.org/OWNER/REPO",
			host:      "github.com",
			wantHost:  "example.org",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:      "full URL",
			input:     "https://example.org/OWNER/REPO.git",
			host:      "github.com",
			wantHost:  "example.org",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:      "SSH URL",
			input:     "git@example.org:OWNER/REPO.git",
			host:      "github.com",
			wantHost:  "example.org",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ParseWithHost(tt.input, tt.host)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantHost, r.Host)
			assert.Equal(t, tt.wantOwner, r.Owner)
			assert.Equal(t, tt.wantName, r.Name)
		})
	}
}

func TestCurrentUsesResolvedBaseRemote(t *testing.T) {
	// Given a higher-ranked fork remote and a parent remote selected by gh
	t.Setenv("GH_REPO", "")
	testutils.StubConfig(t, `
hosts:
  github.com:
    oauth_token: token
`)
	t.Chdir(t.TempDir())
	_, _, err := git.Exec("init", "--quiet")
	require.NoError(t, err)
	_, _, err = git.Exec("remote", "add", "origin", "git@github.com:parent-org/example.git")
	require.NoError(t, err)
	_, _, err = git.Exec("remote", "add", "github", "git@github.com:my-user/example.git")
	require.NoError(t, err)
	_, _, err = git.Exec("config", "remote.origin.gh-resolved", "base")
	require.NoError(t, err)

	// When the current repository is resolved
	repository, err := Current()

	// Then the remote selected by gh is used instead of the name-based ranking
	require.NoError(t, err)
	assert.Equal(t, "github.com", repository.Host)
	assert.Equal(t, "parent-org", repository.Owner)
	assert.Equal(t, "example", repository.Name)
}

func TestCurrentUsesExplicitResolvedRepository(t *testing.T) {
	// Given a remote whose gh resolution names a repository on a different host
	t.Setenv("GH_REPO", "")
	testutils.StubConfig(t, `
hosts:
  github.com:
    oauth_token: token
`)
	t.Chdir(t.TempDir())
	_, _, err := git.Exec("init", "--quiet")
	require.NoError(t, err)
	_, _, err = git.Exec("remote", "add", "origin", "git@github.com:my-user/example.git")
	require.NoError(t, err)
	_, _, err = git.Exec("config", "remote.origin.gh-resolved", "ghe.example/parent-org/example")
	require.NoError(t, err)

	// When the current repository is resolved
	repository, err := Current()

	// Then the explicit repository is returned using the remote's host
	require.NoError(t, err)
	assert.Equal(t, "github.com", repository.Host)
	assert.Equal(t, "parent-org", repository.Owner)
	assert.Equal(t, "example", repository.Name)
}

const noKnownHostRemoteErr = "unable to determine current repository, none of the git remotes configured for this repository point to a known GitHub host"

func TestCurrentResolvesFetchURLThroughSSHHostAlias(t *testing.T) {
	// Given a fetch URL that uses an SSH host alias for github.com
	loggedInTo(t, "github.com")
	sshConfigMapsHost(t, "github.com-work", "github.com")
	inRepoWithOrigin(t, "git@github.com-work:acme/widgets.git")

	// When the current repository is resolved
	repository, err := Current()

	// Then the repository is on the aliased host
	require.NoError(t, err)
	assert.Equal(t, "github.com", repository.Host)
	assert.Equal(t, "acme", repository.Owner)
	assert.Equal(t, "widgets", repository.Name)
}

func TestCurrentPrefersFetchURLOverPushURL(t *testing.T) {
	// Given fetch and push URLs that name different repositories
	loggedInTo(t, "github.com")
	inRepoWithOrigin(t, "https://github.com/acme/widgets", "https://github.com/my-user/widgets")

	// When the current repository is resolved
	repository, err := Current()

	// Then the fetch URL's repository is used
	require.NoError(t, err)
	assert.Equal(t, "github.com", repository.Host)
	assert.Equal(t, "acme", repository.Owner)
	assert.Equal(t, "widgets", repository.Name)
}

func TestCurrentFallsBackToPushURLWhenFetchURLIsLocalPath(t *testing.T) {
	// Given a fetch URL that is a local path
	loggedInTo(t, "github.com")
	inRepoWithOrigin(t, "/srv/git/widgets.git", "https://github.com/acme/widgets")

	// When the current repository is resolved
	repository, err := Current()

	// Then the push URL's repository is used
	require.NoError(t, err)
	assert.Equal(t, "github.com", repository.Host)
	assert.Equal(t, "acme", repository.Owner)
	assert.Equal(t, "widgets", repository.Name)
}

func TestCurrentResolvesPushURLThroughSSHHostAlias(t *testing.T) {
	// Given a local fetch path and a push URL that uses an SSH host alias for github.com
	loggedInTo(t, "github.com")
	sshConfigMapsHost(t, "github.com-work", "github.com")
	inRepoWithOrigin(t, "/srv/git/widgets.git", "git@github.com-work:acme/widgets.git")

	// When the current repository is resolved
	repository, err := Current()

	// Then the push URL's repository is on the aliased host
	require.NoError(t, err)
	assert.Equal(t, "github.com", repository.Host)
	assert.Equal(t, "acme", repository.Owner)
	assert.Equal(t, "widgets", repository.Name)
}

func TestCurrentIgnoresPushURLWhenFetchURLIsOnUnknownHost(t *testing.T) {
	// Given a fetch URL on a host I'm not logged in to and a push URL on one I am
	loggedInTo(t, "github.com")
	inRepoWithOrigin(t, "https://gitlab.com/acme/widgets", "https://github.com/acme/widgets")

	// When the current repository is resolved
	_, err := Current()

	// Then the remote is not treated as pointing to a known host
	require.EqualError(t, err, noKnownHostRemoteErr)
}

func TestCurrentUsesLastPushURL(t *testing.T) {
	// Given a local fetch path and several push URLs
	loggedInTo(t, "github.com")
	inRepoWithOrigin(t, "/srv/git/widgets.git", "https://github.com/acme/widgets", "https://github.com/backup/mirror")

	// When the current repository is resolved
	repository, err := Current()

	// Then the last push URL's repository is used
	require.NoError(t, err)
	assert.Equal(t, "github.com", repository.Host)
	assert.Equal(t, "backup", repository.Owner)
	assert.Equal(t, "mirror", repository.Name)
}

func TestCurrentIgnoresEarlierPushURLsWhenLastIsLocalPath(t *testing.T) {
	// Given a local fetch path and a last push URL that is also a local path
	loggedInTo(t, "github.com")
	inRepoWithOrigin(t, "/srv/git/widgets.git", "https://github.com/acme/widgets", "/srv/git/mirror.git")

	// When the current repository is resolved
	_, err := Current()

	// Then the earlier push URL is not used
	require.EqualError(t, err, noKnownHostRemoteErr)
}

func TestCurrentRejectsKnownHostThatSSHMapsToAnotherHost(t *testing.T) {
	// Given the only known host is one that SSH config maps to an IP address
	loggedInTo(t, "github.company.example")
	sshConfigMapsHost(t, "github.company.example", "192.0.2.10")
	inRepoWithOrigin(t, "git@github.company.example:acme/widgets.git")

	// When the current repository is resolved
	_, err := Current()

	// Then the remote is not treated as pointing to a known host
	require.EqualError(t, err, noKnownHostRemoteErr)
}

// loggedInTo makes host the only known GitHub host, ignoring any host or token from the environment.
func loggedInTo(t *testing.T, host string) {
	t.Helper()
	for _, name := range []string{"GH_HOST", "GH_TOKEN", "GITHUB_TOKEN", "GH_ENTERPRISE_TOKEN", "GITHUB_ENTERPRISE_TOKEN"} {
		t.Setenv(name, "")
	}
	testutils.StubConfig(t, fmt.Sprintf("hosts:\n  %s:\n    oauth_token: token\n", host))
}

// inRepoWithOrigin changes into a new git repository whose origin fetches from fetchURL and pushes to pushURLs in order.
func inRepoWithOrigin(t *testing.T, fetchURL string, pushURLs ...string) {
	t.Helper()
	t.Setenv("GH_REPO", "")
	t.Chdir(t.TempDir())
	_, _, err := git.Exec("init", "--quiet")
	require.NoError(t, err)
	_, _, err = git.Exec("remote", "add", "origin", fetchURL)
	require.NoError(t, err)
	for _, pushURL := range pushURLs {
		_, _, err = git.Exec("remote", "set-url", "--add", "--push", "origin", pushURL)
		require.NoError(t, err)
	}
}

const fakeSSHHostnameEnv = "GO_GH_TEST_FAKE_SSH_HOSTNAME"

// sshConfigMapsHost puts a fake ssh first on PATH whose "ssh -G" reports hostname for alias.
// The fake is this test binary, which TestMain runs as ssh when fakeSSHHostnameEnv is set.
func sshConfigMapsHost(t *testing.T, alias, hostname string) {
	t.Helper()
	dir := t.TempDir()
	name := "ssh"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	copyTestBinary(t, filepath.Join(dir, name))
	t.Setenv(fakeSSHHostnameEnv, alias+"="+hostname)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func copyTestBinary(t *testing.T, dst string) {
	t.Helper()
	src, err := os.Executable()
	require.NoError(t, err)
	in, err := os.Open(src)
	require.NoError(t, err)
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0o755)
	require.NoError(t, err)
	_, err = io.Copy(out, in)
	require.NoError(t, err)
	require.NoError(t, out.Close())
}

func TestMain(m *testing.M) {
	if mapping, ok := os.LookupEnv(fakeSSHHostnameEnv); ok {
		runFakeSSH(mapping, os.Args[1:])
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// runFakeSSH mimics "ssh -G HOST", printing the mapped hostname for the alias and HOST itself otherwise.
func runFakeSSH(mapping string, args []string) {
	if len(args) != 2 || args[0] != "-G" {
		fmt.Fprintf(os.Stderr, "fake ssh: unexpected arguments %q\n", args)
		os.Exit(1)
	}
	host := args[1]
	alias, hostname, _ := strings.Cut(mapping, "=")
	if strings.EqualFold(host, alias) {
		host = hostname
	}
	fmt.Printf("hostname %s\n", host)
}
