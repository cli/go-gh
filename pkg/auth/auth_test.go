package auth

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	}{
		{
			name:       "given there is no env token and no config token, when we get the token for github.com, then it returns the empty string and default source",
			host:       "github.com",
			config:     testNoHostsConfig(),
			wantToken:  "",
			wantSource: defaultSource,
		},
		{
			name:       "given there is no env token and no config token, when we get the token for an enterprise server host, then it returns the empty string and default source",
			host:       "enterprise.com",
			config:     testNoHostsConfig(),
			wantToken:  "",
			wantSource: defaultSource,
		},
		{
			name:        "given GH_TOKEN and GITHUB_TOKEN and a config token are set, when we get the token for github.com, then it returns GH_TOKEN as the priority",
			host:        "github.com",
			ghToken:     "GH_TOKEN",
			githubToken: "GITHUB_TOKEN",
			config:      testHostsConfig(),
			wantToken:   "GH_TOKEN",
			wantSource:  ghToken,
		},
		{
			name:        "given GITHUB_TOKEN and a config token are set, when we get the token for github.com, then it returns GITHUB_TOKEN as the priority",
			host:        "github.com",
			githubToken: "GITHUB_TOKEN",
			config:      testHostsConfig(),
			wantToken:   "GITHUB_TOKEN",
			wantSource:  githubToken,
		},
		{
			name:       "given a config token is set for github.com, when we get the token, then it returns that token and oauth_token source",
			host:       "github.com",
			config:     testHostsConfig(),
			wantToken:  "xxxxxxxxxxxxxxxxxxxx",
			wantSource: oauthToken,
		},
		{
			name:        "given GH_TOKEN and GITHUB_TOKEN and a config token are set, when we get the token for any subdomain of ghe.com, then it returns GH_TOKEN as the priority",
			host:        "tenant.ghe.com",
			ghToken:     "GH_TOKEN",
			githubToken: "GITHUB_TOKEN",
			config:      testHostsConfig(),
			wantToken:   "GH_TOKEN",
			wantSource:  ghToken,
		},
		{
			name:        "given GITHUB_TOKEN and a config token are set, when we get the token for any subdomain of ghe.com, then it returns GITHUB_TOKEN as the priority",
			host:        "tenant.ghe.com",
			githubToken: "GITHUB_TOKEN",
			config:      testHostsConfig(),
			wantToken:   "GITHUB_TOKEN",
			wantSource:  githubToken,
		},
		{
			name:       "given a config token is set for a subdomain of ghe.com, when we get the token for that subdomain, then it returns that token and oauth_token source",
			host:       "tenant.ghe.com",
			config:     testHostsConfig(),
			wantToken:  "zzzzzzzzzzzzzzzzzzzz",
			wantSource: oauthToken,
		},
		{
			name:        "given GH_TOKEN and GITHUB_TOKEN and a config token are set, when we get the token for github.localhost, then it returns GH_TOKEN as the priority",
			host:        "github.localhost",
			ghToken:     "GH_TOKEN",
			githubToken: "GITHUB_TOKEN",
			config:      testHostsConfig(),
			wantToken:   "GH_TOKEN",
			wantSource:  ghToken,
		},
		{
			name:        "given GITHUB_TOKEN and a config token are set, when we get the token for any subdomain of github.localhost, then it returns GITHUB_TOKEN as the priority",
			host:        "github.localhost",
			githubToken: "GITHUB_TOKEN",
			config:      testHostsConfig(),
			wantToken:   "GITHUB_TOKEN",
			wantSource:  githubToken,
		},
		{
			name:                  "given GH_ENTERPRISE_TOKEN and GITHUB_ENTERPRISE_TOKEN and a config token are set, when we get the token for an enterprise server host, then it returns GH_ENTERPRISE_TOKEN as the priority",
			host:                  "enterprise.com",
			ghEnterpriseToken:     "GH_ENTERPRISE_TOKEN",
			githubEnterpriseToken: "GITHUB_ENTERPRISE_TOKEN",
			config:                testHostsConfig(),
			wantToken:             "GH_ENTERPRISE_TOKEN",
			wantSource:            ghEnterpriseToken,
		},
		{
			name:                  "given GITHUB_ENTERPRISE_TOKEN and a config token are set, when we get the token for an enterprise server host, then it returns GITHUB_ENTERPRISE_TOKEN as the priority",
			host:                  "enterprise.com",
			githubEnterpriseToken: "GITHUB_ENTERPRISE_TOKEN",
			config:                testHostsConfig(),
			wantToken:             "GITHUB_ENTERPRISE_TOKEN",
			wantSource:            githubEnterpriseToken,
		},
		{
			name:       "given a config token is set for an enterprise server host, when we get the token for that host, then it returns that token and oauth_token source",
			host:       "enterprise.com",
			config:     testHostsConfig(),
			wantToken:  "yyyyyyyyyyyyyyyyyyyy",
			wantSource: oauthToken,
		},
		{
			name:        "given GH_TOKEN or GITHUB_TOKEN are set, when I get the token for any host not owned by GitHub, we do not get those tokens",
			host:        "unknown.com",
			config:      testNoHostsConfig(),
			ghToken:     "GH_TOKEN",
			githubToken: "GITHUB_TOKEN",
			wantToken:   "",
			wantSource:  defaultSource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_TOKEN", tt.githubToken)
			t.Setenv("GITHUB_ENTERPRISE_TOKEN", tt.githubEnterpriseToken)
			t.Setenv("GH_TOKEN", tt.ghToken)
			t.Setenv("GH_ENTERPRISE_TOKEN", tt.ghEnterpriseToken)
			token, source := tokenForHost(tt.config, tt.host)
			require.Equal(t, tt.wantToken, token, "Expected token for \"%s\" to be \"%s\", got \"%s\"", tt.host, tt.wantToken, token)
			require.Equal(t, tt.wantSource, source, "Expected source for \"%s\" to be \"%s\", got \"%s\"", tt.host, tt.wantSource, source)
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
			name:    "github API",
			host:    "api.github.com",
			wantOut: false,
		},
		{
			name:    "localhost",
			host:    "github.localhost",
			wantOut: false,
		},
		{
			name:    "localhost API",
			host:    "api.github.localhost",
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
		{
			name:    "tenant API",
			host:    "api.tenant.ghe.com",
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
			name:    "github API",
			host:    "api.github.com",
			wantOut: false,
		},
		{
			name:    "localhost",
			host:    "github.localhost",
			wantOut: false,
		},
		{
			name:    "localhost API",
			host:    "api.github.localhost",
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
		{
			name:    "tenant API",
			host:    "api.tenant.ghe.com",
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
			normalized := NormalizeHostname(tt.host)
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
