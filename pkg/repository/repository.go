// Package repository is a set of types and functions for modeling and
// interacting with GitHub repositories.
package repository

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cli/go-gh/v2/internal/git"
	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/cli/go-gh/v2/pkg/ssh"
)

// Repository holds information representing a GitHub repository.
type Repository struct {
	Host  string
	Name  string
	Owner string
}

// Parse extracts the repository information from the following
// string formats: "OWNER/REPO", "HOST/OWNER/REPO", and a full URL.
// If the format does not specify a host, use the config to determine a host.
func Parse(s string) (Repository, error) {
	var r Repository

	if git.IsURL(s) {
		u, err := git.ParseURL(s)
		if err != nil {
			return r, err
		}

		host, owner, name, err := git.RepoInfoFromURL(u)
		if err != nil {
			return r, err
		}

		r.Host = host
		r.Name = name
		r.Owner = owner

		return r, nil
	}

	parts := strings.SplitN(s, "/", 4)
	for _, p := range parts {
		if len(p) == 0 {
			return r, fmt.Errorf(`expected the "[HOST/]OWNER/REPO" format, got %q`, s)
		}
	}

	switch len(parts) {
	case 3:
		r.Host = parts[0]
		r.Owner = parts[1]
		r.Name = parts[2]
		return r, nil
	case 2:
		r.Host, _ = auth.DefaultHost()
		r.Owner = parts[0]
		r.Name = parts[1]
		return r, nil
	default:
		return r, fmt.Errorf(`expected the "[HOST/]OWNER/REPO" format, got %q`, s)
	}
}

// ParseWithHost extracts the repository information from the following
// string formats: "OWNER/REPO", "HOST/OWNER/REPO", and a full URL.
// If the format does not specify a host, use the host provided.
func ParseWithHost(s, host string) (Repository, error) {
	var r Repository

	if git.IsURL(s) {
		u, err := git.ParseURL(s)
		if err != nil {
			return r, err
		}

		host, owner, name, err := git.RepoInfoFromURL(u)
		if err != nil {
			return r, err
		}

		r.Host = host
		r.Owner = owner
		r.Name = name

		return r, nil
	}

	parts := strings.SplitN(s, "/", 4)
	for _, p := range parts {
		if len(p) == 0 {
			return r, fmt.Errorf(`expected the "[HOST/]OWNER/REPO" format, got %q`, s)
		}
	}

	switch len(parts) {
	case 3:
		r.Host = parts[0]
		r.Owner = parts[1]
		r.Name = parts[2]
		return r, nil
	case 2:
		r.Host = host
		r.Owner = parts[0]
		r.Name = parts[1]
		return r, nil
	default:
		return r, fmt.Errorf(`expected the "[HOST/]OWNER/REPO" format, got %q`, s)
	}
}

// Current uses git remotes to determine the GitHub repository
// the current directory is tracking.
func Current() (Repository, error) {
	var r Repository

	override := os.Getenv("GH_REPO")
	if override != "" {
		return Parse(override)
	}

	remotes, err := git.Remotes()
	if err != nil {
		return r, err
	}
	if len(remotes) == 0 {
		return r, errors.New("unable to determine current repository, no git remotes configured for this repository")
	}

	translator := ssh.NewTranslator()
	for _, r := range remotes {
		if r.FetchURL != nil {
			r.FetchURL = translator.Translate(r.FetchURL)
		}
		if r.PushURL != nil {
			r.PushURL = translator.Translate(r.PushURL)
		}
	}

	hosts := auth.KnownHosts()

	filteredRemotes := remotes.FilterByHosts(hosts)
	if len(filteredRemotes) == 0 {
		return r, errors.New("unable to determine current repository, none of the git remotes configured for this repository point to a known GitHub host")
	}

	rem := filteredRemotes[0]
	r.Host = rem.Host
	r.Owner = rem.Owner
	r.Name = rem.Repo

	return r, nil
}
