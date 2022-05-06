package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestGQLClientDo(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		httpMocks  func()
		wantErr    bool
		wantErrMsg string
		wantLogin  string
	}{
		{
			name: "success request",
			httpMocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					BodyString(`{"query":"QUERY","variables":{"var":"test"}}`).
					Reply(200).
					JSON(`{"data":{"viewer":{"login":"hubot"}}}`)
			},
			wantLogin: "hubot",
		},
		{
			name: "fail request",
			httpMocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					BodyString(`{"query":"QUERY","variables":{"var":"test"}}`).
					Reply(200).
					JSON(`{"errors":[{"message":"OH NO"},{"message":"this is fine"}]}`)
			},
			wantErr:    true,
			wantErrMsg: "GQL error: OH NO\nthis is fine",
		},
		{
			name: "http fail request empty response",
			httpMocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					BodyString(`{"query":"QUERY","variables":{"var":"test"}}`).
					Reply(404).
					JSON(`{}`)
			},
			wantErr:    true,
			wantErrMsg: "HTTP 404 (https://api.github.com/graphql)",
		},
		{
			name: "http fail request message response",
			httpMocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					BodyString(`{"query":"QUERY","variables":{"var":"test"}}`).
					Reply(422).
					JSON(`{"message": "OH NO"}`)
			},
			wantErr:    true,
			wantErrMsg: "HTTP 422: OH NO (https://api.github.com/graphql)",
		},
		{
			name: "http fail request errors response",
			httpMocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					BodyString(`{"query":"QUERY","variables":{"var":"test"}}`).
					Reply(502).
					JSON(`{"errors":[{"message":"Something went wrong"}]}`)
			},
			wantErr:    true,
			wantErrMsg: "HTTP 502: Something went wrong (https://api.github.com/graphql)",
		},
		{
			name: "support enterprise hosts",
			host: "enterprise.com",
			httpMocks: func() {
				gock.New("https://enterprise.com").
					Post("/api/graphql").
					BodyString(`{"query":"QUERY","variables":{"var":"test"}}`).
					Reply(200).
					JSON(`{"data":{"viewer":{"login":"hubot"}}}`)
			},
			wantLogin: "hubot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(gock.Off)
			if tt.host == "" {
				tt.host = "github.com"
			}
			if tt.httpMocks != nil {
				tt.httpMocks()
			}
			client := NewGQLClient(tt.host, nil)
			vars := map[string]interface{}{"var": "test"}
			res := struct{ Viewer struct{ Login string } }{}
			err := client.Do("QUERY", vars, &res)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.True(t, gock.IsDone())
			assert.Equal(t, tt.wantLogin, res.Viewer.Login)
		})
	}
}
