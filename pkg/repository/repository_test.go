package repository

import (
	"net/url"
	"testing"

	"github.com/cli/go-gh/v2/internal/git"
	"github.com/cli/go-gh/v2/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestTranslateRemotesRefreshesRepositoryInfo(t *testing.T) {
	translate := func(u *url.URL) *url.URL {
		translated := *u
		translated.Host = "github.com"
		return &translated
	}

	t.Run("SSH host alias", func(t *testing.T) {
		u, err := url.Parse("ssh://git@github.com-work/owner/repo.git")
		if err != nil {
			t.Fatal(err)
		}
		remotes := git.RemoteSet{&git.Remote{
			Name:     "origin",
			FetchURL: u,
			Host:     "github.com-work",
			Owner:    "owner",
			Repo:     "repo",
		}}

		translateRemotes(remotes, translate)

		filtered := remotes.FilterByHosts([]string{"github.com"})
		assert.Len(t, filtered, 1)
		assert.Equal(t, "github.com", remotes[0].Host)
		assert.Equal(t, "owner", remotes[0].Owner)
		assert.Equal(t, "repo", remotes[0].Repo)
	})

	t.Run("invalid fetch URL falls back to push URL", func(t *testing.T) {
		fetchURL, err := url.Parse("ssh://git@github.com-work/")
		if err != nil {
			t.Fatal(err)
		}
		pushURL, err := url.Parse("ssh://git@github.com-work/owner/repo.git")
		if err != nil {
			t.Fatal(err)
		}
		remotes := git.RemoteSet{&git.Remote{
			Name:     "origin",
			FetchURL: fetchURL,
			PushURL:  pushURL,
			Host:     "github.com-work",
		}}

		translateRemotes(remotes, translate)

		assert.Equal(t, "github.com", remotes[0].Host)
		assert.Equal(t, "owner", remotes[0].Owner)
		assert.Equal(t, "repo", remotes[0].Repo)
	})
}

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
