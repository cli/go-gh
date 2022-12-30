package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRESTClient(t *testing.T) {
	stubConfig(t, testConfig())
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Get("/some/test/path").
		MatchHeader("Authorization", "token abc123").
		Reply(200).
		JSON(`{"message": "success"}`)

	client, err := DefaultRESTClient()
	assert.NoError(t, err)

	res := struct{ Message string }{}
	err = client.Do("GET", "some/test/path", nil, &res)
	assert.NoError(t, err)
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
	assert.Equal(t, "success", res.Message)
}

func TestRESTClientRequest(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		path       string
		httpMocks  func()
		wantErr    bool
		wantErrMsg string
		wantBody   string
	}{
		{
			name: "success request empty response",
			path: "some/test/path",
			httpMocks: func() {
				gock.New("https://api.github.com").
					Get("/some/test/path").
					Reply(204)
			},
			wantBody: ``,
		},
		{
			name: "success request non-empty response",
			path: "some/test/path",
			httpMocks: func() {
				gock.New("https://api.github.com").
					Get("/some/test/path").
					Reply(200).
					JSON(`{"message": "success"}`)
			},
			wantBody: `{"message": "success"}`,
		},
		{
			name: "fail request empty response",
			path: "some/test/path",
			httpMocks: func() {
				gock.New("https://api.github.com").
					Get("/some/test/path").
					Reply(404).
					JSON(`{}`)
			},
			wantErr:    true,
			wantErrMsg: "HTTP 404 (https://api.github.com/some/test/path)",
			wantBody:   `{}`,
		},
		{
			name: "fail request non-empty response",
			path: "some/test/path",
			httpMocks: func() {
				gock.New("https://api.github.com").
					Get("/some/test/path").
					Reply(422).
					JSON(`{"message": "OH NO"}`)
			},
			wantErr:    true,
			wantErrMsg: "HTTP 422: OH NO (https://api.github.com/some/test/path)",
			wantBody:   `{"message": "OH NO"}`,
		},
		{
			name: "support full urls",
			path: "https://example.com/someother/test/path",
			httpMocks: func() {
				gock.New("https://example.com").
					Get("/someother/test/path").
					Reply(200).
					JSON(`{}`)
			},
			wantBody: `{}`,
		},
		{
			name: "support enterprise hosts",
			host: "enterprise.com",
			path: "some/test/path",
			httpMocks: func() {
				gock.New("https://enterprise.com").
					Get("/some/test/path").
					Reply(200).
					JSON(`{}`)
			},
			wantBody: `{}`,
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
			client, _ := NewRESTClient(ClientOptions{
				Host:      tt.host,
				AuthToken: "token",
				Transport: http.DefaultTransport,
			})

			resp, err := client.Request("GET", tt.path, nil)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}

			if err == nil {
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantBody, string(body))
			}

			assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
		})
	}
}

func TestRESTClientDo(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		path       string
		httpMocks  func()
		wantErr    bool
		wantErrMsg string
		wantMsg    string
	}{
		{
			name: "success request empty response",
			path: "some/test/path",
			httpMocks: func() {
				gock.New("https://api.github.com").
					Get("/some/test/path").
					Reply(204).
					JSON(`{}`)
			},
		},
		{
			name: "success request non-empty response",
			path: "some/test/path",
			httpMocks: func() {
				gock.New("https://api.github.com").
					Get("/some/test/path").
					Reply(200).
					JSON(`{"message": "success"}`)
			},
			wantMsg: "success",
		},
		{
			name: "fail request empty response",
			path: "some/test/path",
			httpMocks: func() {
				gock.New("https://api.github.com").
					Get("/some/test/path").
					Reply(404).
					JSON(`{}`)
			},
			wantErr:    true,
			wantErrMsg: "HTTP 404 (https://api.github.com/some/test/path)",
		},
		{
			name: "fail request non-empty response",
			path: "some/test/path",
			httpMocks: func() {
				gock.New("https://api.github.com").
					Get("/some/test/path").
					Reply(422).
					JSON(`{"message": "OH NO"}`)
			},
			wantErr:    true,
			wantErrMsg: "HTTP 422: OH NO (https://api.github.com/some/test/path)",
		},
		{
			name: "support full urls",
			path: "https://example.com/someother/test/path",
			httpMocks: func() {
				gock.New("https://example.com").
					Get("/someother/test/path").
					Reply(204).
					JSON(`{}`)
			},
		},
		{
			name: "support enterprise hosts",
			host: "enterprise.com",
			path: "some/test/path",
			httpMocks: func() {
				gock.New("https://enterprise.com").
					Get("/some/test/path").
					Reply(204).
					JSON(`{}`)
			},
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
			client, _ := NewRESTClient(ClientOptions{
				Host:      tt.host,
				AuthToken: "token",
				Transport: http.DefaultTransport,
			})
			res := struct{ Message string }{}
			err := client.Do("GET", tt.path, nil, &res)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
			assert.Equal(t, tt.wantMsg, res.Message)
		})
	}
}

