package api

import (
	"bytes"
	"testing"

	"github.com/cli/go-gh/internal/transportmock"
	"github.com/cli/go-gh/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestRESTClientDo(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		path       string
		matcher    transportmock.Matcher
		responder  transportmock.Responder
		wantErr    bool
		wantErrMsg string
		wantHost   string
		wantMsg    string
	}{
		{
			name: "success request empty response",
			path: "some/test/path",
		},
		{
			name:      "success request non-empty response",
			path:      "some/test/path",
			responder: transportmock.RESTResponse(`{"message": "success"}`, nil),
			wantMsg:   "success",
		},
		{
			name:       "fail request empty response",
			path:       "some/test/path",
			responder:  transportmock.HTTPResponse(404, nil, `{}`, nil),
			wantErr:    true,
			wantErrMsg: "HTTP 404 (https://api.github.com/some/test/path)",
		},
		{
			name:       "fail request non-empty response",
			path:       "some/test/path",
			responder:  transportmock.HTTPResponse(422, nil, `{"message": "OH NO"}`, nil),
			wantErr:    true,
			wantErrMsg: "HTTP 422: OH NO (https://api.github.com/some/test/path)",
		},
		{
			name:     "support full urls",
			path:     "https://example.com/someother/test/path",
			matcher:  transportmock.REST("GET", "someother/test/path"),
			wantHost: "example.com",
		},
		{
			name:     "support enterprise hosts",
			host:     "enterprise.com",
			path:     "some/test/path",
			matcher:  transportmock.REST("GET", "api/v3/some/test/path"),
			wantHost: "enterprise.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.host == "" {
				tt.host = "github.com"
			}
			if tt.wantHost == "" {
				tt.wantHost = "api.github.com"
			}
			if tt.matcher == nil {
				tt.matcher = transportmock.REST("GET", "some/test/path")
			}
			if tt.responder == nil {
				tt.responder = transportmock.HTTPResponse(204, nil, "{}", nil)
			}
			http := transportmock.NewRegistry(t)
			client := NewRESTClient(tt.host, &api.ClientOptions{Transport: http})
			http.Register(tt.name, tt.matcher, tt.responder)
			res := struct{ Message string }{}
			err := client.Do("GET", tt.path, nil, &res)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantMsg, res.Message)
			assert.Equal(t, tt.wantHost, http.Requests()[0].URL.Hostname())
		})
	}
}

func TestRESTClientDelete(t *testing.T) {
	http := transportmock.NewRegistry(t)
	client := NewRESTClient("github.com", &api.ClientOptions{Transport: http})
	http.Register(
		"TestRESTClientDelete",
		transportmock.REST("DELETE", "some/path/here"),
		transportmock.HTTPResponse(204, nil, "{}", nil),
	)
	err := client.Delete("some/path/here", nil)
	assert.NoError(t, err)
}

func TestRESTClientGet(t *testing.T) {
	http := transportmock.NewRegistry(t)
	client := NewRESTClient("github.com", &api.ClientOptions{Transport: http})
	http.Register(
		"TestRESTClientGet",
		transportmock.REST("GET", "some/path/here"),
		transportmock.HTTPResponse(204, nil, "{}", nil),
	)
	err := client.Get("some/path/here", nil)
	assert.NoError(t, err)
}

func TestRESTClientPatch(t *testing.T) {
	http := transportmock.NewRegistry(t)
	client := NewRESTClient("github.com", &api.ClientOptions{Transport: http})
	http.Register(
		"TestRESTClientPatch",
		transportmock.REST("PATCH", "some/path/here"),
		transportmock.HTTPResponse(204, nil, "{}", nil),
	)
	r := bytes.NewReader([]byte(`{}`))
	err := client.Patch("some/path/here", r, nil)
	assert.NoError(t, err)
}

func TestRESTClientPost(t *testing.T) {
	http := transportmock.NewRegistry(t)
	client := NewRESTClient("github.com", &api.ClientOptions{Transport: http})
	http.Register(
		"TestRESTClientPost",
		transportmock.REST("POST", "some/path/here"),
		transportmock.HTTPResponse(204, nil, "{}", nil),
	)
	r := bytes.NewReader([]byte(`{}`))
	err := client.Post("some/path/here", r, nil)
	assert.NoError(t, err)
}

func TestRESTClientPut(t *testing.T) {
	http := transportmock.NewRegistry(t)
	client := NewRESTClient("github.com", &api.ClientOptions{Transport: http})
	http.Register(
		"TestRESTClientPut",
		transportmock.REST("PUT", "some/path/here"),
		transportmock.HTTPResponse(204, nil, "{}", nil),
	)
	r := bytes.NewReader([]byte(`{}`))
	err := client.Put("some/path/here", r, nil)
	assert.NoError(t, err)
}
