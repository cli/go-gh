package config

import (
	"os"
	"strings"

	"github.com/cli/go-gh/internal/set"
	"github.com/cli/go-gh/internal/yamlmap"
)

const (
	defaultHost           = "github.com"
	ghEnterpriseToken     = "GH_ENTERPRISE_TOKEN"
	ghHost                = "GH_HOST"
	ghToken               = "GH_TOKEN"
	githubEnterpriseToken = "GITHUB_ENTERPRISE_TOKEN"
	githubToken           = "GITHUB_TOKEN"
	oauthToken            = "oauth_token"
)

type HostEntries interface {
	AuthToken(host string) Value
	DefaultHost() Value
	Dirty() bool
	Get(host string, key string) Value
	Keys() []string
	Remove(host string)
	Set(host string, key string, value string) error
	String() string
}

type hostEntries struct {
	entries yamlmap.Map
	dirty   bool
}

type hostOption struct {
	Key           string
	AllowedValues []string
}

var hostOptions = map[string]hostOption{
	"git_protocol": {Key: "git_protocol", AllowedValues: []string{"https", "ssh"}},
	"oauth_token":  {Key: "oauth_token"},
	"user":         {Key: "user"},
}

func (c *hostEntries) Dirty() bool {
	return c.dirty
}

func (c *hostEntries) Keys() []string {
	hosts := set.NewStringSet()
	if host := os.Getenv(ghHost); host != "" {
		hosts.Add(host)
	}
	if token := c.AuthToken(defaultHost); !token.NotFound() {
		hosts.Add(defaultHost)
	}
	entries := c.entries.Keys()
	hosts.AddValues(entries)
	return hosts.ToSlice()
}

func (c *hostEntries) AuthToken(host string) Value {
	host = normalizeHostname(host)
	v := &value{}
	if isEnterprise(host) {
		if token := os.Getenv(ghEnterpriseToken); token != "" {
			v.source = ghEnterpriseToken
			v.value = token
			return v
		}
		if token := os.Getenv(githubEnterpriseToken); token != "" {
			v.source = githubEnterpriseToken
			v.value = token
			return v
		}
		return c.Get(host, oauthToken)
	}
	if token := os.Getenv(ghToken); token != "" {
		v.source = ghToken
		v.value = token
		return v
	}
	if token := os.Getenv(githubToken); token != "" {
		v.source = githubToken
		v.value = token
		return v
	}
	return c.Get(host, oauthToken)
}

func (c *hostEntries) DefaultHost() Value {
	v := &value{}
	if host := os.Getenv(ghHost); host != "" {
		v.source = ghHost
		v.value = host
		return v
	}
	keys := c.entries.Keys()
	if len(keys) == 1 {
		//TODO: What should the source be here?
		v.source = "host"
		v.value = keys[0]
		return v
	}
	//TODO: What should the source be here? What should notFound be here?
	v.notFound = true
	v.value = defaultHost
	return v
}

func (c *hostEntries) Get(host, key string) Value {
	host = normalizeHostname(host)
	key = normalizeKey(key)
	hostMap, err := c.entries.FindEntry(host)
	if err != nil {
		return &value{notFound: true}
	}
	var val string
	m, err := hostMap.FindEntry(key)
	if err == nil {
		val = m.Value
	}
	return &value{
		source:   key,
		value:    val,
		notFound: err != nil,
	}
}

func (c *hostEntries) Remove(host string) {
	host = normalizeHostname(host)
	c.dirty = true
	c.entries.RemoveEntry(host)
}

func (c *hostEntries) Set(host, key, value string) error {
	host = normalizeHostname(host)
	key = normalizeKey(key)
	err := validateHostEntry(host, key, value)
	if err != nil {
		return err
	}
	c.dirty = true
	hostMap, err := c.entries.FindEntry(host)
	if err != nil {
		hostMap := yamlmap.MapValue()
		hostMap.AddEntry(key, yamlmap.StringValue(value))
		c.entries.AddEntry(host, hostMap)
		return nil
	}
	m, err := hostMap.FindEntry(key)
	if err == nil {
		m.Value = value
		return nil
	}
	hostMap.AddEntry(key, yamlmap.StringValue(value))
	return nil
}

func (c *hostEntries) String() string {
	data, err := yamlmap.Marshal(c.entries)
	if err != nil {
		return ""
	}
	return string(data)
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

func validateHostEntry(host, key, value string) error {
	if option, ok := hostOptions[key]; ok {
		if option.AllowedValues == nil {
			return nil
		}
		for _, aValue := range option.AllowedValues {
			if aValue == value {
				return nil
			}
		}
		return SetInvalidValueError{Host: host, Key: key, Value: value}
	} else {
		return SetInvalidKeyError{Host: host, Key: key, Value: value}
	}
}
