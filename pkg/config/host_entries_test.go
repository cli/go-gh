package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostEntriesAuthToken(t *testing.T) {
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
		name                    string
		host                    string
		GITHUB_TOKEN            string
		GITHUB_ENTERPRISE_TOKEN string
		GH_TOKEN                string
		GH_ENTERPRISE_TOKEN     string
		entries                 HostEntries
		wantToken               string
		wantSource              string
		wantNotFound            bool
	}{
		{
			name:         "token for github.com with no env tokens and no config token",
			host:         "github.com",
			entries:      testNoHostEntries(),
			wantToken:    "",
			wantSource:   "default",
			wantNotFound: true,
		},
		{
			name:         "token for enterprise.com with no env tokens and no config token",
			host:         "enterprise.com",
			entries:      testNoHostEntries(),
			wantToken:    "",
			wantSource:   "default",
			wantNotFound: true,
		},
		{
			name:         "token for github.com with GH_TOKEN, GITHUB_TOKEN, and config token",
			host:         "github.com",
			GH_TOKEN:     "GH_TOKEN",
			GITHUB_TOKEN: "GITHUB_TOKEN",
			entries:      testHostEntries(),
			wantToken:    "GH_TOKEN",
			wantSource:   "GH_TOKEN",
		},
		{
			name:         "token for github.com with GITHUB_TOKEN, and config token",
			host:         "github.com",
			GITHUB_TOKEN: "GITHUB_TOKEN",
			entries:      testHostEntries(),
			wantToken:    "GITHUB_TOKEN",
			wantSource:   "GITHUB_TOKEN",
		},
		{
			name:       "token for github.com with config token",
			host:       "github.com",
			entries:    testHostEntries(),
			wantToken:  "xxxxxxxxxxxxxxxxxxxx",
			wantSource: "oauth_token",
		},
		{
			name:                    "token for enterprise.com with GH_ENTERPRISE_TOKEN, GITHUB_ENTERPRISE_TOKEN, and config token",
			host:                    "enterprise.com",
			GH_ENTERPRISE_TOKEN:     "GH_ENTERPRISE_TOKEN",
			GITHUB_ENTERPRISE_TOKEN: "GITHUB_ENTERPRISE_TOKEN",
			entries:                 testHostEntries(),
			wantToken:               "GH_ENTERPRISE_TOKEN",
			wantSource:              "GH_ENTERPRISE_TOKEN",
		},
		{
			name:                    "token for enterprise.com with GITHUB_ENTERPRISE_TOKEN, and config token",
			host:                    "enterprise.com",
			GITHUB_ENTERPRISE_TOKEN: "GITHUB_ENTERPRISE_TOKEN",
			entries:                 testHostEntries(),
			wantToken:               "GITHUB_ENTERPRISE_TOKEN",
			wantSource:              "GITHUB_ENTERPRISE_TOKEN",
		},
		{
			name:       "token for enterprise.com with config token",
			host:       "enterprise.com",
			entries:    testHostEntries(),
			wantToken:  "yyyyyyyyyyyyyyyyyyyy",
			wantSource: "oauth_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GITHUB_TOKEN", tt.GITHUB_TOKEN)
			os.Setenv("GITHUB_ENTERPRISE_TOKEN", tt.GITHUB_ENTERPRISE_TOKEN)
			os.Setenv("GH_TOKEN", tt.GH_TOKEN)
			os.Setenv("GH_ENTERPRISE_TOKEN", tt.GH_ENTERPRISE_TOKEN)
			token := tt.entries.AuthToken(tt.host)
			assert.Equal(t, tt.wantToken, token.Value())
			assert.Equal(t, tt.wantSource, token.Source())
			assert.Equal(t, tt.wantNotFound, token.NotFound())
		})
	}
}

func TestHostEntriesDefaultHost(t *testing.T) {
	tests := []struct {
		name         string
		entries      HostEntries
		ghHost       string
		wantHost     string
		wantSource   string
		wantNotFound bool
	}{
		{
			name:       "GH_HOST if set",
			entries:    testHostEntries(),
			ghHost:     "test.com",
			wantHost:   "test.com",
			wantSource: "GH_HOST",
		},
		{
			name:       "authenticated host if only one",
			entries:    testSingleHostEntry(),
			wantHost:   "enterprise.com",
			wantSource: "host",
		},
		{
			name:         "default host if more than one authenticated host",
			entries:      testHostEntries(),
			wantHost:     "github.com",
			wantSource:   "default",
			wantNotFound: true,
		},
		{
			name:         "default host if no authenticated host",
			entries:      testNoHostEntries(),
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
			host := tt.entries.DefaultHost()
			assert.Equal(t, tt.wantHost, host.Value())
			assert.Equal(t, tt.wantSource, host.Source())
			assert.Equal(t, tt.wantNotFound, host.NotFound())
		})
	}
}

