package api

import (
	"testing"

	"github.com/cli/go-gh/v2/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestAPIHost(t *testing.T) {
	tests := []struct {
		name   string
		host   string
		config string
		want   string
		wantOK bool
	}{
		{
			name: "missing api_host",
			host: "example.ghe.com",
			config: `
hosts:
  example.ghe.com:
    oauth_token: token
`,
		},
		{
			name: "null api_host",
			host: "example.ghe.com",
			config: `
hosts:
  example.ghe.com:
    api_host:
`,
		},
		{
			name: "empty api_host",
			host: "example.ghe.com",
			config: `
hosts:
  example.ghe.com:
    api_host: ""
`,
		},
		{
			name: "host absent from config",
			host: "other.ghe.com",
			config: `
hosts:
  example.ghe.com:
    api_host: gw.example.net
`,
		},
		{
			name: "configured value",
			host: "example.ghe.com",
			config: `
hosts:
  example.ghe.com:
    api_host: gw.example.net
`,
			want:   "gw.example.net",
			wantOK: true,
		},
		{
			name: "normalizes canonical host for lookup",
			host: "Example.ghe.com",
			config: `
hosts:
  example.ghe.com:
    api_host: GW.Example.NET
`,
			want:   "GW.Example.NET",
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.StubConfig(t, tt.config)

			got, ok := apiHost(tt.host)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}
