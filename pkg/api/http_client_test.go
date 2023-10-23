package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/config"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestHTTPClient(t *testing.T) {
	stubConfig(t, testConfig())
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Get("/some/test/path").
		MatchHeader("Authorization", "token abc123").
		Reply(200).
		JSON(`{"message": "success"}`)

	client, err := DefaultHTTPClient()
	assert.NoError(t, err)

	res, err := client.Get("https://api.github.com/some/test/path")
	assert.NoError(t, err)
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
	assert.Equal(t, 200, res.StatusCode)
}

func TestNewHTTPClient(t *testing.T) {
	reflectHTTP := tripper{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			header := req.Header.Clone()
			body := "{}"
			return &http.Response{
				StatusCode: 200,
				Header:     header,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
			}, nil
		},
	}

	tests := []struct {
		name        string
		enableLog   bool
		log         *bytes.Buffer
		host        string
		headers     map[string]string
		skipHeaders bool
		wantHeaders http.Header
	}{
		{
			name:        "sets default headers",
			wantHeaders: defaultHeaders(),
		},
		{
			name: "allows overriding default headers",
			headers: map[string]string{
				authorization: "token new_token",
				accept:        "application/vnd.github.test-preview",
			},
			wantHeaders: func() http.Header {
				h := defaultHeaders()
				h.Set(authorization, "token new_token")
				h.Set(accept, "application/vnd.github.test-preview")
				return h
			}(),
		},
		{
			name: "allows setting custom headers",
			headers: map[string]string{
				"custom": "testing",
			},
			wantHeaders: func() http.Header {
				h := defaultHeaders()
				h.Set("custom", "testing")
				return h
			}(),
		},
		{
			name:        "allows setting logger",
			enableLog:   true,
			log:         &bytes.Buffer{},
			wantHeaders: defaultHeaders(),
		},
		{
			name: "does not add an authorization header for non-matching host",
			host: "notauthorized.com",
			wantHeaders: func() http.Header {
				h := defaultHeaders()
				h.Del(authorization)
				return h
			}(),
		},
		{
			name: "does not add an authorization header for non-matching host subdomain",
			host: "test.company",
			wantHeaders: func() http.Header {
				h := defaultHeaders()
				h.Del(authorization)
				return h
			}(),
		},
		{
			name:        "adds an authorization header for a matching host",
			host:        "test.com",
			wantHeaders: defaultHeaders(),
		},
		{
			name:        "adds an authorization header if hosts match but differ in case",
			host:        "TeSt.CoM",
			wantHeaders: defaultHeaders(),
		},
		{
			name:        "skips default headers",
			skipHeaders: true,
			wantHeaders: func() http.Header {
				h := defaultHeaders()
				h.Del(accept)
				h.Del(contentType)
				h.Del(timeZone)
				h.Del(userAgent)
				return h
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.host == "" {
				tt.host = "test.com"
			}
			opts := ClientOptions{
				Host:               tt.host,
				AuthToken:          "oauth_token",
				Headers:            tt.headers,
				SkipDefaultHeaders: tt.skipHeaders,
				Transport:          reflectHTTP,
				LogIgnoreEnv:       true,
			}
			if tt.enableLog {
				opts.Log = tt.log
			}
			client, _ := NewHTTPClient(opts)
			res, err := client.Get("https://test.com")
			assert.NoError(t, err)
			assert.Equal(t, tt.wantHeaders, res.Header)
			if tt.enableLog {
				assert.NotEmpty(t, tt.log)
			}
		})
	}
}

func TestIsEnterprise(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantOut bool
	}{
		{
			name:    "github",
			host:    "github.com",
			wantOut: false,
		},
		{
			name:    "localhost",
			host:    "github.localhost",
			wantOut: false,
		},
		{
			name:    "enterprise",
			host:    "mygithub.com",
			wantOut: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := isEnterprise(tt.host)
			assert.Equal(t, tt.wantOut, out)
		})
	}
}

func TestNormalizeHostname(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		wantHost string
	}{
		{
			name:     "github domain",
			host:     "test.github.com",
			wantHost: "github.com",
		},
		{
			name:     "capitalized",
			host:     "GitHub.com",
			wantHost: "github.com",
		},
		{
			name:     "localhost domain",
			host:     "test.github.localhost",
			wantHost: "github.localhost",
		},
		{
			name:     "enterprise domain",
			host:     "mygithub.com",
			wantHost: "mygithub.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized := normalizeHostname(tt.host)
			assert.Equal(t, tt.wantHost, normalized)
		})
	}
}

type tripper struct {
	roundTrip func(*http.Request) (*http.Response, error)
}

func (tr tripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return tr.roundTrip(req)
}

func defaultHeaders() http.Header {
	h := http.Header{}
	a := "application/vnd.github.merge-info-preview+json"
	a += ", application/vnd.github.nebula-preview"
	h.Set(contentType, jsonContentType)
	h.Set(userAgent, "go-gh")
	h.Set(authorization, fmt.Sprintf("token %s", "oauth_token"))
	h.Set(timeZone, currentTimeZone())
	h.Set(accept, a)
	return h
}

func stubConfig(t *testing.T, cfgStr string) {
	t.Helper()
	old := config.Read
	config.Read = func(_ *config.Config) (*config.Config, error) {
		return config.ReadFromString(cfgStr), nil
	}
	t.Cleanup(func() {
		config.Read = old
	})
}

func printPendingMocks(mocks []gock.Mock) string {
	paths := []string{}
	for _, mock := range mocks {
		paths = append(paths, mock.Request().URLStruct.String())
	}
	return fmt.Sprintf("%d unmatched mocks: %s", len(paths), strings.Join(paths, ", "))
}
