package yamlmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapAddEntry(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		value      string
		wantValue  string
		wantLength int
	}{
		{
			name:       "add entry with key that is not present",
			key:        "notPresent",
			value:      "test1",
			wantValue:  "test1",
			wantLength: 10,
		},
		{
			name:       "add entry with key that is already present",
			key:        "erroneous",
			value:      "test2",
			wantValue:  "same",
			wantLength: 10,
		},
	}

	for _, tt := range tests {
		m := testMap()
		t.Run(tt.name, func(t *testing.T) {
			m.AddEntry(tt.key, StringValue(tt.value))
			entry, err := m.FindEntry(tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValue, entry.Value)
			assert.Equal(t, tt.wantLength, len(m.Content))
		})
	}
}

func TestMapEmpty(t *testing.T) {
	m := NewMap()
	assert.Equal(t, true, m.Empty())
	m.AddEntry("test", StringValue("test"))
	assert.Equal(t, false, m.Empty())
}

func TestMapFindEntry(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		output  string
		wantErr bool
	}{
		{
			name:   "find key",
			key:    "valid",
			output: "present",
		},
		{
			name:    "find key that is not present",
			key:     "invalid",
			wantErr: true,
		},
		{
			name:   "find key with blank value",
			key:    "blank",
			output: "",
		},
		{
			name:   "find key that has same content as a value",
			key:    "same",
			output: "logical",
		},
	}

	for _, tt := range tests {
		m := testMap()
		t.Run(tt.name, func(t *testing.T) {
			out, err := m.FindEntry(tt.key)
			if tt.wantErr {
				assert.EqualError(t, err, "not found")
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.output, out.Value)
		})
	}
}

func TestMapFindEntryModified(t *testing.T) {
	m := testMap()
	entry, err := m.FindEntry("valid")
	assert.NoError(t, err)
	assert.Equal(t, "present", entry.Value)
	entry.Value = "test"
	assert.Equal(t, "test", entry.Value)
	entry2, err := m.FindEntry("valid")
	assert.NoError(t, err)
	assert.Equal(t, "test", entry2.Value)
}

func TestMapKeys(t *testing.T) {
	tests := []struct {
		name     string
		m        Map
		wantKeys []string
	}{
		{
			name:     "keys for full map",
			m:        testMap(),
			wantKeys: []string{"valid", "erroneous", "blank", "same"},
		},
		{
			name:     "keys for empty map",
			m:        NewMap(),
			wantKeys: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := tt.m.Keys()
			assert.Equal(t, tt.wantKeys, keys)
		})
	}
}

func TestMapRemoveEntry(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		wantLength int
	}{
		{
			name:       "remove key",
			key:        "erroneous",
			wantLength: 6,
		},
		{
			name:       "remove key that is not present",
			key:        "invalid",
			wantLength: 8,
		},
		{
			name:       "remove key that has same content as a value",
			key:        "same",
			wantLength: 6,
		},
	}

	for _, tt := range tests {
		m := testMap()
		t.Run(tt.name, func(t *testing.T) {
			m.RemoveEntry(tt.key)
			assert.Equal(t, tt.wantLength, len(m.Content))
			_, err := m.FindEntry(tt.key)
			assert.EqualError(t, err, "not found")
		})
	}
}

func TestMapUpdateEntry(t *testing.T) {
	//TODO: Write tests.
}

func testMap() Map {
	var data = `
valid: present
erroneous: same
blank:
same: logical
`
	m, _ := Unmarshal([]byte(data))
	return m
}
