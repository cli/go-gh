package git

import (
	"net/url"
	"strings"
)

type RemoteSet []*Remote

type Remote struct {
	Name     string
	FetchURL *url.URL
	PushURL  *url.URL
	Resolved string
	Host     string
	Owner    string
	Repo     string
}

func (r *Remote) String() string {
	return r.Name
}

func (r RemoteSet) Len() int      { return len(r) }
func (r RemoteSet) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r RemoteSet) Less(i, j int) bool {
	return remoteNameSortScore(r[i].Name) > remoteNameSortScore(r[j].Name)
}

func remoteNameSortScore(name string) int {
	switch strings.ToLower(name) {
	case "upstream":
		return 3
	case "github":
		return 2
	case "origin":
		return 1
	default:
		return 0
	}
}

// Filter remotes by given hostnames, maintains original order.
func (rs RemoteSet) FilterByHosts(hosts []string) RemoteSet {
	filtered := make(RemoteSet, 0)
	for _, remote := range rs {
		for _, host := range hosts {
			if strings.EqualFold(remote.Host, host) {
				filtered = append(filtered, remote)
				break
			}
		}
	}
	return filtered
}
