package config

import (
	"sort"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh/internal/yamlmap"
)

type GeneralEntries interface {
	All() []Value
	Dirty() bool
	Get(key string) Value
	Set(key string, value string) error
	String() string
}

type generalEntries struct {
	entries yamlmap.Map
	dirty   bool
}

type generalEntriesOption struct {
	Key           string
	Description   string
	DefaultValue  string
	AllowedValues []string
}

var generalEntriesOptions = map[string]generalEntriesOption{
	"browser": {
		Key:          "browser",
		Description:  "the web browser to use for opening URLs",
		DefaultValue: "",
	},
	"editor": {
		Key:          "editor",
		Description:  "the text editor program to use for authoring text",
		DefaultValue: "",
	},
	"git_protocol": {
		Key:           "git_protocol",
		Description:   "the protocol to use for git clone and push operations",
		DefaultValue:  "https",
		AllowedValues: []string{"https", "ssh"},
	},
	"http_unix_socket": {
		Key:          "http_unix_socket",
		Description:  "the path to a Unix socket through which to make an HTTP connection",
		DefaultValue: "",
	},
	"pager": {
		Key:          "pager",
		Description:  "the terminal pager program to send standard output to",
		DefaultValue: "",
	},
	"prompt": {
		Key:           "prompt",
		Description:   "toggle interactive prompting in the terminal",
		DefaultValue:  "enabled",
		AllowedValues: []string{"enabled", "disabled"},
	},
}

var defaultGeneralEntries = heredoc.Doc(`
# What protocol to use when performing git operations. Supported values: ssh, https
git_protocol: https
# What editor gh should run when creating issues, pull requests, etc. If blank, will refer to environment.
editor:
# When to interactively prompt. This is a global config that cannot be overridden by hostname. Supported values: enabled, disabled
prompt: enabled
# A pager program to send command output to, e.g. "less". Set the value to "cat" to disable the pager.
pager:
# Aliases allow you to create nicknames for gh commands
aliases:
  co: pr checkout
# The path to a unix socket through which send HTTP connections. If blank, HTTP traffic will be handled by net/http.DefaultTransport.
http_unix_socket:
# What web browser gh should use when opening URLs. If blank, will refer to environment.
browser:
`)

func (c *generalEntries) All() []Value {
	vs := []Value{}
	keys := make([]string, 0, len(generalEntriesOptions))
	for key := range generalEntriesOptions {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		vs = append(vs, c.Get(key))
	}
	return vs
}

func (c *generalEntries) Dirty() bool {
	return c.dirty
}

func (c *generalEntries) Get(key string) Value {
	key = normalizeKey(key)
	val := generalEntriesOptions[key].DefaultValue
	m, err := c.entries.FindEntry(key)
	if err == nil {
		val = m.Value
	}
	return &value{
		source:   key,
		value:    val,
		notFound: err != nil,
	}
}

func (c *generalEntries) Set(key, value string) error {
	key = normalizeKey(key)
	value = normalizeValue(value)
	err := validateGeneralEntry(key, value)
	if err != nil {
		return err
	}
	c.dirty = true
	entry, err := c.entries.FindEntry(key)
	if err == nil {
		entry.Value = value
		return nil
	}
	c.entries.AddEntry(key, yamlmap.StringValue(value))
	return nil
}

func (c *generalEntries) String() string {
	data, err := yamlmap.Marshal(c.entries)
	if err != nil {
		return ""
	}
	return string(data)
}

func normalizeKey(key string) string {
	return strings.ToLower(key)
}

func normalizeValue(value string) string {
	return strings.ToLower(value)
}

func validateGeneralEntry(key, value string) error {
	if option, ok := generalEntriesOptions[key]; ok {
		if option.AllowedValues == nil {
			return nil
		}
		for _, aValue := range option.AllowedValues {
			if aValue == value {
				return nil
			}
		}
		return SetInvalidValueError{Key: key, Value: value}
	} else {
		return SetInvalidKeyError{Key: key, Value: value}
	}
}
