package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveOptions(t *testing.T) {
	stubConfig(t, testConfigWithSocket())

	tests := []struct {
		name          string
		opts          ClientOptions
		wantAuthToken string
		wantHost      string
		wantSocket    string
	}{
		{
			name: "honors consumer provided ClientOptions",
			opts: ClientOptions{
				Host:             "test.com",
				AuthToken:        "token_from_opts",
				UnixDomainSocket: "socket_from_opts",
			},
			wantAuthToken: "token_from_opts",
			wantHost:      "test.com",
			wantSocket:    "socket_from_opts",
		},
		{
			name:          "uses config values if there are no consumer provided ClientOptions",
			opts:          ClientOptions{},
			wantAuthToken: "token",
			wantHost:      "github.com",
			wantSocket:    "socket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := resolveOptions(tt.opts)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantHost, opts.Host)
			assert.Equal(t, tt.wantAuthToken, opts.AuthToken)
			assert.Equal(t, tt.wantSocket, opts.UnixDomainSocket)
		})
	}
}

func TestOptionsNeedResolution(t *testing.T) {
	tests := []struct {
		name string
		opts ClientOptions
		out  bool
	}{
		{
			name: "Host, AuthToken, and UnixDomainSocket specified",
			opts: ClientOptions{
				Host:             "test.com",
				AuthToken:        "token",
				UnixDomainSocket: "socket",
			},
			out: false,
		},
		{
			name: "Host, AuthToken, and Transport specified",
			opts: ClientOptions{
				Host:      "test.com",
				AuthToken: "token",
				Transport: http.DefaultTransport,
			},
			out: false,
		},
		{
			name: "Host, and AuthToken specified",
			opts: ClientOptions{
				Host:      "test.com",
				AuthToken: "token",
			},
			out: true,
		},
		{
			name: "Host, and UnixDomainSocket specified",
			opts: ClientOptions{
				Host:             "test.com",
				UnixDomainSocket: "socket",
			},
			out: true,
		},
		{
			name: "Host, and Transport specified",
			opts: ClientOptions{
				Host:      "test.com",
				Transport: http.DefaultTransport,
			},
			out: true,
		},
		{
			name: "AuthToken, and UnixDomainSocket specified",
			opts: ClientOptions{
				AuthToken:        "token",
				UnixDomainSocket: "socket",
			},
			out: true,
		},
		{
			name: "AuthToken, and Transport specified",
			opts: ClientOptions{
				AuthToken: "token",
				Transport: http.DefaultTransport,
			},
			out: true,
		},
		{
			name: "Host specified",
			opts: ClientOptions{
				Host: "test.com",
			},
			out: true,
		},
		{
			name: "AuthToken specified",
			opts: ClientOptions{
				AuthToken: "token",
			},
			out: true,
		},
		{
			name: "UnixDomainSocket specified",
			opts: ClientOptions{
				UnixDomainSocket: "socket",
			},
			out: true,
		},
		{
			name: "Transport specified",
			opts: ClientOptions{
				Transport: http.DefaultTransport,
			},
			out: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.out, optionsNeedResolution(tt.opts))
		})
	}
}

func testConfig() string {
	return `
hosts:
  github.com:
    user: user1
    oauth_token: abc123
    git_protocol: ssh
`
}

func testConfigWithSocket() string {
	return `
http_unix_socket: socket
hosts:
  github.com:
    user: user1
    oauth_token: token
    git_protocol: ssh
`
}
