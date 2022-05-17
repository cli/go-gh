// Package yamlmap is a wrapper of gopkg.in/yaml.v3 for interacting
// with yaml data as if it were a map.
package yamlmap

import (
	"errors"

	"gopkg.in/yaml.v3"
)

type Map struct {
	*yaml.Node
}

var ErrNotFound = errors.New("not found")
var ErrInvalidYaml = errors.New("invalid yaml")
var ErrInvalidFormat = errors.New("invalid format")

func NewMap() Map {
	return Map{&yaml.Node{}}
}

func StringValue(value string) Map {
	return Map{&yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: value,
	}}
}

func MapValue() Map {
	return Map{&yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}}
}

func Unmarshal(data []byte) (Map, error) {
	root := NewMap()
	err := yaml.Unmarshal(data, root.Node)
	if err != nil {
		return root, ErrInvalidYaml
	}
	if len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		return root, ErrInvalidFormat
	}
	return Map{root.Content[0]}, nil
}

func Marshal(m Map) ([]byte, error) {
	return yaml.Marshal(m)
}

func (m *Map) AddEntry(key string, value Map) {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: key,
	}

	m.Content = append(m.Content, keyNode, value.Node)
}

func (m *Map) Empty() bool {
	return m.Content == nil || len(m.Content) == 0
}

func (m *Map) FindEntry(key string) (Map, error) {
	if m.Empty() {
		return NewMap(), ErrNotFound
	}

	// Content slice goes [key1, value1, key2, value2, ...].
	topLevelPairs := m.Content
	for i, v := range topLevelPairs {
		// Skip every other slice item since we only want to check against keys.
		if i%2 != 0 {
			continue
		}
		if v.Value == key {
			if i+1 < len(topLevelPairs) {
				return Map{topLevelPairs[i+1]}, nil
			}
		}
	}

	return NewMap(), ErrNotFound
}

func (m *Map) Keys() []string {
	keys := []string{}
	if m.Empty() {
		return keys
	}

	// Content slice goes [key1, value1, key2, value2, ...].
	for i, v := range m.Content {
		// Skip every other slice item since we only want keys.
		if i%2 != 0 {
			continue
		}
		keys = append(keys, v.Value)
	}

	return keys
}

func (m *Map) RemoveEntry(key string) {
	if m.Empty() {
		return
	}

	newContent := []*yaml.Node{}

	var skipNext bool
	for i, v := range m.Content {
		if skipNext {
			skipNext = false
			continue
		}
		if i%2 != 0 || v.Value != key {
			newContent = append(newContent, v)
		} else {
			// Don't append current node and skip the next which is this key's value.
			skipNext = true
		}
	}

	m.Content = newContent
}

func (m *Map) UpdateEntry(key string, value Map) error {
	if m.Empty() {
		return ErrNotFound
	}

	for i, v := range m.Content {
		if i%2 != 0 || v.Value != key {
			continue
		}
		if v.Value == key {
			if i+1 < len(m.Content) {
				m.Content[i+1] = value.Node
				return nil
			}
		}
	}

	return ErrNotFound
}
