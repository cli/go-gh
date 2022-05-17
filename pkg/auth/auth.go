package auth

import (
	"os"
	"strings"

	"github.com/cli/go-gh/internal/set"
	"github.com/cli/go-gh/pkg/config"
)

const (
	defaultHost           = "github.com"
	ghEnterpriseToken     = "GH_ENTERPRISE_TOKEN"
	ghHost                = "GH_HOST"
	ghToken               = "GH_TOKEN"
	githubEnterpriseToken = "GITHUB_ENTERPRISE_TOKEN"
	githubToken           = "GITHUB_TOKEN"
	oauthToken            = "oauth_token"
	hostsKey              = "hosts"
)

func TokenForHost(cfg *config.Config, host string) (string, string) {
	host = normalizeHostname(host)
	if isEnterprise(host) {
		if token := os.Getenv(ghEnterpriseToken); token != "" {
			return token, ghEnterpriseToken
		}
		if token := os.Getenv(githubEnterpriseToken); token != "" {
			return token, githubEnterpriseToken
		}
		token, _ := config.Get(cfg, []string{hostsKey, host, oauthToken})
		return token, oauthToken
	}
	if token := os.Getenv(ghToken); token != "" {
		return token, ghToken
	}
	if token := os.Getenv(githubToken); token != "" {
		return token, githubToken
	}
	token, _ := config.Get(cfg, []string{hostsKey, host, oauthToken})
	return token, oauthToken
}

func KnownHosts(cfg *config.Config) []string {
	hosts := set.NewStringSet()
	if host := os.Getenv(ghHost); host != "" {
		hosts.Add(host)
	}
	if token, _ := TokenForHost(cfg, defaultHost); token != "" {
		hosts.Add(defaultHost)
	}
	keys, err := config.Keys(cfg, []string{hostsKey})
	if err == nil {
		hosts.AddValues(keys)
	}
	return hosts.ToSlice()
}

func DefaultHost(cfg *config.Config) (string, string) {
	if host := os.Getenv(ghHost); host != "" {
		return host, ghHost
	}
	keys, err := config.Keys(cfg, []string{hostsKey})
	if err == nil && len(keys) == 1 {
		return keys[0], hostsKey
	}
	return defaultHost, "default"
}

func isEnterprise(host string) bool {
	return host != defaultHost
}

func normalizeHostname(host string) string {
	hostname := strings.ToLower(host)
	if strings.HasSuffix(hostname, "."+defaultHost) {
		return defaultHost
	}
	return hostname
}
