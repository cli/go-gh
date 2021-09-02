package api

import (
	"testing"

	"github.com/cli/go-gh/internal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestGQLClientDo(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		matcher    httpmock.Matcher
		responder  httpmock.Responder
		wantErr    bool
		wantErrMsg string
		wantHost   string
		wantLogin  string
	}{
		{
			name:      "success request",
			responder: httpmock.StringResponse(`{"data":{"viewer":{"login":"hubot"}}}`),
			wantLogin: "hubot",
		},
		{
			name:       "fail request",
			responder:  httpmock.StringResponse(`{"errors":[{"message":"OH NO"},{"message":"this is fine"}]}`),
			wantErr:    true,
			wantErrMsg: "GQL error: OH NO\nthis is fine",
		},
		{
			name:       "http fail request empty response",
			responder:  httpmock.StatusStringResponse(404, `{}`),
			wantErr:    true,
			wantErrMsg: "HTTP 404 (https://api.github.com/graphql)",
		},
		{
			name:       "http fail request message response",
			responder:  httpmock.StatusStringResponse(422, `{"message": "OH NO"}`),
			wantErr:    true,
			wantErrMsg: "HTTP 422: OH NO (https://api.github.com/graphql)",
		},
		{
			name:       "http fail request errors response",
			responder:  httpmock.StatusStringResponse(502, `{"errors":[{"message":"Something went wrong"}]}`),
			wantErr:    true,
			wantErrMsg: "HTTP 502: Something went wrong (https://api.github.com/graphql)",
		},
		{
			name:      "support enterprise hosts",
			responder: httpmock.StatusStringResponse(204, "{}"),
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
			http := httpmock.NewRegistry(t)
			client := NewGQLClient(tt.host, ClientOptions{Transport: http})
			http.Register(
				httpmock.GQL("QUERY"),
				tt.responder,
			)
			vars := map[string]interface{}{"var": "test"}
			res := struct{ Viewer struct{ Login string } }{}
			err := client.Do("QUERY", vars, &res)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantLogin, res.Viewer.Login)
			assert.Equal(t, tt.wantHost, http.Requests[0].URL.Hostname())
		})
	}
}
