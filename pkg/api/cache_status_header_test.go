package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCacheStatusHeader_TransitionsAcrossRequests exercises the full
// hit / miss / expired lifecycle of the X-GH-Cache-Status header. The most
// important property: a "miss" response that gets stored to disk must NOT
// later be served back as "miss" on a subsequent hit. The header is set on
// the in-memory response only, never persisted with the wire bytes.
func TestCacheStatusHeader_TransitionsAcrossRequests(t *testing.T) {
	counter := 0
	fakeHTTP := tripper{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			counter++
			body := fmt.Sprintf("response %d", counter)
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{},
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
			CacheTTL:     50 * time.Millisecond,
			CacheDir:     cacheDir,
			LogIgnoreEnv: true,
		},
	)
	assert.NoError(t, err)

	doGET := func(url string) (*http.Response, string) {
		req, err := http.NewRequest("GET", url, nil)
		assert.NoError(t, err)
		res, err := httpClient.Do(req)
		assert.NoError(t, err)
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		return res, string(body)
	}

	t.Run("first request is a miss", func(t *testing.T) {
		res, body := doGET("http://example.com/path")
		assert.Equal(t, "miss", res.Header.Get(CacheStatusHeader))
		assert.Equal(t, "response 1", body)
	})

	t.Run("second request within TTL is a hit", func(t *testing.T) {
		res, body := doGET("http://example.com/path")
		assert.Equal(t, "hit", res.Header.Get(CacheStatusHeader))
		assert.Equal(t, "response 1", body, "hit must serve the originally stored body")
	})

	t.Run("hit response does not carry a stored miss header", func(t *testing.T) {
		// A second hit specifically asserts the persisted bytes do not
		// include any X-GH-Cache-Status header from the original miss store.
		// If the header had been persisted, this would observe "miss" instead
		// of "hit".
		res, _ := doGET("http://example.com/path")
		assert.Equal(t, "hit", res.Header.Get(CacheStatusHeader))
	})

	t.Run("after TTL expires, request is expired", func(t *testing.T) {
		time.Sleep(100 * time.Millisecond)
		res, body := doGET("http://example.com/path")
		assert.Equal(t, "expired", res.Header.Get(CacheStatusHeader))
		assert.NotEqual(t, "response 1", body, "expired must re-fetch from network")
	})

	t.Run("subsequent request after expired refetch is a hit", func(t *testing.T) {
		res, _ := doGET("http://example.com/path")
		assert.Equal(t, "hit", res.Header.Get(CacheStatusHeader))
	})
}

// TestCacheStatusHeader_BypassPathsHaveNoHeader documents the contract that a
// response without an X-GH-Cache-Status header was not consulted by the cache
// at all (uncacheable method, GraphQL mutation, explicit per-request opt-out,
// or no global TTL configured).
func TestCacheStatusHeader_BypassPathsHaveNoHeader(t *testing.T) {
	fakeHTTP := tripper{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{},
				Body:       io.NopCloser(bytes.NewBufferString("ok")),
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
			CacheTTL:     time.Hour,
			CacheDir:     cacheDir,
			LogIgnoreEnv: true,
		},
	)
	assert.NoError(t, err)

	do := func(req *http.Request) *http.Response {
		res, err := httpClient.Do(req)
		assert.NoError(t, err)
		_, _ = io.ReadAll(res.Body)
		res.Body.Close()
		return res
	}

	t.Run("uncacheable method has no header", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", "http://example.com/x", strings.NewReader("body"))
		res := do(req)
		assert.Equal(t, "", res.Header.Get(CacheStatusHeader))
	})

	t.Run("graphql mutation has no header", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://example.com/graphql",
			strings.NewReader(`{"query":"mutation M { x { id } }"}`))
		res := do(req)
		assert.Equal(t, "", res.Header.Get(CacheStatusHeader))
	})

	t.Run("explicit per-request opt-out has no header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/y", nil)
		req.Header.Set("X-GH-CACHE-TTL", "0")
		res := do(req)
		assert.Equal(t, "", res.Header.Get(CacheStatusHeader))
	})
}
