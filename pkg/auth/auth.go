package auth

import (
	"errors"
	"os"
	"strings"

	"github.com/cli/go-gh/pkg/config"
)

const (
	GH_HOST                 = "GH_HOST"
	GH_TOKEN                = "GH_TOKEN"
	GITHUB_TOKEN            = "GITHUB_TOKEN"
	GH_ENTERPRISE_TOKEN     = "GH_ENTERPRISE_TOKEN"
	GITHUB_ENTERPRISE_TOKEN = "GITHUB_ENTERPRISE_TOKEN"
	defaultHostname         = "github.com"
	oauthToken              = "oauth_token"
)

type NotFoundError struct {
	error
}

func Token(host string) (string, error) {
	return token(host, config.Load)
}

func token(host string, configFunc func() (config.Config, error)) (string, error) {
	hostname := normalizeHostname(host)
	if isEnterprise(hostname) {
		if token := os.Getenv(GH_ENTERPRISE_TOKEN); token != "" {
			return token, nil
		}
		if token := os.Getenv(GITHUB_ENTERPRISE_TOKEN); token != "" {
			return token, nil
		}
		if configFunc == nil {
			return "", NotFoundError{errors.New("not found")}
		}
		cfg, err := configFunc()
		if err != nil {
			return "", NotFoundError{errors.New("not found")}
		}
		if token, err := cfg.GetForHost(hostname, oauthToken); err == nil {
			return token, nil
		}
		return "", NotFoundError{errors.New("not found")}
	}

	if token := os.Getenv(GH_TOKEN); token != "" {
		return token, nil
	}
	if token := os.Getenv(GITHUB_TOKEN); token != "" {
		return token, nil
	}
	if configFunc == nil {
		return "", NotFoundError{errors.New("not found")}
	}
	cfg, err := configFunc()
	if err != nil {
		return "", NotFoundError{errors.New("not found")}
	}
	if token, err := cfg.GetForHost(hostname, oauthToken); err == nil {
		return token, nil
	}
	return "", NotFoundError{errors.New("not found")}
}

func isEnterprise(host string) bool {
	return host != defaultHostname
}

func normalizeHostname(host string) string {
	hostname := strings.ToLower(host)
	if strings.HasSuffix(hostname, "."+defaultHostname) {
		return defaultHostname
	}
	return hostname
}
