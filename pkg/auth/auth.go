package auth

import (
	"os"
	"strings"

	"github.com/cli/go-gh/internal/set"
	"github.com/cli/go-gh/pkg/config"
)

const (
	github                = "github.com"
	ghEnterpriseToken     = "GH_ENTERPRISE_TOKEN"
	ghHost                = "GH_HOST"
	ghToken               = "GH_TOKEN"
	githubEnterpriseToken = "GITHUB_ENTERPRISE_TOKEN"
	githubToken           = "GITHUB_TOKEN"
	oauthToken            = "oauth_token"
	hostsKey              = "hosts"
)

func TokenForHost(host string) (string, string) {
	cfg, _ := config.Read()
	return tokenForHost(cfg, host)
}

func tokenForHost(cfg *config.Config, host string) (string, string) {
	host = normalizeHostname(host)
	if isEnterprise(host) {
		if token := os.Getenv(ghEnterpriseToken); token != "" {
			return token, ghEnterpriseToken
		}
		if token := os.Getenv(githubEnterpriseToken); token != "" {
			return token, githubEnterpriseToken
		}
		if cfg != nil {
			token, _ := config.Get(cfg, []string{hostsKey, host, oauthToken})
			return token, oauthToken
		}
	}
	if token := os.Getenv(ghToken); token != "" {
		return token, ghToken
	}
	if token := os.Getenv(githubToken); token != "" {
		return token, githubToken
	}
	if cfg != nil {
		token, _ := config.Get(cfg, []string{hostsKey, host, oauthToken})
		return token, oauthToken
	}
	return "", ""
}

func KnownHosts() []string {
	cfg, _ := config.Read()
	return knownHosts(cfg)
}

func knownHosts(cfg *config.Config) []string {
	hosts := set.NewStringSet()
	if host := os.Getenv(ghHost); host != "" {
		hosts.Add(host)
	}
	if token, _ := tokenForHost(cfg, github); token != "" {
		hosts.Add(github)
	}
	if cfg != nil {
		keys, err := config.Keys(cfg, []string{hostsKey})
		if err == nil {
			hosts.AddValues(keys)
		}
	}
	return hosts.ToSlice()
}

func DefaultHost() (string, string) {
	cfg, _ := config.Read()
	return defaultHost(cfg)
}

func defaultHost(cfg *config.Config) (string, string) {
	if host := os.Getenv(ghHost); host != "" {
		return host, ghHost
	}
	if cfg != nil {
		keys, err := config.Keys(cfg, []string{hostsKey})
		if err == nil && len(keys) == 1 {
			return keys[0], hostsKey
		}
	}
	return github, "default"
}

func isEnterprise(host string) bool {
	return host != github
}

func normalizeHostname(host string) string {
	hostname := strings.ToLower(host)
	if strings.HasSuffix(hostname, "."+github) {
		return github
	}
	return hostname
}