func TestRESTClientDelete(t *testing.T) {
	t.Cleanup(gock.Off)
	gock.New("https://api.github.com").
		Delete("/some/path/here").
		Reply(204).
		JSON(`{}`)
	client, _ := NewRESTClient(ClientOptions{
		Host:      "github.com",
		AuthToken: "token",
		Transport: http.DefaultTransport,
	})
	err := client.Delete("some/path/here", nil)
	assert.NoError(t, err)
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
}

func TestRESTClientGet(t *testing.T) {
	t.Cleanup(gock.Off)
	gock.New("https://api.github.com").
		Get("/some/path/here").
		Reply(204).
		JSON(`{}`)
	client, _ := NewRESTClient(ClientOptions{
		Host:      "github.com",
		AuthToken: "token",
		Transport: http.DefaultTransport,
	})
	err := client.Get("some/path/here", nil)
	assert.NoError(t, err)
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
}

func TestRESTClientPatch(t *testing.T) {
	t.Cleanup(gock.Off)
	gock.New("https://api.github.com").
		Patch("/some/path/here").
		BodyString(`{}`).
		Reply(204).
		JSON(`{}`)
	client, _ := NewRESTClient(ClientOptions{
		Host:      "github.com",
		AuthToken: "token",
		Transport: http.DefaultTransport,
	})
	r := bytes.NewReader([]byte(`{}`))
	err := client.Patch("some/path/here", r, nil)
	assert.NoError(t, err)
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
}

func TestRESTClientPost(t *testing.T) {
	t.Cleanup(gock.Off)
	gock.New("https://api.github.com").
		Post("/some/path/here").
		BodyString(`{}`).
		Reply(204).
		JSON(`{}`)
	client, _ := NewRESTClient(ClientOptions{
		Host:      "github.com",
		AuthToken: "token",
		Transport: http.DefaultTransport,
	})
	r := bytes.NewReader([]byte(`{}`))
	err := client.Post("some/path/here", r, nil)
	assert.NoError(t, err)
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
}

func TestRESTClientPut(t *testing.T) {
	t.Cleanup(gock.Off)
	gock.New("https://api.github.com").
		Put("/some/path/here").
		BodyString(`{}`).
		Reply(204).
		JSON(`{}`)
	client, _ := NewRESTClient(ClientOptions{
		Host:      "github.com",
		AuthToken: "token",
		Transport: http.DefaultTransport,
	})
	r := bytes.NewReader([]byte(`{}`))
	err := client.Put("some/path/here", r, nil)
	assert.NoError(t, err)
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
}

func TestRESTClientDoWithContext(t *testing.T) {
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
			wantErrMsg: `Get "https://api.github.com/some/path": context canceled`,
		},
		{
			name: "http fail request timed out",
			getCtx: func() context.Context {
				// pass current time to ensure that deadline has already passed
				ctx, cancel := context.WithDeadline(context.Background(), time.Now())
				cancel()
				return ctx
			},
			wantErrMsg: `Get "https://api.github.com/some/path": context deadline exceeded`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			t.Cleanup(gock.Off)
			gock.New("https://api.github.com").
				Get("/some/path").
				Reply(204).
				JSON(`{}`)

			client, _ := NewRESTClient(ClientOptions{
				Host:      "github.com",
				AuthToken: "token",
				Transport: http.DefaultTransport,
			})
			res := struct{ Message string }{}

			// when
			ctx := tt.getCtx()
			gotErr := client.DoWithContext(ctx, http.MethodGet, "some/path", nil, &res)

			// then
			assert.EqualError(t, gotErr, tt.wantErrMsg)
			assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
		})
	}
}

func TestRESTClientRequestWithContext(t *testing.T) {
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
			wantErrMsg: `Get "https://api.github.com/some/path": context canceled`,
		},
		{
			name: "http fail request timed out",
			getCtx: func() context.Context {
				// pass current time to ensure that deadline has already passed
				ctx, cancel := context.WithDeadline(context.Background(), time.Now())
				cancel()
				return ctx
			},
			wantErrMsg: `Get "https://api.github.com/some/path": context deadline exceeded`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			t.Cleanup(gock.Off)
			gock.New("https://api.github.com").
				Get("/some/path").
				Reply(204).
				JSON(`{}`)

			client, _ := NewRESTClient(ClientOptions{
				Host:      "github.com",
				AuthToken: "token",
				Transport: http.DefaultTransport,
			})

			// when
			ctx := tt.getCtx()
			_, gotErr := client.RequestWithContext(ctx, http.MethodGet, "some/path", nil)

			// then
			assert.EqualError(t, gotErr, tt.wantErrMsg)
			assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
		})
	}
}

func TestRestPrefix(t *testing.T) {
	tests := []struct {
		name         string
		host         string
		wantEndpoint string
	}{
		{
			name:         "github",
			host:         "github.com",
			wantEndpoint: "https://api.github.com/",
		},
		{
			name:         "localhost",
			host:         "github.localhost",
			wantEndpoint: "http://api.github.localhost/",
		},
		{
			name:         "garage",
			host:         "garage.github.com",
			wantEndpoint: "https://garage.github.com/api/v3/",
		},
		{
			name:         "enterprise",
			host:         "enterprise.com",
			wantEndpoint: "https://enterprise.com/api/v3/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := restPrefix(tt.host)
			assert.Equal(t, tt.wantEndpoint, endpoint)
		})
	}
}
