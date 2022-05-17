package config

import (
	"sort"

	"github.com/cli/go-gh/internal/yamlmap"
)

type AliasEntries interface {
	Add(key string, value string)
	All() []Value
	Dirty() bool
	Get(key string) Value
	Remove(key string)
	String() string
}

type aliasEntries struct {
	entries yamlmap.Map
	dirty   bool
}

func (c *aliasEntries) Add(key, value string) {
	c.dirty = true
	c.entries.AddEntry(key, yamlmap.StringValue(value))
}

func (c *aliasEntries) All() []Value {
	vs := []Value{}
	keys := c.entries.Keys()
	sort.Strings(keys)
	for _, key := range c.entries.Keys() {
		vs = append(vs, c.Get(key))
	}
	return vs
}

func (c *aliasEntries) Dirty() bool {
	return c.dirty
}

func (c *aliasEntries) Get(key string) Value {
	var val string
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

func (c *aliasEntries) Remove(key string) {
	c.dirty = true
	c.entries.RemoveEntry(key)
}

func (c *aliasEntries) String() string {
	data, err := yamlmap.Marshal(c.entries)
	if err != nil {
		return ""
	}
	return string(data)
}
