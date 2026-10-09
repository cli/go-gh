package api_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cli/go-gh/v2/internal/testutils"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheResponse(t *testing.T) {
	testutils.StubConfig(t, "")

	counter := 0
	fakeHTTP := cacheTestTransport(func(req *http.Request) (*http.Response, error) {
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
	})

	cacheDir := filepath.Join(t.TempDir(), "gh-cli-cache")

	httpClient, err := api.NewHTTPClient(
		api.ClientOptions{
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
	testutils.StubConfig(t, "")

	counter := 0
	fakeHTTP := cacheTestTransport(func(req *http.Request) (*http.Response, error) {
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
	})

	cacheDir := filepath.Join(t.TempDir(), "gh-cli-cache")

	httpClient, err := api.NewHTTPClient(
		api.ClientOptions{
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
	body, err := io.ReadAll(res.Body)
	assert.Equal(t, "incomplete response", string(body))
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
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
	var (
		workers    sync.WaitGroup
		refreshErr error
	)
	t.Cleanup(func() {
		_ = writer.Close()
		_ = body.Close()
		workers.Wait()
	})

	// When another client reads while the refresh is still being written.
	workers.Go(func() {
		defer body.Close()
		// This refresh is reading from the pipe, simulating a response body that is being written concurrently.
		// Since the pipe will need to EOF, the refresh will block until the writer is closed.
		res, err := refresh.Get("https://api.github.com/cache-test")
		if err == nil {
			err = res.Body.Close()
		}
		refreshErr = err
	})
	_, err := io.WriteString(writer, "replacement ")
	require.NoError(t, err)
	// At this point, the refresh has not finished writing, so the reader should still see the previous response.
	assert.Equal(t, "previous response", cacheTestFetch(t, reader))

	// Finish the replacement and wait for publication.
	_, err = io.WriteString(writer, "complete")
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	workers.Wait()
	require.NoError(t, refreshErr)

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
	require.Len(t, files, 2)
	entry := cacheTestEntryFile(t, files)
	require.NoError(t, os.Remove(entry))
	require.NoError(t, os.Mkdir(entry, 0700))
	client := cacheTestClient(t, dir, time.Hour, cacheTestResponse(strings.NewReader("fresh response")))

	// When the HTTP request succeeds but cache publication fails.
	body := cacheTestFetch(t, client)

	// Then the caller still receives the response and no temporary file leaks.
	assert.Equal(t, "fresh response", body)
	assert.Empty(t, cacheTestFiles(t, dir))
}

func TestCacheEvictsEntryWithCorruptBody(t *testing.T) {
	t.Parallel()

	// Given a cached response.
	dir := t.TempDir()
	counter := 0
	transport := cacheTestTransport(func(*http.Request) (*http.Response, error) {
		counter += 1
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(fmt.Sprintf("response %d", counter))),
		}, nil
	})
	client := cacheTestClient(t, dir, time.Hour, transport)
	assert.Equal(t, "response 1", cacheTestFetch(t, client))
	assert.Equal(t, "response 1", cacheTestFetch(t, client))

	// When the cached entry is corrupted on disk, as produced by the
	// pre-atomic-publish races (cli/go-gh#252, cli/cli#14394). Appending
	// trailing bytes keeps the HTTP framing parseable, so the failure
	// exercises checksum verification specifically.
	files := cacheTestFiles(t, dir)
	require.Len(t, files, 2)
	entry := cacheTestEntryFile(t, files)
	content, err := os.ReadFile(entry)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(entry, append(content, []byte("corrupt")...), 0644))

	// Then the corrupt entry is evicted and the response is refetched...
	assert.Equal(t, "response 2", cacheTestFetch(t, client))
	// ...and the healed entry serves subsequent requests.
	assert.Equal(t, "response 2", cacheTestFetch(t, client))
	require.Len(t, cacheTestFiles(t, dir), 2)
}

func TestCacheEvictsLegacyEntryWithoutChecksum(t *testing.T) {
	t.Parallel()

	// Given a cached response written before checksums existed.
	dir := t.TempDir()
	counter := 0
	transport := cacheTestTransport(func(*http.Request) (*http.Response, error) {
		counter += 1
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(fmt.Sprintf("response %d", counter))),
		}, nil
	})
	client := cacheTestClient(t, dir, time.Hour, transport)
	assert.Equal(t, "response 1", cacheTestFetch(t, client))

	// When the checksum sidecar is missing.
	for _, f := range cacheTestFiles(t, dir) {
		if strings.HasSuffix(f, ".sha256") {
			require.NoError(t, os.Remove(f))
		}
	}

	// Then the entry is refreshed instead of served...
	assert.Equal(t, "response 2", cacheTestFetch(t, client))
	// ...and the healed entry serves subsequent requests.
	assert.Equal(t, "response 2", cacheTestFetch(t, client))
	require.Len(t, cacheTestFiles(t, dir), 2)
}

func cacheTestClient(t *testing.T, dir string, ttl time.Duration, transport http.RoundTripper) *http.Client {
	t.Helper()
	client, err := api.NewHTTPClient(api.ClientOptions{
		Host:         "github.com",
		APIHost:      "api.github.com",
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

// cacheTestEntryFile returns the cache entry file from a cache directory
// listing, skipping integrity sidecars.
func cacheTestEntryFile(t *testing.T, files []string) string {
	t.Helper()
	for _, f := range files {
		if !strings.HasSuffix(f, ".sha256") {
			return f
		}
	}
	t.Fatal("no cache entry file found")
	return ""
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
