package repository

import (
	"fmt"
	"os"
	"testing"

	"github.com/cli/go-gh/v2/internal/git"
	"github.com/cli/go-gh/v2/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantHost  string
		wantOwner string
		wantName  string
	}{
		{
			name:      "OWNER/REPO uses github.com by default",
			input:     "OWNER/REPO",
			wantHost:  "github.com",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:      "HOST/OWNER/REPO",
			input:     "example.org/OWNER/REPO",
			wantHost:  "example.org",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:      "HTTPS URL",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given no GH_HOST and no configured hosts
			t.Setenv("GH_HOST", "")
			testutils.StubConfig(t, "")

			// When the repository is parsed
			r, err := Parse(tt.input)

			// Then its host, owner, and name are returned
			require.NoError(t, err)
			assert.Equal(t, tt.wantHost, r.Host)
			assert.Equal(t, tt.wantOwner, r.Owner)
			assert.Equal(t, tt.wantName, r.Name)
		})
	}
}

func TestParseWithGHHost(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantHost string
	}{
		{
			name:     "OWNER/REPO uses GH_HOST",
			input:    "OWNER/REPO",
			wantHost: "override.com",
		},
		{
			name:     "HOST/OWNER/REPO ignores GH_HOST",
			input:    "example.com/OWNER/REPO",
			wantHost: "example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given GH_HOST names a host
			t.Setenv("GH_HOST", "override.com")
			testutils.StubConfig(t, "")

			// When the repository is parsed
			r, err := Parse(tt.input)

			// Then the host comes from the input when it has one, and GH_HOST otherwise
			require.NoError(t, err)
			assert.Equal(t, tt.wantHost, r.Host)
			assert.Equal(t, "OWNER", r.Owner)
			assert.Equal(t, "REPO", r.Name)
		})
	}
}

func TestParseUsesTheOnlyConfiguredHost(t *testing.T) {
	// Given no GH_HOST and a single configured host
	t.Setenv("GH_HOST", "")
	testutils.StubConfig(t, `
hosts:
  enterprise.com:
    user: user2
    oauth_token: yyyyyyyyyyyyyyyyyyyy
    git_protocol: https
`)

	// When a repository without a host is parsed
	r, err := Parse("OWNER/REPO")

	// Then the configured host is used
	require.NoError(t, err)
	assert.Equal(t, "enterprise.com", r.Host)
	assert.Equal(t, "OWNER", r.Owner)
	assert.Equal(t, "REPO", r.Name)
}

func TestParseRejectsMalformedRepository(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// When a malformed repository is parsed
			_, err := Parse(tt.input)

			// Then the expected format is explained
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestParseWithHost(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantHost  string
		wantOwner string
		wantName  string
	}{
		{
			name:      "OWNER/REPO uses the given host",
			input:     "OWNER/REPO",
			wantHost:  "ghe.example",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:      "HOST/OWNER/REPO ignores the given host",
			input:     "example.org/OWNER/REPO",
			wantHost:  "example.org",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:      "HTTPS URL ignores the given host",
			input:     "https://example.org/OWNER/REPO.git",
			wantHost:  "example.org",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:      "SSH URL ignores the given host",
			input:     "git@example.org:OWNER/REPO.git",
			wantHost:  "example.org",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// When the repository is parsed with a fallback host
			r, err := ParseWithHost(tt.input, "ghe.example")

			// Then its host, owner, and name are returned
			require.NoError(t, err)
			assert.Equal(t, tt.wantHost, r.Host)
			assert.Equal(t, tt.wantOwner, r.Owner)
			assert.Equal(t, tt.wantName, r.Name)
		})
	}
}

func TestParseWithHostRejectsMalformedRepository(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// When a malformed repository is parsed with a fallback host
			_, err := ParseWithHost(tt.input, "github.com")

			// Then the expected format is explained
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestCurrentUsesResolvedBaseRemote(t *testing.T) {
	// Given a higher-ranked fork remote and a parent remote selected by gh
	loggedInTo(t, "github.com")
	inNewRepo(t)
	runGit(t, "remote", "add", "origin", "https://github.com/parent-org/example.git")
	runGit(t, "remote", "add", "github", "https://github.com/my-user/example.git")
	runGit(t, "config", "remote.origin.gh-resolved", "base")

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
	loggedInTo(t, "github.com")
	inRepoWithOrigin(t, "https://github.com/my-user/example.git")
	runGit(t, "config", "remote.origin.gh-resolved", "ghe.example/parent-org/example")

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
	testutils.StubSSH(t, map[string]testutils.SSHResponse{"github.com-work": testutils.SSHReportsHostname("github.com")})
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
	testutils.StubSSH(t, map[string]testutils.SSHResponse{"github.com-work": testutils.SSHReportsHostname("github.com")})
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
	testutils.StubSSH(t, map[string]testutils.SSHResponse{"github.company.example": testutils.SSHReportsHostname("192.0.2.10")})
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

// inNewRepo changes into a new git repository with no remotes, ignoring any GH_REPO from the environment.
func inNewRepo(t *testing.T) {
	t.Helper()
	t.Setenv("GH_REPO", "")
	t.Chdir(t.TempDir())
	runGit(t, "init", "--quiet")
}

// inRepoWithOrigin changes into a new git repository whose origin fetches from fetchURL and pushes to pushURLs in order.
func inRepoWithOrigin(t *testing.T, fetchURL string, pushURLs ...string) {
	t.Helper()
	inNewRepo(t)
	runGit(t, "remote", "add", "origin", fetchURL)
	for _, pushURL := range pushURLs {
		runGit(t, "remote", "set-url", "--add", "--push", "origin", pushURL)
	}
}

func runGit(t *testing.T, args ...string) {
	t.Helper()
	_, _, err := git.Exec(args...)
	require.NoError(t, err)
}

func TestMain(m *testing.M) {
	testutils.RunFakeSSHIfRequested()
	os.Exit(m.Run())
}
