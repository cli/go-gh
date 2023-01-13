package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheResponse(t *testing.T) {
	counter := 0
	fakeHTTP := tripper{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			counter += 1
			body := fmt.Sprintf("%d: %s %s", counter, req.Method, req.URL.String())
			status := 200
			if req.URL.Path == "/error" {
				status = 500
			}
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
			}, nil
		},
	}

	cacheDir := filepath.Join(t.TempDir(), "gh-cli-cache")

	httpClient, err := NewHTTPClient(
		ClientOptions{
			Host:         "github.com",
			AuthToken:    "token",
			Transport:    fakeHTTP,
			EnableCache:  true,
			CacheDir:     cacheDir,
			LogIgnoreEnv: true,
		},
	)
	assert.NoError(t, err)

	do := func(method, url string, body io.Reader) (string, error) {
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return "", err
		}
		res, err := httpClient.Do(req)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			err = fmt.Errorf("ReadAll: %w", err)
		}
		return string(resBody), err
	}

	var res string

	res, err = do("GET", "http://example.com/path", nil)
	assert.NoError(t, err)
	assert.Equal(t, "1: GET http://example.com/path", res)
	res, err = do("GET", "http://example.com/path", nil)
	assert.NoError(t, err)
	assert.Equal(t, "1: GET http://example.com/path", res)

	res, err = do("GET", "http://example.com/path2", nil)
	assert.NoError(t, err)
	assert.Equal(t, "2: GET http://example.com/path2", res)

	res, err = do("POST", "http://example.com/path2", nil)
	assert.NoError(t, err)
	assert.Equal(t, "3: POST http://example.com/path2", res)

	res, err = do("POST", "http://example.com/graphql", bytes.NewBufferString(`hello`))
	assert.NoError(t, err)
	assert.Equal(t, "4: POST http://example.com/graphql", res)
	res, err = do("POST", "http://example.com/graphql", bytes.NewBufferString(`hello`))
	assert.NoError(t, err)
	assert.Equal(t, "4: POST http://example.com/graphql", res)

	res, err = do("POST", "http://example.com/graphql", bytes.NewBufferString(`hello2`))
	assert.NoError(t, err)
	assert.Equal(t, "5: POST http://example.com/graphql", res)

	res, err = do("GET", "http://example.com/error", nil)
	assert.NoError(t, err)
	assert.Equal(t, "6: GET http://example.com/error", res)
	res, err = do("GET", "http://example.com/error", nil)
	assert.NoError(t, err)
	assert.Equal(t, "7: GET http://example.com/error", res)
}

func TestCacheResponseRequestCacheOptions(t *testing.T) {
	counter := 0
	fakeHTTP := tripper{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			counter += 1
			body := fmt.Sprintf("%d: %s %s", counter, req.Method, req.URL.String())
			status := 200
			if req.URL.Path == "/error" {
				status = 500
			}
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
			}, nil
		},
	}

	cacheDir := filepath.Join(t.TempDir(), "gh-cli-cache")

	httpClient, err := NewHTTPClient(
		ClientOptions{
			Host:         "github.com",
			AuthToken:    "token",
			Transport:    fakeHTTP,
			EnableCache:  false,
			CacheDir:     cacheDir,
			LogIgnoreEnv: true,
		},
	)
	assert.NoError(t, err)

	do := func(method, url string, body io.Reader) (string, error) {
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return "", err
		}
		req.Header.Set("X-GH-CACHE-DIR", cacheDir)
		req.Header.Set("X-GH-CACHE-TTL", "1h")
		res, err := httpClient.Do(req)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			err = fmt.Errorf("ReadAll: %w", err)
		}
		return string(resBody), err
	}

	var res string

	res, err = do("GET", "http://example.com/path", nil)
	assert.NoError(t, err)
	assert.Equal(t, "1: GET http://example.com/path", res)
	res, err = do("GET", "http://example.com/path", nil)
	assert.NoError(t, err)
	assert.Equal(t, "1: GET http://example.com/path", res)

	res, err = do("GET", "http://example.com/path2", nil)
	assert.NoError(t, err)
	assert.Equal(t, "2: GET http://example.com/path2", res)

	res, err = do("POST", "http://example.com/path2", nil)
	assert.NoError(t, err)
	assert.Equal(t, "3: POST http://example.com/path2", res)

	res, err = do("POST", "http://example.com/graphql", bytes.NewBufferString(`hello`))
	assert.NoError(t, err)
	assert.Equal(t, "4: POST http://example.com/graphql", res)
	res, err = do("POST", "http://example.com/graphql", bytes.NewBufferString(`hello`))
	assert.NoError(t, err)
	assert.Equal(t, "4: POST http://example.com/graphql", res)

	res, err = do("POST", "http://example.com/graphql", bytes.NewBufferString(`hello2`))
	assert.NoError(t, err)
	assert.Equal(t, "5: POST http://example.com/graphql", res)

	res, err = do("GET", "http://example.com/error", nil)
	assert.NoError(t, err)
	assert.Equal(t, "6: GET http://example.com/error", res)
	res, err = do("GET", "http://example.com/error", nil)
	assert.NoError(t, err)
	assert.Equal(t, "7: GET http://example.com/error", res)
}

func TestRequestCacheOptions(t *testing.T) {
	req, err := http.NewRequest("GET", "some/url", nil)
	assert.NoError(t, err)
	req.Header.Set("X-GH-CACHE-DIR", "some/dir/path")
	req.Header.Set("X-GH-CACHE-TTL", "1h")
	dir, ttl := requestCacheOptions(req)
	assert.Equal(t, dir, "some/dir/path")
	assert.Equal(t, ttl, time.Hour)
}
