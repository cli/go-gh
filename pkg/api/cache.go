package api

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type cache struct {
	dir string
	ttl time.Duration
}

type cacheRoundTripper struct {
	fs fileStorage
	rt http.RoundTripper
}

type fileStorage struct {
	dir string
	ttl time.Duration
	mu  *sync.RWMutex
}

type readCloser struct {
	io.Reader
	io.Closer
}

// isCacheableRequest decides whether a request is eligible for caching. GET
// and HEAD are always eligible. POST is eligible only for the GraphQL endpoint
// AND only when the GraphQL document can be confidently identified as
// containing only query operations (no mutations or subscriptions). The
// GraphQL body inspection is fail-closed: anything ambiguous is treated as
// non-cacheable.
func isCacheableRequest(req *http.Request) bool {
	if strings.EqualFold(req.Method, "GET") || strings.EqualFold(req.Method, "HEAD") {
		return true
	}

	if strings.EqualFold(req.Method, "POST") && (req.URL.Path == "/graphql" || req.URL.Path == "/api/graphql") {
		return isCacheableGraphQLRequest(req)
	}

	return false
}

// isCacheableResponse decides whether a response is safe to persist in the
// cache. Only 2xx responses qualify. Caching non-2xx responses creates several
// problems for callers, especially under aggressive opt-in caching:
//
//   - 401 Unauthorized can hide a token that has just been refreshed.
//   - 404 Not Found can hide newly granted permissions or a freshly created
//     resource.
//   - 429 Too Many Requests would be replayed to the caller for the entire
//     TTL window, making rate-limit recovery worse rather than better.
//   - 5xx responses describe transient server-side failures and should never
//     be served from cache.
//
// The previous policy of "status < 500 && status != 403" excluded 5xx and 403
// but cached the cases above. Narrowing to 2xx is a behavior change for
// existing callers of gh api --cache but moves them in the correct direction.
func isCacheableResponse(res *http.Response) bool {
	return res.StatusCode >= 200 && res.StatusCode < 300
}

func cacheKey(req *http.Request) (string, error) {
	h := sha256.New()
	fmt.Fprintf(h, "%s:", req.Method)
	fmt.Fprintf(h, "%s:", req.URL.String())
	fmt.Fprintf(h, "%s:", req.Header.Get("Accept"))
	fmt.Fprintf(h, "%s:", req.Header.Get("Authorization"))

	if req.Body != nil {
		var bodyCopy io.ReadCloser
		req.Body, bodyCopy = copyStream(req.Body)
		defer bodyCopy.Close()
		if _, err := io.Copy(h, bodyCopy); err != nil {
			return "", err
		}
	}

	digest := h.Sum(nil)
	return fmt.Sprintf("%x", digest), nil
}

func (c cache) RoundTripper(rt http.RoundTripper) http.RoundTripper {
	fs := fileStorage{
		dir: c.dir,
		ttl: c.ttl,
		mu:  &sync.RWMutex{},
	}
	return cacheRoundTripper{fs: fs, rt: rt}
}

func (crt cacheRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqDir, reqTTL, reqTTLSet := requestCacheOptions(req)

	// An explicit X-GH-CACHE-TTL: 0 is a hard bypass for this request, even when
	// the transport has a non-zero global TTL configured. This is the escape hatch
	// for callers that have opted into a global cache (e.g. via a process-wide
	// transport configuration) but want to force a fresh response for a specific
	// request.
	if reqTTLSet && reqTTL == 0 {
		return crt.rt.RoundTrip(req)
	}

	effectiveTTL := crt.fs.ttl
	if reqTTLSet {
		effectiveTTL = reqTTL
	}

	if effectiveTTL == 0 {
		return crt.rt.RoundTrip(req)
	}

	if !isCacheableRequest(req) {
		return crt.rt.RoundTrip(req)
	}

	origDir := crt.fs.dir
	if reqDir != "" {
		crt.fs.dir = reqDir
	}
	origTTL := crt.fs.ttl
	crt.fs.ttl = effectiveTTL

	key, keyErr := cacheKey(req)
	if keyErr == nil {
		if res, err := crt.fs.read(key); err == nil {
			res.Request = req
			return res, nil
		}
	}

	res, err := crt.rt.RoundTrip(req)
	if err == nil && keyErr == nil && isCacheableResponse(res) {
		_ = crt.fs.store(key, res)
	}

	crt.fs.dir = origDir
	crt.fs.ttl = origTTL

	return res, err
}

// requestCacheOptions inspects per-request override headers. The returned ttlSet
// flag indicates whether X-GH-CACHE-TTL was present on the request and parsed to
// a valid duration. This distinguishes "no override" (ttlSet=false) from
// "explicit zero override" (ttlSet=true, ttl=0), the latter of which forces a
// cache bypass even when the transport has a non-zero global TTL configured.
//
// An unparseable X-GH-CACHE-TTL value is treated the same as an absent header
// (ttlSet=false): the caller's intent is ambiguous, so we fall back to whatever
// global TTL is configured rather than silently bypassing or surfacing an error.
func requestCacheOptions(req *http.Request) (dir string, ttl time.Duration, ttlSet bool) {
	dir = req.Header.Get("X-GH-CACHE-DIR")
	if ttlHeader := req.Header.Get("X-GH-CACHE-TTL"); ttlHeader != "" {
		if parsed, err := time.ParseDuration(ttlHeader); err == nil {
			ttl = parsed
			ttlSet = true
		}
	}
	return
}

func (fs *fileStorage) filePath(key string) string {
	if len(key) >= 6 {
		return filepath.Join(fs.dir, key[0:2], key[2:4], key[4:])
	}
	return filepath.Join(fs.dir, key)
}

func (fs *fileStorage) read(key string) (*http.Response, error) {
	cacheFile := fs.filePath(key)

	fs.mu.RLock()
	defer fs.mu.RUnlock()

	f, err := os.Open(cacheFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	age := time.Since(stat.ModTime())
	if age > fs.ttl {
		return nil, errors.New("cache expired")
	}

	body := &bytes.Buffer{}
	_, err = io.Copy(body, f)
	if err != nil {
		return nil, err
	}

	res, err := http.ReadResponse(bufio.NewReader(body), nil)
	return res, err
}

func (fs *fileStorage) store(key string, res *http.Response) (storeErr error) {
	cacheFile := fs.filePath(key)

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if storeErr = os.MkdirAll(filepath.Dir(cacheFile), 0755); storeErr != nil {
		return
	}

	var f *os.File
	if f, storeErr = os.OpenFile(cacheFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); storeErr != nil {
		return
	}

	defer func() {
		if err := f.Close(); storeErr == nil && err != nil {
			storeErr = err
		}
	}()

	var origBody io.ReadCloser
	if res.Body != nil {
		origBody, res.Body = copyStream(res.Body)
		defer res.Body.Close()
	}

	storeErr = res.Write(f)
	if origBody != nil {
		res.Body = origBody
	}

	return
}

func copyStream(r io.ReadCloser) (io.ReadCloser, io.ReadCloser) {
	b := &bytes.Buffer{}
	nr := io.TeeReader(r, b)
	return io.NopCloser(b), &readCloser{
		Reader: nr,
		Closer: r,
	}
}
