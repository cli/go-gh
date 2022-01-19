package repository

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		hostOverride string
		wantOwner    string
		wantName     string
		wantHost     string
		wantErr      string
	}{
		{
			name:      "OWNER/REPO combo",
			input:     "OWNER/REPO",
			wantHost:  "github.com",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:    "too few elements",
			input:   "OWNER",
			wantErr: `expected the "[HOST/]OWNER/REPO" format, got "OWNER"`,
		},
		{
			name:    "too many elements",
			input:   "a/b/c/d",
			wantErr: `expected the "[HOST/]OWNER/REPO" format, got "a/b/c/d"`,
		},
		{
			name:    "blank value",
			input:   "a/",
			wantErr: `expected the "[HOST/]OWNER/REPO" format, got "a/"`,
		},
		{
			name:      "with hostname",
			input:     "example.org/OWNER/REPO",
			wantHost:  "example.org",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:      "full URL",
			input:     "https://example.org/OWNER/REPO.git",
			wantHost:  "example.org",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:      "SSH URL",
			input:     "git@example.org:OWNER/REPO.git",
			wantHost:  "example.org",
			wantOwner: "OWNER",
			wantName:  "REPO",
		},
		{
			name:         "OWNER/REPO with default host override",
			input:        "OWNER/REPO",
			hostOverride: "override.com",
			wantHost:     "override.com",
			wantOwner:    "OWNER",
			wantName:     "REPO",
		},
		{
			name:         "HOST/OWNER/REPO with default host override",
			input:        "example.com/OWNER/REPO",
			hostOverride: "override.com",
			wantHost:     "example.com",
			wantOwner:    "OWNER",
			wantName:     "REPO",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.hostOverride != "" {
				old := os.Getenv("GH_HOST")
				os.Setenv("GH_HOST", tt.hostOverride)
				defer os.Setenv("GH_HOST", old)
			}
			r, err := Parse(tt.input)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantHost, r.Host())
			assert.Equal(t, tt.wantOwner, r.Owner())
			assert.Equal(t, tt.wantName, r.Name())
		})
	}
}
