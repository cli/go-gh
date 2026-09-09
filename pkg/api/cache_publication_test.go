package api_test

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	// Then another client can still read the previous entry, without leaked files.
	reader := cacheTestClient(t, dir, time.Hour, cacheTestTransport(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("unexpected cache miss")
	}))
	assert.Equal(t, "previous response", cacheTestFetch(t, reader))
	assert.Len(t, cacheTestFiles(t, dir), 1)
}

func TestCachePublishesOnlyCompleteResponses(t *testing.T) {
	t.Parallel()

	// Given independent clients sharing a cache and a refresh paused mid-body.
	dir := t.TempDir()
	seed := cacheTestClient(t, dir, time.Hour, cacheTestResponse(strings.NewReader("previous response")))
	assert.Equal(t, "previous response", cacheTestFetch(t, seed))
	started := make(chan struct{})
	signalStarted := sync.OnceFunc(func() { close(started) })
	release := make(chan struct{})
	unblock := sync.OnceFunc(func() { close(release) })
	var workers sync.WaitGroup
	t.Cleanup(func() {
		unblock()
		workers.Wait()
	})
	remainder := strings.NewReader("complete")
	refresh := cacheTestClient(t, dir, -time.Second, cacheTestResponse(io.MultiReader(
		strings.NewReader("replacement "),
		cacheTestReader(func(p []byte) (int, error) {
			signalStarted()
			<-release
			return remainder.Read(p)
		}),
	)))
	reader := cacheTestClient(t, dir, time.Hour, cacheTestTransport(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("unexpected cache miss")
	}))
	finished := make(chan error, 1)

	// When another client reads while the refresh is still being written.
	workers.Add(1)
	go func() {
		defer workers.Done()
		res, err := refresh.Get("https://api.github.com/cache-test")
		if err == nil {
			err = res.Body.Close()
		}
		finished <- err
	}()
	<-started
	assert.Equal(t, "previous response", cacheTestFetch(t, reader))
	unblock()
	require.NoError(t, <-finished)

	// Then readers see the complete replacement and no temporary file remains.
	assert.Equal(t, "replacement complete", cacheTestFetch(t, reader))
	assert.Len(t, cacheTestFiles(t, dir), 1)
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