func TestHostEntriesDirty(t *testing.T) {
	//TODO: Write tests.
}

func TestHostEntriesGet(t *testing.T) {
	entries := testHostEntries()

	tests := []struct {
		name         string
		host         string
		key          string
		wantValue    string
		wantSource   string
		wantNotFound bool
	}{
		{
			name:       "get github user value",
			host:       "github.com",
			key:        "user",
			wantValue:  "user1",
			wantSource: "user",
		},
		{
			name:       "get github oauth_token value",
			host:       "github.com",
			key:        "oauth_token",
			wantValue:  "xxxxxxxxxxxxxxxxxxxx",
			wantSource: "oauth_token",
		},
		{
			name:       "get github git_protocol value",
			host:       "github.com",
			key:        "git_protocol",
			wantValue:  "ssh",
			wantSource: "git_protocol",
		},
		{
			name:       "get enterprise user value",
			host:       "enterprise.com",
			key:        "user",
			wantValue:  "user2",
			wantSource: "user",
		},
		{
			name:       "get enterprise oauth_token value",
			host:       "enterprise.com",
			key:        "oauth_token",
			wantValue:  "yyyyyyyyyyyyyyyyyyyy",
			wantSource: "oauth_token",
		},
		{
			name:       "get enterprise git_protocol value",
			host:       "enterprise.com",
			key:        "git_protocol",
			wantValue:  "https",
			wantSource: "git_protocol",
		},
		{
			name:         "unknown host",
			host:         "unknown",
			key:          "user",
			wantValue:    "",
			wantSource:   "default",
			wantNotFound: true,
		},
		{
			name:         "unknown key",
			host:         "github.com",
			key:          "unknown",
			wantValue:    "",
			wantSource:   "default",
			wantNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := entries.Get(tt.host, tt.key)
			assert.Equal(t, tt.wantValue, entry.Value())
			assert.Equal(t, tt.wantSource, entry.Source())
			assert.Equal(t, tt.wantNotFound, entry.NotFound())
		})
	}
}

func TestHostEntriesKeys(t *testing.T) {
	tests := []struct {
		name      string
		entries   HostEntries
		ghHost    string
		ghToken   string
		wantHosts []string
	}{
		{
			name:      "no known hosts",
			entries:   testNoHostEntries(),
			wantHosts: []string{},
		},
		{
			name:      "includes GH_HOST",
			entries:   testNoHostEntries(),
			ghHost:    "test.com",
			wantHosts: []string{"test.com"},
		},
		{
			name:      "includes authenticated hosts",
			entries:   testHostEntries(),
			wantHosts: []string{"github.com", "enterprise.com"},
		},
		{
			name:      "includes default host if environment auth token",
			entries:   testNoHostEntries(),
			ghToken:   "TOKEN",
			wantHosts: []string{"github.com"},
		},
		{
			name:      "deduplicates hosts",
			entries:   testHostEntries(),
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
			hosts := tt.entries.Keys()
			assert.Equal(t, tt.wantHosts, hosts)
		})
	}
}

func TestHostEntriesRemove(t *testing.T) {
	//TODO: Write tests.
}

func TestHostEntriesSet(t *testing.T) {
	//TODO: Write tests.
}

func TestHostEntriesString(t *testing.T) {
	//TODO: Write tests.
}

func testNoHostEntries() HostEntries {
	var data = ``
	cfg := ReadFromString(data)
	return cfg.Hosts()
}

func testSingleHostEntry() HostEntries {
	var data = `
hosts:
  enterprise.com:
    user: user2
    oauth_token: yyyyyyyyyyyyyyyyyyyy
    git_protocol: https
`
	cfg := ReadFromString(data)
	return cfg.Hosts()
}

func testHostEntries() HostEntries {
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
	cfg := ReadFromString(data)
	return cfg.Hosts()
}
