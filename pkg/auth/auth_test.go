package auth

import (
	"os"
	"testing"

	"github.com/cli/go-gh/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	orig_GITHUB_TOKEN := os.Getenv("GITHUB_TOKEN")
	orig_GITHUB_ENTERPRISE_TOKEN := os.Getenv("GITHUB_ENTERPRISE_TOKEN")
	orig_GH_TOKEN := os.Getenv("GH_TOKEN")
	orig_GH_ENTERPRISE_TOKEN := os.Getenv("GH_ENTERPRISE_TOKEN")
	t.Cleanup(func() {
		os.Setenv("GITHUB_TOKEN", orig_GITHUB_TOKEN)
		os.Setenv("GITHUB_ENTERPRISE_TOKEN", orig_GITHUB_ENTERPRISE_TOKEN)
		os.Setenv("GH_TOKEN", orig_GH_TOKEN)
		os.Setenv("GH_ENTERPRISE_TOKEN", orig_GH_ENTERPRISE_TOKEN)
	})

	tests := []struct {
		name                    string
		host                    string
		GITHUB_TOKEN            string
		GITHUB_ENTERPRISE_TOKEN string
		GH_TOKEN                string
		GH_ENTERPRISE_TOKEN     string
		configFunc              func() (config.Config, error)
		wantToken               string
		wantErr                 bool
		wantErrMsg              string
	}{
		{
			name:       "token for github.com with no env tokens and no config token",
			host:       "github.com",
			wantErr:    true,
			wantErrMsg: "not found",
		},
		{
			name:       "token for enterprise.com with no env tokens and no config token",
			host:       "enterprise.com",
			wantErr:    true,
			wantErrMsg: "not found",
		},
		{
			name:         "token for github.com with GH_TOKEN, GITHUB_TOKEN, and config token",
			host:         "github.com",
			GH_TOKEN:     "GH_TOKEN",
			GITHUB_TOKEN: "GITHUB_TOKEN",
			configFunc:   testConfigFunc(),
			wantToken:    "GH_TOKEN",
		},
		{
			name:         "token for github.com with GITHUB_TOKEN, and config token",
			host:         "github.com",
			GITHUB_TOKEN: "GITHUB_TOKEN",
			configFunc:   testConfigFunc(),
			wantToken:    "GITHUB_TOKEN",
		},
		{
			name:       "token for github.com with config token",
			host:       "github.com",
			configFunc: testConfigFunc(),
			wantToken:  "xxxxxxxxxxxxxxxxxxxx",
		},
		{
			name:                    "token for enterprise.com with GH_ENTERPRISE_TOKEN, GITHUB_ENTERPRISE_TOKEN, and config token",
			host:                    "enterprise.com",
			GH_ENTERPRISE_TOKEN:     "GH_ENTERPRISE_TOKEN",
			GITHUB_ENTERPRISE_TOKEN: "GITHUB_ENTERPRISE_TOKEN",
			configFunc:              testConfigFunc(),
			wantToken:               "GH_ENTERPRISE_TOKEN",
		},
		{
			name:                    "token for enterprise.com with GITHUB_ENTERPRISE_TOKEN, and config token",
			host:                    "enterprise.com",
			GITHUB_ENTERPRISE_TOKEN: "GITHUB_ENTERPRISE_TOKEN",
			configFunc:              testConfigFunc(),
			wantToken:               "GITHUB_ENTERPRISE_TOKEN",
		},
		{
			name:       "token for enterprise.com with config token",
			host:       "enterprise.com",
			configFunc: testConfigFunc(),
			wantToken:  "yyyyyyyyyyyyyyyyyyyy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GITHUB_TOKEN", tt.GITHUB_TOKEN)
			os.Setenv("GITHUB_ENTERPRISE_TOKEN", tt.GITHUB_ENTERPRISE_TOKEN)
			os.Setenv("GH_TOKEN", tt.GH_TOKEN)
			os.Setenv("GH_ENTERPRISE_TOKEN", tt.GH_ENTERPRISE_TOKEN)
			token, err := token(tt.host, tt.configFunc)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantToken, token)
		})
	}
}

func testConfigFunc() func() (config.Config, error) {
	var data = `
hosts:
  github.com:
    oauth_token: xxxxxxxxxxxxxxxxxxxx
  enterprise.com:
    oauth_token: yyyyyyyyyyyyyyyyyyyy
`
	return func() (config.Config, error) {
		return config.FromString(data)
	}
}
