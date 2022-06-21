package auth

import (
	"os"
	"testing"

	"github.com/cli/go-gh/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestTokenForHost(t *testing.T) {
	orig_GITHUB_TOKEN := os.Getenv("GITHUB_TOKEN")
	orig_GITHUB_ENTERPRISE_TOKEN := os.Getenv("GITHUB_ENTERPRISE_TOKEN")
	orig_GH_TOKEN := os.Getenv("GH_TOKEN")
	orig_GH_ENTERPRISE_TOKEN := os.Getenv("GH_ENTERPRISE_TOKEN")
	t.Cleanup(func() {
		os.Setenv("GITHUB_TOKEN", orig_GITHUB_TOKEN)
		os.Setenv("GITHUB_ENTERPRISE_TOKEN", orig_GITHUB_ENTERPRISE_TOKEN)
		os.Setenv("GH_TOKEN", orig_GH_TOKEN)
		os.Setenv("GH_ENTERPRISE_TOKEN", orig_GH_ENTERPRISE_TOKEN)
	})

	tests := []struct {
		name                  string
		host                  string
		githubToken           string
		githubEnterpriseToken string
		ghToken               string
		ghEnterpriseToken     string
		config                *config.Config
		wantToken             string
		wantSource            string
		wantNotFound          bool
	}{
		{
			name:         "token for github.com with no env tokens and no config token",
			host:         "github.com",
			config:       testNoHostsConfig(),
			wantToken:    "",
			wantSource:   "oauth_token",
			wantNotFound: true,
		},
		{
			name:         "token for enterprise.com with no env tokens and no config token",
			host:         "enterprise.com",
			config:       testNoHostsConfig(),
			wantToken:    "",
			wantSource:   "oauth_token",
			wantNotFound: true,
		},
		{
			name:        "token for github.com with GH_TOKEN, GITHUB_TOKEN, and config token",
			host:        "github.com",
			ghToken:     "GH_TOKEN",
			githubToken: "GITHUB_TOKEN",
			config:      testHostsConfig(),
			wantToken:   "GH_TOKEN",
			wantSource:  "GH_TOKEN",
		},
		{
			name:        "token for github.com with GITHUB_TOKEN, and config token",
			host:        "github.com",
			githubToken: "GITHUB_TOKEN",
			config:      testHostsConfig(),
			wantToken:   "GITHUB_TOKEN",
			wantSource:  "GITHUB_TOKEN",
		},
		{
			name:       "token for github.com with config token",
			host:       "github.com",
			config:     testHostsConfig(),
			wantToken:  "xxxxxxxxxxxxxxxxxxxx",
			wantSource: "oauth_token",
		},
		{
			name:                  "token for enterprise.com with GH_ENTERPRISE_TOKEN, GITHUB_ENTERPRISE_TOKEN, and config token",
			host:                  "enterprise.com",
			ghEnterpriseToken:     "GH_ENTERPRISE_TOKEN",
			githubEnterpriseToken: "GITHUB_ENTERPRISE_TOKEN",
			config:                testHostsConfig(),
			wantToken:             "GH_ENTERPRISE_TOKEN",
			wantSource:            "GH_ENTERPRISE_TOKEN",
		},
		{
			name:                  "token for enterprise.com with GITHUB_ENTERPRISE_TOKEN, and config token",
			host:                  "enterprise.com",
			githubEnterpriseToken: "GITHUB_ENTERPRISE_TOKEN",
			config:                testHostsConfig(),
			wantToken:             "GITHUB_ENTERPRISE_TOKEN",
			wantSource:            "GITHUB_ENTERPRISE_TOKEN",
		},
		{
			name:       "token for enterprise.com with config token",
			host:       "enterprise.com",
			config:     testHostsConfig(),
			wantToken:  "yyyyyyyyyyyyyyyyyyyy",
			wantSource: "oauth_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GITHUB_TOKEN", tt.githubToken)
			os.Setenv("GITHUB_ENTERPRISE_TOKEN", tt.githubEnterpriseToken)
			os.Setenv("GH_TOKEN", tt.ghToken)
			os.Setenv("GH_ENTERPRISE_TOKEN", tt.ghEnterpriseToken)
			token, source := tokenForHost(tt.config, tt.host)
			assert.Equal(t, tt.wantToken, token)
			assert.Equal(t, tt.wantSource, source)
		})
	}
}

func TestDefaultHost(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.Config
		ghHost       string
		wantHost     string
		wantSource   string
		wantNotFound bool
	}{
		{
			name:       "GH_HOST if set",
			config:     testHostsConfig(),
			ghHost:     "test.com",
			wantHost:   "test.com",
			wantSource: "GH_HOST",
		},
		{
			name:       "authenticated host if only one",
			config:     testSingleHostConfig(),
			wantHost:   "enterprise.com",
			wantSource: "hosts",
		},
		{
			name:         "default host if more than one authenticated host",
			config:       testHostsConfig(),
			wantHost:     "github.com",
			wantSource:   "default",
			wantNotFound: true,
		},
		{
			name:         "default host if no authenticated host",
			config:       testNoHostsConfig(),
			wantHost:     "github.com",
			wantSource:   "default",
			wantNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ghHost != "" {
				k := "GH_HOST"
				old := os.Getenv(k)
				os.Setenv(k, tt.ghHost)
				defer os.Setenv(k, old)
			}
			host, source := defaultHost(tt.config)
			assert.Equal(t, tt.wantHost, host)
			assert.Equal(t, tt.wantSource, source)
		})
	}
}

func TestKnownHosts(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		ghHost    string
		ghToken   string
		wantHosts []string
	}{
		{
			name:      "no known hosts",
			config:    testNoHostsConfig(),
			wantHosts: []string{},
		},
		{
			name:      "includes GH_HOST",
			config:    testNoHostsConfig(),
			ghHost:    "test.com",
			wantHosts: []string{"test.com"},
		},
		{
			name:      "includes authenticated hosts",
			config:    testHostsConfig(),
			wantHosts: []string{"github.com", "enterprise.com"},
		},
		{
			name:      "includes default host if environment auth token",
			config:    testNoHostsConfig(),
			ghToken:   "TOKEN",
			wantHosts: []string{"github.com"},
		},
		{
			name:      "deduplicates hosts",
			config:    testHostsConfig(),
			ghHost:    "test.com",
			ghToken:   "TOKEN",
			wantHosts: []string{"test.com", "github.com", "enterprise.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ghHost != "" {
				k := "GH_HOST"
				old := os.Getenv(k)
				os.Setenv(k, tt.ghHost)
				defer os.Setenv(k, old)
			}
			if tt.ghToken != "" {
				k := "GH_TOKEN"
				old := os.Getenv(k)
				os.Setenv(k, tt.ghToken)
				defer os.Setenv(k, old)
			}
			hosts := knownHosts(tt.config)
			assert.Equal(t, tt.wantHosts, hosts)
		})
	}
}

func testNoHostsConfig() *config.Config {
	var data = ``
	return config.ReadFromString(data)
}

func testSingleHostConfig() *config.Config {
	var data = `
hosts:
  enterprise.com:
    user: user2
    oauth_token: yyyyyyyyyyyyyyyyyyyy
    git_protocol: https
`
	return config.ReadFromString(data)
}

func testHostsConfig() *config.Config {
	var data = `
hosts:
  github.com:
    user: user1
    oauth_token: xxxxxxxxxxxxxxxxxxxx
    git_protocol: ssh
  enterprise.com:
    user: user2
    oauth_token: yyyyyyyyyyyyyyyyyyyy
    git_protocol: https
`
	return config.ReadFromString(data)
}
