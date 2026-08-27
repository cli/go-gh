package api

import (
	"errors"
	"net/http"
	"testing"

	"github.com/cli/go-gh/v2/internal/testutils"
	"github.com/cli/go-gh/v2/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveOptions(t *testing.T) {
	testutils.StubConfig(t, testConfigWithSocket())

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

func TestValidAPIHost(t *testing.T) {
	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{name: "bare hostname", value: "gw.example.net", valid: true},
		{name: "mixed-case hostname", value: "GW.Example.NET", valid: true},
		{name: "empty"},
		{name: "scheme", value: "https://gw.example.net"},
		{name: "path", value: "gw.example.net/api"},
		{name: "query", value: "gw.example.net?trace=1"},
		{name: "fragment", value: "gw.example.net#fragment"},
		{name: "userinfo", value: "user@gw.example.net"},
		{name: "leading whitespace", value: " gw.example.net"},
		{name: "trailing whitespace", value: "gw.example.net "},
		{name: "empty host", value: ":8443"},
		{name: "missing port", value: "gw.example.net:"},
		{name: "non-numeric port", value: "gw.example.net:http"},
		{name: "minimum port", value: "gw.example.net:1"},
		{name: "typical port", value: "gw.example.net:8443"},
		{name: "maximum port", value: "gw.example.net:65535"},
		{name: "zero port", value: "gw.example.net:0"},
		{name: "out-of-range port", value: "gw.example.net:65536"},
		{name: "IPv6 literal", value: "[::1]"},
		{name: "bare IPv6 address", value: "::1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, validAPIHost(tt.value))
		})
	}
}

func TestResolveAPIHost(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		apiHost string
		config  string
		want    string
		wantErr string
	}{
		{
			name:    "explicit option takes precedence",
			host:    "example.ghe.com",
			apiHost: "explicit.example.net",
			config:  "hosts:\n  example.ghe.com:\n    api_host: configured.example.net\n",
			want:    "explicit.example.net",
		},
		{
			name:   "config fills an empty option",
			host:   "example.ghe.com",
			config: "hosts:\n  example.ghe.com:\n    api_host: configured.example.net\n",
			want:   "configured.example.net",
		},
		{
			name: "no option or config stays empty",
			host: "example.ghe.com",
		},
		{
			name:    "invalid explicit option fails",
			host:    "example.ghe.com",
			apiHost: "https://explicit.example.net",
			wantErr: `invalid api_host for example.ghe.com: "https://explicit.example.net" must be a hostname without a scheme or port, for example "api.example.com"`,
		},
		{
			name:    "explicit port fails",
			host:    "example.ghe.com",
			apiHost: "explicit.example.net:8443",
			wantErr: `invalid api_host for example.ghe.com: "explicit.example.net:8443" must be a hostname without a scheme or port, for example "api.example.com"`,
		},
		{
			name:    "invalid configured option fails",
			host:    "example.ghe.com",
			config:  "hosts:\n  example.ghe.com:\n    api_host: https://configured.example.net\n",
			wantErr: `invalid api_host for example.ghe.com: "https://configured.example.net" must be a hostname without a scheme or port, for example "api.example.com"`,
		},
		{
			name:    "configured port fails",
			host:    "example.ghe.com",
			config:  "hosts:\n  example.ghe.com:\n    api_host: configured.example.net:8443\n",
			wantErr: `invalid api_host for example.ghe.com: "configured.example.net:8443" must be a hostname without a scheme or port, for example "api.example.com"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.StubConfig(t, tt.config)
			opts := ClientOptions{Host: tt.host, APIHost: tt.apiHost}

			got, err := resolveAPIHost(opts)

			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got.APIHost)
		})
	}
}

func TestAPIClientConstructorsRejectInvalidAPIHost(t *testing.T) {
	constructors := []struct {
		name      string
		construct func(ClientOptions) error
	}{
		{
			name: "HTTP",
			construct: func(opts ClientOptions) error {
				_, err := NewHTTPClient(opts)
				return err
			},
		},
		{
			name: "REST",
			construct: func(opts ClientOptions) error {
				_, err := NewRESTClient(opts)
				return err
			},
		},
		{
			name: "GraphQL",
			construct: func(opts ClientOptions) error {
				_, err := NewGraphQLClient(opts)
				return err
			},
		},
	}

	for _, constructor := range constructors {
		t.Run(constructor.name+" configured value", func(t *testing.T) {
			testutils.StubConfig(t, "hosts:\n  example.ghe.com:\n    api_host: gw.example.net:8443\n")
			opts := ClientOptions{
				Host:      "example.ghe.com",
				AuthToken: "token",
				Transport: http.DefaultTransport,
			}

			err := constructor.construct(opts)

			require.EqualError(
				t,
				err,
				`invalid api_host for example.ghe.com: "gw.example.net:8443" must be a hostname without a scheme or port, for example "api.example.com"`,
			)
		})

		t.Run(constructor.name+" explicit value", func(t *testing.T) {
			testutils.StubConfig(t, "")
			opts := ClientOptions{
				Host:      "example.ghe.com",
				APIHost:   "gw.example.net:8443",
				AuthToken: "token",
				Transport: http.DefaultTransport,
			}

			err := constructor.construct(opts)

			require.EqualError(
				t,
				err,
				`invalid api_host for example.ghe.com: "gw.example.net:8443" must be a hostname without a scheme or port, for example "api.example.com"`,
			)
		})
	}
}

func TestAPIClientConstructorsIgnoreConfigReadErrors(t *testing.T) {
	oldRead := config.Read
	config.Read = func(*config.Config) (*config.Config, error) {
		return nil, &config.InvalidConfigFileError{
			Path: "hosts.yml",
			Err:  errors.New("invalid YAML"),
		}
	}
	t.Cleanup(func() {
		config.Read = oldRead
	})

	opts := ClientOptions{
		Host:      "example.ghe.com",
		AuthToken: "token",
		Transport: http.DefaultTransport,
	}

	resolved, err := resolveAPIHost(opts)
	require.NoError(t, err)
	assert.Equal(t, opts, resolved)

	constructors := []struct {
		name      string
		construct func(ClientOptions) error
	}{
		{
			name: "HTTP",
			construct: func(opts ClientOptions) error {
				_, err := NewHTTPClient(opts)
				return err
			},
		},
		{
			name: "REST",
			construct: func(opts ClientOptions) error {
				_, err := NewRESTClient(opts)
				return err
			},
		},
		{
			name: "GraphQL",
			construct: func(opts ClientOptions) error {
				_, err := NewGraphQLClient(opts)
				return err
			},
		},
	}

	for _, constructor := range constructors {
		t.Run(constructor.name, func(t *testing.T) {
			require.NoError(t, constructor.construct(opts))
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
