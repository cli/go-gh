package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/cli/go-gh/pkg/api"
	"github.com/stretchr/testify/assert"
)

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
		headers     map[string]string
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := api.ClientOptions{
				AuthToken: "oauth_token",
				Headers:   tt.headers,
				Transport: reflectHTTP,
			}
			if tt.enableLog {
				opts.Log = tt.log
			}
			client := newHTTPClient(&opts)
			res, err := client.Get("test.com")
			assert.NoError(t, err)
			assert.Equal(t, tt.wantHeaders, res.Header)
			if tt.enableLog {
				assert.NotEmpty(t, tt.log)
			}
		})
	}
}

func TestNewHTTPClientWithDifferentHost(t *testing.T) {
	reflectHTTP := tripper{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			header := req.Header.Clone()
			body := "{}"
			return &http.Response{
				StatusCode: 200,
				Header:     header,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
				Request:    req,
			}, nil
		},
	}

	tests := []struct {
		name          string
		headers       map[string]string
		optsHost      string
		reqHost       string
		authToken     string
		wantAuthToken string
	}{
		{
			name:          "removes authorization header for a different host",
			optsHost:      "github.com",
			reqHost:       "https://nothub.com",
			authToken:     "oauth_token",
			wantAuthToken: "",
		},
		{
			name:          "does not remove authorization header for a matching host",
			optsHost:      "github.com",
			reqHost:       "https://api.github.com",
			authToken:     "oauth_token",
			wantAuthToken: "token oauth_token",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := api.ClientOptions{
				Host:      tt.optsHost,
				AuthToken: tt.authToken,
				Headers:   tt.headers,
				Transport: reflectHTTP,
			}

			req, err := http.NewRequest(http.MethodGet, tt.reqHost, nil)

			if err != nil {
				return
			}

			client := newHTTPClient(&opts)
			res, err := client.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantAuthToken, res.Request.Header.Get("authorization"))
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
	a += ", application/vnd.github.antiope-preview"
	a += ", application/vnd.github.shadow-cat-preview"
	h.Set(contentType, jsonContentType)
	h.Set(userAgent, "go-gh")
	h.Set(authorization, fmt.Sprintf("token %s", "oauth_token"))
	h.Set(timeZone, currentTimeZone())
	h.Set(accept, a)
	return h
}
