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

	res, err = do("POST", "http://example.com/graphql", bytes.NewBufferString(`{"query":"query Q1 { viewer { login } }"}`))
	assert.NoError(t, err)
	assert.Equal(t, "4: POST http://example.com/graphql", res)
	res, err = do("POST", "http://example.com/graphql", bytes.NewBufferString(`{"query":"query Q1 { viewer { login } }"}`))
	assert.NoError(t, err)
	assert.Equal(t, "4: POST http://example.com/graphql", res)

	res, err = do("POST", "http://example.com/graphql", bytes.NewBufferString(`{"query":"query Q2 { viewer { login } }"}`))
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

	res, err = do("POST", "http://example.com/graphql", bytes.NewBufferString(`{"query":"query Q1 { viewer { login } }"}`))
	assert.NoError(t, err)
	assert.Equal(t, "4: POST http://example.com/graphql", res)
	res, err = do("POST", "http://example.com/graphql", bytes.NewBufferString(`{"query":"query Q1 { viewer { login } }"}`))
	assert.NoError(t, err)
	assert.Equal(t, "4: POST http://example.com/graphql", res)

	res, err = do("POST", "http://example.com/graphql", bytes.NewBufferString(`{"query":"query Q2 { viewer { login } }"}`))
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
	t.Run("both headers set", func(t *testing.T) {
		req, err := http.NewRequest("GET", "some/url", nil)
		assert.NoError(t, err)
		req.Header.Set("X-GH-CACHE-DIR", "some/dir/path")
		req.Header.Set("X-GH-CACHE-TTL", "1h")
		dir, ttl, ttlSet := requestCacheOptions(req)
		assert.Equal(t, "some/dir/path", dir)
		assert.Equal(t, time.Hour, ttl)
		assert.True(t, ttlSet)
	})

	t.Run("explicit zero TTL is set", func(t *testing.T) {
		req, err := http.NewRequest("GET", "some/url", nil)
		assert.NoError(t, err)
		req.Header.Set("X-GH-CACHE-TTL", "0")
		_, ttl, ttlSet := requestCacheOptions(req)
		assert.Equal(t, time.Duration(0), ttl)
		assert.True(t, ttlSet, "explicit zero must report ttlSet=true so RoundTrip can distinguish bypass from default")
	})

	t.Run("absent header is not set", func(t *testing.T) {
		req, err := http.NewRequest("GET", "some/url", nil)
		assert.NoError(t, err)
		_, ttl, ttlSet := requestCacheOptions(req)
		assert.Equal(t, time.Duration(0), ttl)
		assert.False(t, ttlSet)
	})

	t.Run("unparseable header is treated as not set", func(t *testing.T) {
		req, err := http.NewRequest("GET", "some/url", nil)
		assert.NoError(t, err)
		req.Header.Set("X-GH-CACHE-TTL", "not-a-duration")
		_, ttl, ttlSet := requestCacheOptions(req)
		assert.Equal(t, time.Duration(0), ttl)
		assert.False(t, ttlSet, "unparseable values fall back to global TTL rather than silently bypassing")
	})
}

// TestCacheRoundTrip_PerRequestOptOut verifies that an explicit
// X-GH-CACHE-TTL: 0 header forces a network call even when the transport has
// a non-zero global TTL configured. This is the escape hatch that makes a
// process-wide cache safe to opt into: callers can always force fresh data
// for individual requests.
func TestCacheRoundTrip_PerRequestOptOut(t *testing.T) {
	counter := 0
	fakeHTTP := tripper{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			counter += 1
			body := fmt.Sprintf("%d: %s %s", counter, req.Method, req.URL.String())
			return &http.Response{
				StatusCode: 200,
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
			CacheTTL:     time.Hour,
			CacheDir:     cacheDir,
			LogIgnoreEnv: true,
		},
	)
	assert.NoError(t, err)

	do := func(method, url string, ttlOverride string) (string, error) {
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return "", err
		}
		if ttlOverride != "" {
			req.Header.Set("X-GH-CACHE-TTL", ttlOverride)
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

	res, err := do("GET", "http://example.com/path", "")
	assert.NoError(t, err)
	assert.Equal(t, "1: GET http://example.com/path", res, "first request populates the cache")

	res, err = do("GET", "http://example.com/path", "")
	assert.NoError(t, err)
	assert.Equal(t, "1: GET http://example.com/path", res, "second request without override returns cached body")

	res, err = do("GET", "http://example.com/path", "0")
	assert.NoError(t, err)
	assert.Equal(t, "2: GET http://example.com/path", res, "X-GH-CACHE-TTL: 0 must force a fresh network call even though global TTL is 1h")

	// The bypass response must not be stored, so the next default request still
	// returns the originally cached body, not the bypass body.
	res, err = do("GET", "http://example.com/path", "")
	assert.NoError(t, err)
	assert.Equal(t, "1: GET http://example.com/path", res, "X-GH-CACHE-TTL: 0 must not overwrite the cache entry")
}

// TestCacheKey_VariesByHeaders verifies that headers known to affect the
// representation of a GitHub API response produce distinct cache keys, so a
// caller varying one of these headers does not get served a cached response
// generated for a different representation.
func TestCacheKey_VariesByHeaders(t *testing.T) {
	baseHeaders := map[string]string{
		"Accept":               "application/vnd.github+json",
		"Accept-Encoding":      "gzip",
		"Authorization":        "token a",
		"X-GitHub-Api-Version": "2022-11-28",
		"GraphQL-Features":     "feature-a",
		"Time-Zone":            "UTC",
	}

	makeReq := func(headerOverride, value string) *http.Request {
		req, err := http.NewRequest("GET", "https://api.github.com/repos/cli/cli", nil)
		assert.NoError(t, err)
		for k, v := range baseHeaders {
			req.Header.Set(k, v)
		}
		if headerOverride != "" {
			req.Header.Set(headerOverride, value)
		}
		return req
	}

	baseKey, err := cacheKey(makeReq("", ""))
	assert.NoError(t, err)

	for _, header := range []string{
		"Accept",
		"Accept-Encoding",
		"Authorization",
		"X-GitHub-Api-Version",
		"GraphQL-Features",
		"Time-Zone",
	} {
		t.Run("varies by "+header, func(t *testing.T) {
			altKey, err := cacheKey(makeReq(header, "different-value"))
			assert.NoError(t, err)
			assert.NotEqual(t, baseKey, altKey, "key should change when %s differs", header)
		})
	}

	t.Run("identical requests produce identical keys", func(t *testing.T) {
		dupKey, err := cacheKey(makeReq("", ""))
		assert.NoError(t, err)
		assert.Equal(t, baseKey, dupKey)
	})
}

// TestIsCacheableResponse documents the policy that only 2xx responses are
// persisted to the cache, especially the explicit exclusion of 401, 404, 429,
// 304, and 5xx, all of which would create poor outcomes if replayed for the
// duration of a TTL.
func TestIsCacheableResponse(t *testing.T) {
	cases := []struct {
		status   int
		expected bool
	}{
		{200, true},
		{201, true},
		{204, true},
		{299, true},
		{300, false},
		{301, false},
		{304, false},
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{409, false},
		{422, false},
		{429, false},
		{500, false},
		{502, false},
		{503, false},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("status %d", c.status), func(t *testing.T) {
			got := isCacheableResponse(&http.Response{StatusCode: c.status})
			assert.Equal(t, c.expected, got)
		})
	}
}
