package repository

import (
	"os"
	"testing"

	"github.com/cli/go-gh/v2/internal/git"
	"github.com/cli/go-gh/v2/internal/testutils"
	"github.com/stretchr/testify/assert"
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

func TestCurrentPrefersResolvedRemote(t *testing.T) {
	testutils.StubConfig(t, `
hosts:
  github.com:
    oauth_token: token
`)

	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	assert.NoError(t, err)
	assert.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	_, _, err = git.Exec("init", "--quiet")
	assert.NoError(t, err)
	_, _, err = git.Exec("remote", "add", "origin", "git@github.com:parent-org/example.git")
	assert.NoError(t, err)
	_, _, err = git.Exec("remote", "add", "github", "git@github.com:my-user/example.git")
	assert.NoError(t, err)
	_, _, err = git.Exec("config", "remote.origin.gh-resolved", "base")
	assert.NoError(t, err)

	repository, err := Current()

	assert.NoError(t, err)
	assert.Equal(t, "github.com", repository.Host)
	assert.Equal(t, "parent-org", repository.Owner)
	assert.Equal(t, "example", repository.Name)
}
