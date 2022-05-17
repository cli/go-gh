package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneralEntriesAll(t *testing.T) {
	//TODO: Write tests.
}

func TestGeneralEntriesDirty(t *testing.T) {
	//TODO: Write tests.
}

func TestGeneralEntriesGet(t *testing.T) {
	entries := testGeneralEntries()

	tests := []struct {
		name         string
		key          string
		wantValue    string
		wantSource   string
		wantNotFound bool
	}{
		{
			name:       "get git_protocol value",
			key:        "git_protocol",
			wantValue:  "ssh",
			wantSource: "git_protocol",
		},
		{
			name:       "get editor value",
			key:        "editor",
			wantValue:  "",
			wantSource: "editor",
		},
		{
			name:       "get prompt value",
			key:        "prompt",
			wantValue:  "enabled",
			wantSource: "prompt",
		},
		{
			name:       "get pager value",
			key:        "pager",
			wantValue:  "less",
			wantSource: "pager",
		},
		{
			name:         "unknown key",
			key:          "unknown",
			wantNotFound: true,
			wantValue:    "",
			wantSource:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := entries.Get(tt.key)
			assert.Equal(t, tt.wantValue, value.Value())
			assert.Equal(t, tt.wantSource, value.Source())
			assert.Equal(t, tt.wantNotFound, value.NotFound())
		})
	}
}

func TestGeneralEntriesSet(t *testing.T) {
	//TODO: Write tests.
}

func TestGeneralEntriesString(t *testing.T) {
	//TODO: Write tests.
}

func TestDefaultGeneralEntries(t *testing.T) {
	cfg := ReadFromString(defaultGeneralEntries)
	entries := cfg.General()

	git_protocol := entries.Get("git_protocol")
	assert.Equal(t, "https", git_protocol.Value())
	assert.Equal(t, "git_protocol", git_protocol.Source())
	assert.False(t, git_protocol.NotFound())

	editor := entries.Get("editor")
	assert.Equal(t, "", editor.Value())
	assert.Equal(t, "editor", editor.Source())
	assert.False(t, editor.NotFound())

	prompt := entries.Get("prompt")
	assert.Equal(t, "enabled", prompt.Value())
	assert.Equal(t, "prompt", prompt.Source())
	assert.False(t, prompt.NotFound())

	pager := entries.Get("pager")
	assert.Equal(t, "", pager.Value())
	assert.Equal(t, "pager", pager.Source())
	assert.False(t, pager.NotFound())

	unix_socket := entries.Get("http_unix_socket")
	assert.Equal(t, "", unix_socket.Value())
	assert.Equal(t, "http_unix_socket", unix_socket.Source())
	assert.False(t, unix_socket.NotFound())

	browser := entries.Get("browser")
	assert.Equal(t, "", browser.Value())
	assert.Equal(t, "browser", browser.Source())
	assert.False(t, browser.NotFound())

	unknown := entries.Get("unknown")
	assert.Equal(t, "", unknown.Value())
	assert.Equal(t, "default", unknown.Source())
	assert.True(t, unknown.NotFound())
}

func testGeneralEntries() GeneralEntries {
	var data = `
git_protocol: ssh
editor:
prompt: enabled
pager: less
`
	cfg := ReadFromString(data)
	return cfg.General()
}
