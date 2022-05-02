package api

import (
	"testing"

	"github.com/cli/go-gh/internal/transportmock"
	"github.com/cli/go-gh/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestGQLClientDo(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		matcher    transportmock.Matcher
		responder  transportmock.Responder
		wantErr    bool
		wantErrMsg string
		wantHost   string
		wantLogin  string
	}{
		{
			name:      "success request",
			responder: transportmock.RESTResponse(`{"data":{"viewer":{"login":"hubot"}}}`, nil),
			wantLogin: "hubot",
		},
		{
			name:       "fail request",
			responder:  transportmock.RESTResponse(`{"errors":[{"message":"OH NO"},{"message":"this is fine"}]}`, nil),
			wantErr:    true,
			wantErrMsg: "GQL error: OH NO\nthis is fine",
		},
		{
			name:       "http fail request empty response",
			responder:  transportmock.HTTPResponse(404, nil, `{}`, nil),
			wantErr:    true,
			wantErrMsg: "HTTP 404 (https://api.github.com/graphql)",
		},
		{
			name:       "http fail request message response",
			responder:  transportmock.HTTPResponse(422, nil, `{"message": "OH NO"}`, nil),
			wantErr:    true,
			wantErrMsg: "HTTP 422: OH NO (https://api.github.com/graphql)",
		},
		{
			name:       "http fail request errors response",
			responder:  transportmock.HTTPResponse(502, nil, `{"errors":[{"message":"Something went wrong"}]}`, nil),
			wantErr:    true,
			wantErrMsg: "HTTP 502: Something went wrong (https://api.github.com/graphql)",
		},
		{
			name:      "support enterprise hosts",
			responder: transportmock.HTTPResponse(204, nil, "{}", nil),
			host:      "enterprise.com",
			wantHost:  "enterprise.com",
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
			http := transportmock.NewRegistry(t)
			client := NewGQLClient(tt.host, &api.ClientOptions{Transport: http})
			matcher := transportmock.GQL("QUERY")
			http.Register(tt.name, matcher, tt.responder)
			vars := map[string]interface{}{"var": "test"}
			res := struct{ Viewer struct{ Login string } }{}
			err := client.Do("QUERY", vars, &res)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantLogin, res.Viewer.Login)
			assert.Equal(t, tt.wantHost, http.Requests()[0].URL.Hostname())
		})
	}
}
