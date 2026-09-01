package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/internal/testutils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestHTTPClient(t *testing.T) {
	testutils.StubConfig(t, testConfig())
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
	testutils.StubConfig(t, "")

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
		apiHost     string
		reqURL      string
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
			name:        "adds authorization for a canonical subdomain",
			host:        "test.com",
			reqURL:      "https://api.test.com",
			wantHeaders: defaultHeaders(),
		},
		{
			name:        "adds authorization for exact API host",
			host:        "test.com",
			apiHost:     "gateway.example",
			reqURL:      "https://gateway.example",
			wantHeaders: defaultHeaders(),
		},
		{
			name:        "adds authorization for case-differing API host",
			host:        "test.com",
			apiHost:     "gateway.example",
			reqURL:      "https://GATEWAY.example",
			wantHeaders: defaultHeaders(),
		},
		{
			name:    "withholds authorization from canonical host when API host is configured",
			host:    "test.com",
			apiHost: "gateway.example",
			reqURL:  "https://test.com",
			wantHeaders: func() http.Header {
				h := defaultHeaders()
				h.Del(authorization)
				return h
			}(),
		},
		{
			name:    "withholds authorization from an API host subdomain",
			host:    "test.com",
			apiHost: "gateway.example",
			reqURL:  "https://sub.gateway.example",
			wantHeaders: func() http.Header {
				h := defaultHeaders()
				h.Del(authorization)
				return h
			}(),
		},
		{
			name:    "withholds authorization from unrelated host",
			host:    "test.com",
			apiHost: "gateway.example",
			reqURL:  "https://unrelated.example",
			wantHeaders: func() http.Header {
				h := defaultHeaders()
				h.Del(authorization)
				return h
			}(),
		},
		{
			name:   "withholds authorization from a port-only empty hostname without override",
			host:   "test.com",
			reqURL: "http://:1234/x",
			wantHeaders: func() http.Header {
				h := defaultHeaders()
				h.Del(authorization)
				return h
			}(),
		},
		{
			name:   "withholds authorization from an empty hostname without override",
			host:   "test.com",
			reqURL: "http:///x",
			wantHeaders: func() http.Header {
				h := defaultHeaders()
				h.Del(authorization)
				return h
			}(),
		},
		{
			name:    "withholds authorization from an empty hostname with override configured",
			host:    "test.com",
			apiHost: "gateway.example",
			reqURL:  "http://:1234/x",
			wantHeaders: func() http.Header {
				h := defaultHeaders()
				h.Del(authorization)
				return h
			}(),
		},
		{
			name:        "skips default headers",
			skipHeaders: true,
			wantHeaders: func() http.Header {
				h := defaultHeaders()
				h.Del(accept)
				h.Del(apiVersion)
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
			if tt.reqURL == "" {
				tt.reqURL = "https://test.com"
			}
			opts := ClientOptions{
				Host:               tt.host,
				APIHost:            tt.apiHost,
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
			res, err := client.Get(tt.reqURL)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantHeaders, res.Header)
			if tt.enableLog {
				assert.NotEmpty(t, tt.log)
			}
		})
	}
}

func TestNewHTTPClientCheckRedirect(t *testing.T) {
	// Redirect handling belongs to http.Client rather than the transport, so a
	// stub transport still exercises the real policy: the client asks it for the
	// redirected request only if the policy allows the redirect.
	newRecordingTransport := func(methods *[]string) tripper {
		return tripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				*methods = append(*methods, req.Method)
				if len(*methods) == 1 {
					return &http.Response{
						StatusCode: http.StatusMovedPermanently,
						Header:     http.Header{"Location": []string{"https://api.github.com/repos/OWNER/NEW"}},
						Body:       io.NopCloser(bytes.NewBufferString("")),
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}, nil
			},
		}
	}

	t.Run("follows redirects by default, downgrading DELETE to GET", func(t *testing.T) {
		var methods []string
		client, err := NewHTTPClient(ClientOptions{
			Host:      "github.com",
			AuthToken: "oauth_token",
			Transport: newRecordingTransport(&methods),
		})
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodDelete, "https://api.github.com/repos/OWNER/OLD", nil)
		assert.NoError(t, err)
		res, err := client.Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()

		// This is the behaviour that makes the option necessary. Go turns the
		// DELETE into a GET when it follows the redirect, so the caller is told
		// the request succeeded while nothing was deleted.
		assert.Equal(t, []string{http.MethodDelete, http.MethodGet}, methods)
		assert.Equal(t, http.StatusNoContent, res.StatusCode)
	})

	t.Run("honours a CheckRedirect that stops at the redirect", func(t *testing.T) {
		var methods []string
		client, err := NewHTTPClient(ClientOptions{
			Host:      "github.com",
			AuthToken: "oauth_token",
			Transport: newRecordingTransport(&methods),
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		})
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodDelete, "https://api.github.com/repos/OWNER/OLD", nil)
		assert.NoError(t, err)
		res, err := client.Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, []string{http.MethodDelete}, methods)
		assert.Equal(t, http.StatusMovedPermanently, res.StatusCode)
	})
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
	h.Set(apiVersion, apiVersionValue)
	h.Set(contentType, jsonContentType)
	h.Set(userAgent, "go-gh")
	h.Set(authorization, fmt.Sprintf("token %s", "oauth_token"))
	h.Set(timeZone, currentTimeZone())
	h.Set(accept, a)
	return h
}

func printPendingMocks(mocks []gock.Mock) string {
	paths := []string{}
	for _, mock := range mocks {
		paths = append(paths, mock.Request().URLStruct.String())
	}
	return fmt.Sprintf("%d unmatched mocks: %s", len(paths), strings.Join(paths, ", "))
}
