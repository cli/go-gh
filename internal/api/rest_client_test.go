package api

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRESTClientRaw(t *testing.T) {
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
					Reply(204).
					JSON(`{}`)
			},
			wantBody: `{}`,
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
			wantErrMsg: "HTTP 422 (https://api.github.com/some/test/path)",
			wantBody:   `{"message": "OH NO"}`,
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
			wantBody: `{}`,
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
			client := NewRESTClient(tt.host, nil)
			resp, err := client.Raw("GET", tt.path, nil)
			t.Cleanup(func() { resp.Body.Close() })
			body, readErr := io.ReadAll(resp.Body)
			assert.NoError(t, readErr)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
			assert.Equal(t, tt.wantBody, string(body))
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
			client := NewRESTClient(tt.host, nil)
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
	client := NewRESTClient("github.com", nil)
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
	client := NewRESTClient("github.com", nil)
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
	client := NewRESTClient("github.com", nil)
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
	client := NewRESTClient("github.com", nil)
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
	client := NewRESTClient("github.com", nil)
	r := bytes.NewReader([]byte(`{}`))
	err := client.Put("some/path/here", r, nil)
	assert.NoError(t, err)
	assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
}

func printPendingMocks(mocks []gock.Mock) string {
	paths := []string{}
	for _, mock := range mocks {
		paths = append(paths, mock.Request().URLStruct.String())
	}
	return fmt.Sprintf("%d unmatched mocks: %s", len(paths), strings.Join(paths, ", "))
}
