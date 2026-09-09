package api_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		path       string
		nextPath   string
		body       string
		nextBody   string
		status     int
		wantSecond string
	}{
		{name: "GET reuses response", method: "GET", path: "/path", nextPath: "/path", status: 200, wantSecond: "response 1"},
		{name: "different URL misses", method: "GET", path: "/path", nextPath: "/other", status: 200, wantSecond: "response 2"},
		{name: "ordinary POST bypasses cache", method: "POST", path: "/path", nextPath: "/path", body: "hello", nextBody: "hello", status: 200, wantSecond: "response 2"},
		{name: "GraphQL POST reuses response", method: "POST", path: "/graphql", nextPath: "/graphql", body: "hello", nextBody: "hello", status: 200, wantSecond: "response 1"},
		{name: "different GraphQL body misses", method: "POST", path: "/graphql", nextPath: "/graphql", body: "hello", nextBody: "hello2", status: 200, wantSecond: "response 2"},
		{name: "server error is not cached", method: "GET", path: "/path", nextPath: "/path", status: 500, wantSecond: "response 2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Given a caching client whose upstream responses distinguish cache misses.
			counter := 0
			client := cacheTestClient(t, t.TempDir(), time.Hour, cacheTestTransport(func(*http.Request) (*http.Response, error) {
				counter++
				return &http.Response{
					StatusCode: tt.status,
					Body:       io.NopCloser(strings.NewReader(fmt.Sprintf("response %d", counter))),
				}, nil
			}))
			do := func(path, body string) string {
				req, err := http.NewRequest(tt.method, "https://api.github.com"+path, strings.NewReader(body))
				require.NoError(t, err)
				res, err := client.Do(req)
				require.NoError(t, err)
				defer res.Body.Close()
				data, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				return string(data)
			}
			assert.Equal(t, "response 1", do(tt.path, tt.body))

			// When the next request is made.
			got := do(tt.nextPath, tt.nextBody)

			// Then it reuses or bypasses the entry according to the scenario.
			assert.Equal(t, tt.wantSecond, got)
		})
	}
}

func TestCacheRequestTTLEnablesCaching(t *testing.T) {
	t.Parallel()

	// Given a client with caching disabled by default.
	dir := t.TempDir()
	client, err := api.NewHTTPClient(api.ClientOptions{
		Host:         "github.com",
		AuthToken:    "token",
		CacheDir:     dir,
		LogIgnoreEnv: true,
		Transport:    cacheTestResponse(strings.NewReader("cached response")),
	})
	require.NoError(t, err)
	req, err := http.NewRequest("GET", "https://api.github.com/cache-test", nil)
	require.NoError(t, err)
	req.Header.Set("X-GH-CACHE-TTL", "1h")

	// When a request opts into caching.
	res, err := client.Do(req)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	// Then an independent client can reuse the published response.
	reader := cacheTestClient(t, dir, time.Hour, cacheTestTransport(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("unexpected cache miss")
	}))
	assert.Equal(t, "cached response", cacheTestFetch(t, reader))
}

func TestCacheRequestTTLOverridesClientTTL(t *testing.T) {
	t.Parallel()

	// Given a cached response and a client whose default TTL treats it as expired.
	dir := t.TempDir()
	seed := cacheTestClient(t, dir, time.Hour, cacheTestResponse(strings.NewReader("cached response")))
	assert.Equal(t, "cached response", cacheTestFetch(t, seed))
	client := cacheTestClient(t, dir, -time.Second, cacheTestResponse(strings.NewReader("fresh response")))
	req, err := http.NewRequest("GET", "https://api.github.com/cache-test", nil)
	require.NoError(t, err)
	req.Header.Set("X-GH-CACHE-TTL", "1h")

	// When a request specifies a longer TTL.
	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	// Then that request uses the cache, but a request without the header refreshes it.
	assert.Equal(t, "cached response", string(body))
	assert.Equal(t, "fresh response", cacheTestFetch(t, client))
}

func TestCachePublishesResponseWithoutBody(t *testing.T) {
	t.Parallel()

	// Given a response with headers but no body.
	dir := t.TempDir()
	client := cacheTestClient(t, dir, time.Hour, cacheTestTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Header:     http.Header{"Etag": {`"v1"`}},
		}, nil
	}))
	res, err := client.Get("https://api.github.com/cache-test")
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	reader := cacheTestClient(t, dir, time.Hour, cacheTestTransport(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("unexpected cache miss")
	}))

	// When another client requests the cached response.
	cached, err := reader.Get("https://api.github.com/cache-test")
	require.NoError(t, err)
	defer cached.Body.Close()

	// Then its status and headers are preserved, with an empty body.
	assert.Equal(t, http.StatusNoContent, cached.StatusCode)
	assert.Equal(t, `"v1"`, cached.Header.Get("Etag"))
	body, err := io.ReadAll(cached.Body)
	require.NoError(t, err)
	assert.Empty(t, body)
}

