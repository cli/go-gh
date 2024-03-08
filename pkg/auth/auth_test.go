package auth

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestTokenForHost(t *testing.T) {
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
		{
			name:        "token for tenant with GH_TOKEN, GITHUB_TOKEN, and config token",
			host:        "tenant.ghe.com",
			ghToken:     "GH_TOKEN",
			githubToken: "GITHUB_TOKEN",
			config:      testHostsConfig(),
			wantToken:   "GH_TOKEN",
			wantSource:  "GH_TOKEN",
		},
		{
			name:        "token for tenant with GITHUB_TOKEN, and config token",
			host:        "tenant.ghe.com",
			githubToken: "GITHUB_TOKEN",
			config:      testHostsConfig(),
			wantToken:   "GITHUB_TOKEN",
			wantSource:  "GITHUB_TOKEN",
		},
		{
			name:       "token for tenant with config token",
			host:       "tenant.ghe.com",
			config:     testHostsConfig(),
			wantToken:  "zzzzzzzzzzzzzzzzzzzz",
			wantSource: "oauth_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_TOKEN", tt.githubToken)
			t.Setenv("GITHUB_ENTERPRISE_TOKEN", tt.githubEnterpriseToken)
			t.Setenv("GH_TOKEN", tt.ghToken)
			t.Setenv("GH_ENTERPRISE_TOKEN", tt.ghEnterpriseToken)
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
				t.Setenv("GH_HOST", tt.ghHost)
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
			wantHosts: []string{"github.com", "enterprise.com", "tenant.ghe.com"},
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
			wantHosts: []string{"test.com", "github.com", "enterprise.com", "tenant.ghe.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ghHost != "" {
				t.Setenv("GH_HOST", tt.ghHost)
			}
			if tt.ghToken != "" {
				t.Setenv("GH_TOKEN", tt.ghToken)
			}
			hosts := knownHosts(tt.config)
			assert.Equal(t, tt.wantHosts, hosts)
		})
	}
}

func TestIsEnterprise(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantOut bool
	}{
		{
			name:    "github",
			host:    "github.com",
			wantOut: false,
		},
		{
			name:    "localhost",
			host:    "github.localhost",
			wantOut: false,
		},
		{
			name:    "enterprise",
			host:    "mygithub.com",
			wantOut: true,
		},
		{
			name:    "tenant",
			host:    "tenant.ghe.com",
			wantOut: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := IsEnterprise(tt.host)
			assert.Equal(t, tt.wantOut, out)
		})
	}
}

func TestIsTenancy(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantOut bool
	}{
		{
			name:    "github",
			host:    "github.com",
			wantOut: false,
		},
		{
			name:    "localhost",
			host:    "github.localhost",
			wantOut: false,
		},
		{
			name:    "enterprise",
			host:    "mygithub.com",
			wantOut: false,
		},
		{
			name:    "tenant",
			host:    "tenant.ghe.com",
			wantOut: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := IsTenancy(tt.host)
			assert.Equal(t, tt.wantOut, out)
		})
	}
}

func TestNormalizeHostname(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		wantHost string
	}{
		{
			name:     "github domain",
			host:     "test.github.com",
			wantHost: "github.com",
		},
		{
			name:     "capitalized",
			host:     "GitHub.com",
			wantHost: "github.com",
		},
		{
			name:     "localhost domain",
			host:     "test.github.localhost",
			wantHost: "github.localhost",
		},
		{
			name:     "enterprise domain",
			host:     "mygithub.com",
			wantHost: "mygithub.com",
		},
		{
			name:     "bare tenant",
			host:     "tenant.ghe.com",
			wantHost: "tenant.ghe.com",
		},
		{
			name:     "subdomained tenant",
			host:     "api.tenant.ghe.com",
			wantHost: "tenant.ghe.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized := normalizeHostname(tt.host)
			assert.Equal(t, tt.wantHost, normalized)
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
  tenant.ghe.com:
    user: user3
    oauth_token: zzzzzzzzzzzzzzzzzzzz
    git_protocol: https
`
	return config.ReadFromString(data)
}
