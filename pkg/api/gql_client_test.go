package api

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestGQLClient(t *testing.T) {
	stubConfig(t, testConfig())
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Post("/graphql").
		MatchHeader("Authorization", "token abc123").
		BodyString(`{"query":"QUERY","variables":{"var":"test"}}`).
		Reply(200).
		JSON(`{"data":{"viewer":{"login":"hubot"}}}`)

	client, err := DefaultGQLClient()
	assert.NoError(t, err)

	vars := map[string]interface{}{"var": "test"}
	res := struct{ Viewer struct{ Login string } }{}
	err = client.Do("QUERY", vars, &res)
	assert.NoError(t, err)
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
	assert.Equal(t, "hubot", res.Viewer.Login)
}

func TestGQLClientDoError(t *testing.T) {
	stubConfig(t, testConfig())
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Post("/graphql").
		MatchHeader("Authorization", "token abc123").
		BodyString(`{"query":"QUERY","variables":null}`).
		Reply(200).
		JSON(`{"errors":[{"type":"NOT_FOUND","path":["organization"],"message":"Could not resolve to an Organization with the login of 'cli'."}]}`)

	client, err := DefaultGQLClient()
	assert.NoError(t, err)

	res := struct{ Organization struct{ Name string } }{}
	err = client.Do("QUERY", nil, &res)
	var gqlErr *GQLError
	assert.True(t, errors.As(err, &gqlErr))
	assert.EqualError(t, gqlErr, "GraphQL: Could not resolve to an Organization with the login of 'cli'. (organization)")
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
}

func TestGQLClientQueryError(t *testing.T) {
	stubConfig(t, testConfig())
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Post("/graphql").
		MatchHeader("Authorization", "token abc123").
		BodyString(`{"query":"query QUERY{organization{name}}"}`).
		Reply(200).
		JSON(`{"errors":[{"type":"NOT_FOUND","path":["organization"],"message":"Could not resolve to an Organization with the login of 'cli'."}]}`)

	client, err := DefaultGQLClient()
	assert.NoError(t, err)

	var res struct{ Organization struct{ Name string } }
	err = client.Query("QUERY", &res, nil)
	var gqlErr *GQLError
	assert.True(t, errors.As(err, &gqlErr))
	assert.EqualError(t, gqlErr, "GraphQL: Could not resolve to an Organization with the login of 'cli'. (organization)")
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
}

func TestGQLClientMutateError(t *testing.T) {
	stubConfig(t, testConfig())
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Post("/graphql").
		MatchHeader("Authorization", "token abc123").
		BodyString(`{"query":"mutation MUTATE($input:ID!){updateRepository{repository{name}}}","variables":{"input":"variables"}}`).
		Reply(200).
		JSON(`{"errors":[{"type":"NOT_FOUND","path":["organization"],"message":"Could not resolve to an Organization with the login of 'cli'."}]}`)

	client, err := DefaultGQLClient()
	assert.NoError(t, err)

	var mutation struct {
		UpdateRepository struct{ Repository struct{ Name string } }
	}
	variables := map[string]interface{}{"input": "variables"}
	err = client.Mutate("MUTATE", &mutation, variables)
	var gqlErr *GQLError
	assert.True(t, errors.As(err, &gqlErr))
	assert.EqualError(t, gqlErr, "GraphQL: Could not resolve to an Organization with the login of 'cli'. (organization)")
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
}

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
			wantErrMsg: "GraphQL: OH NO, this is fine",
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
			client, _ := NewGQLClient(ClientOptions{
				Host:      tt.host,
				AuthToken: "token",
				Transport: http.DefaultTransport,
			})
			vars := map[string]interface{}{"var": "test"}
			res := struct{ Viewer struct{ Login string } }{}
			err := client.Do("QUERY", vars, &res)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
			assert.Equal(t, tt.wantLogin, res.Viewer.Login)
		})
	}
}

func TestGQLClientDoWithContext(t *testing.T) {
	tests := []struct {
		name       string
		wantErrMsg string
		getCtx     func() context.Context
	}{
		{
			name: "http fail request canceled",
			getCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				// call 'cancel' to ensure that context is already canceled
				cancel()
				return ctx
			},
			wantErrMsg: `Post "https://api.github.com/graphql": context canceled`,
		},
		{
			name: "http fail request timed out",
			getCtx: func() context.Context {
				// pass current time to ensure that deadline has already passed
				ctx, cancel := context.WithDeadline(context.Background(), time.Now())
				cancel()
				return ctx
			},
			wantErrMsg: `Post "https://api.github.com/graphql": context deadline exceeded`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(gock.Off)
			gock.New("https://api.github.com").
				Post("/graphql").
				BodyString(`{"query":"QUERY","variables":{"var":"test"}}`).
				Reply(200).
				JSON(`{}`)

			client, _ := NewGQLClient(ClientOptions{
				Host:      "github.com",
				AuthToken: "token",
				Transport: http.DefaultTransport,
			})

			vars := map[string]interface{}{"var": "test"}
			res := struct{ Viewer struct{ Login string } }{}

			ctx := tt.getCtx()
			gotErr := client.DoWithContext(ctx, "QUERY", vars, &res)

			assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
			assert.EqualError(t, gotErr, tt.wantErrMsg)
		})
	}
}

func TestGQLEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		host         string
		wantEndpoint string
	}{
		{
			name:         "github",
			host:         "github.com",
			wantEndpoint: "https://api.github.com/graphql",
		},
		{
			name:         "localhost",
			host:         "github.localhost",
			wantEndpoint: "http://api.github.localhost/graphql",
		},
		{
			name:         "garage",
			host:         "garage.github.com",
			wantEndpoint: "https://garage.github.com/api/graphql",
		},
		{
			name:         "enterprise",
			host:         "enterprise.com",
			wantEndpoint: "https://enterprise.com/api/graphql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := gqlEndpoint(tt.host)
			assert.Equal(t, tt.wantEndpoint, endpoint)
		})
	}
}