func TestCachePreservesEntryAfterFailedRefresh(t *testing.T) {
	t.Parallel()

	// Given a cached response and a refresh whose body fails partway through.
	dir := t.TempDir()
	seed := cacheTestClient(t, dir, time.Hour, cacheTestResponse(strings.NewReader("previous response")))
	assert.Equal(t, "previous response", cacheTestFetch(t, seed))
	refresh := cacheTestClient(t, dir, -time.Second, cacheTestResponse(io.MultiReader(
		strings.NewReader("incomplete response"),
		cacheTestReader(func([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }),
	)))

	// When writing the refresh fails.
	res, err := refresh.Get("https://api.github.com/cache-test")
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	// Then another client can still read the previous entry.
	reader := cacheTestClient(t, dir, time.Hour, cacheTestTransport(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("unexpected cache miss")
	}))
	assert.Equal(t, "previous response", cacheTestFetch(t, reader))
}

func TestCachePublishesOnlyCompleteResponses(t *testing.T) {
	t.Parallel()

	// Given independent clients sharing a cache and a refresh paused mid-body.
	dir := t.TempDir()
	seed := cacheTestClient(t, dir, time.Hour, cacheTestResponse(strings.NewReader("previous response")))
	assert.Equal(t, "previous response", cacheTestFetch(t, seed))
	body, writer := io.Pipe()
	refresh := cacheTestClient(t, dir, -time.Second, cacheTestResponse(body))
	reader := cacheTestClient(t, dir, time.Hour, cacheTestTransport(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("unexpected cache miss")
	}))
	finished := make(chan error, 1)
	t.Cleanup(func() {
		_ = writer.Close()
		_ = body.Close()
		for range finished {
		}
	})

	// When another client reads while the refresh is still being written.
	go func() {
		defer close(finished)
		defer body.Close()
		res, err := refresh.Get("https://api.github.com/cache-test")
		if err == nil {
			err = res.Body.Close()
		}
		finished <- err
	}()
	_, err := io.WriteString(writer, "replacement ")
	require.NoError(t, err)
	assert.Equal(t, "previous response", cacheTestFetch(t, reader))

	// Finish the replacement and wait for publication.
	_, err = io.WriteString(writer, "complete")
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	require.NoError(t, <-finished)

	// Then readers see the complete replacement.
	assert.Equal(t, "replacement complete", cacheTestFetch(t, reader))
}

func TestCacheCleansUpAfterBodyPanic(t *testing.T) {
	t.Parallel()

	// Given a response body that panics while being cached.
	dir := t.TempDir()
	client := cacheTestClient(t, dir, time.Hour, cacheTestTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(cacheTestReader(func([]byte) (int, error) {
				panic("body read failed")
			})),
		}, nil
	}))

	// When the panic propagates to the caller.
	assert.PanicsWithValue(t, "body read failed", func() {
		_, _ = client.Get("https://api.github.com/cache-test")
	})

	// Then no partial cache entry or temporary file remains.
	assert.Empty(t, cacheTestFiles(t, dir))
}

func TestCachePublicationFailureDoesNotFailResponse(t *testing.T) {
	t.Parallel()

	// Given a cache destination that cannot be replaced by a file.
	dir := t.TempDir()
	seed := cacheTestClient(t, dir, time.Hour, cacheTestResponse(strings.NewReader("previous response")))
	assert.Equal(t, "previous response", cacheTestFetch(t, seed))
	files := cacheTestFiles(t, dir)
	require.Len(t, files, 1)
	require.NoError(t, os.Remove(files[0]))
	require.NoError(t, os.Mkdir(files[0], 0700))
	client := cacheTestClient(t, dir, time.Hour, cacheTestResponse(strings.NewReader("fresh response")))

	synctest.Test(t, func(t *testing.T) {
		// When the HTTP request succeeds but cache publication fails.
		body := cacheTestFetch(t, client)

		// Then the caller still receives the response and no temporary file leaks.
		assert.Equal(t, "fresh response", body)
		assert.Empty(t, cacheTestFiles(t, dir))
	})
}

func cacheTestClient(t *testing.T, dir string, ttl time.Duration, transport http.RoundTripper) *http.Client {
	t.Helper()
	client, err := api.NewHTTPClient(api.ClientOptions{
		Host:         "github.com",
		AuthToken:    "token",
		CacheDir:     dir,
		CacheTTL:     ttl,
		EnableCache:  true,
		LogIgnoreEnv: true,
		Transport:    transport,
	})
	require.NoError(t, err)
	return client
}

func cacheTestFiles(t *testing.T, dir string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	require.NoError(t, err)
	return files
}

func cacheTestResponse(body io.Reader) http.RoundTripper {
	return cacheTestTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(body),
		}, nil
	})
}

func cacheTestFetch(t *testing.T, client *http.Client) string {
	t.Helper()
	res, err := client.Get("https://api.github.com/cache-test")
	require.NoError(t, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return string(body)
}

type cacheTestTransport func(*http.Request) (*http.Response, error)

func (f cacheTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type cacheTestReader func([]byte) (int, error)

func (f cacheTestReader) Read(p []byte) (int, error) {
	return f(p)
}
