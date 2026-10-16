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

func isCacheableRequest(req *http.Request) bool {
	if strings.EqualFold(req.Method, "GET") || strings.EqualFold(req.Method, "HEAD") {
		return true
	}

	if strings.EqualFold(req.Method, "POST") && (req.URL.Path == "/graphql" || req.URL.Path == "/api/graphql") {
		return true
	}

	return false
}

func isCacheableResponse(res *http.Response) bool {
	return res.StatusCode < 500 && res.StatusCode != 403
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
	reqDir, reqTTL := requestCacheOptions(req)

	if crt.fs.ttl == 0 && reqTTL == 0 {
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
	if reqTTL != 0 {
		crt.fs.ttl = reqTTL
	}

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

// Allow an individual request to override cache options.
func requestCacheOptions(req *http.Request) (string, time.Duration) {
	var dur time.Duration
	// Added alongside the TTL header in https://github.com/cli/go-gh/pull/49.
	// No production consumer of the directory override is known; retain it for
	// compatibility.
	dir := req.Header.Get("X-GH-CACHE-DIR")
	ttl := req.Header.Get("X-GH-CACHE-TTL")
	if ttl != "" {
		dur, _ = time.ParseDuration(ttl)
	}
	return dir, dur
}

func (fs *fileStorage) filePath(key string) string {
	if len(key) >= 6 {
		return filepath.Join(fs.dir, key[0:2], key[2:4], key[4:])
	}
	return filepath.Join(fs.dir, key)
}

// checksumPath returns the path of the integrity sidecar accompanying a cache entry.
func (fs *fileStorage) checksumPath(key string) string {
	return fs.filePath(key) + ".sha256"
}

func (fs *fileStorage) read(key string) (*http.Response, error) {
	cacheFile := fs.filePath(key)

	fs.mu.RLock()
	defer fs.mu.RUnlock()

	f, err := os.Open(cacheFile)
	if err != nil {
		// No entry to serve. Drop any orphaned checksum left behind by an
		// interrupted publish so such files cannot accumulate.
		_ = os.Remove(fs.checksumPath(key))
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

	if err := fs.verifyChecksum(key, body.Bytes()); err != nil {
		// Never serve an entry that fails integrity verification. Evict it so
		// subsequent requests refetch instead of re-reading known-bad bytes.
		// This also heals entries corrupted before checksums existed.
		_ = os.Remove(cacheFile)
		return nil, err
	}

	res, err := http.ReadResponse(bufio.NewReader(body), nil)
	return res, err
}

func (fs *fileStorage) store(key string, res *http.Response) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	cacheFilePath := fs.filePath(key)
	dir := filepath.Dir(cacheFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Finish writing before publishing the entry. Same-directory rename gives
	// atomic replacement on Unix; it is best-effort on other platforms.
	tmpCacheFile, err := os.CreateTemp(dir, ".gh-cache-*")
	if err != nil {
		return err
	}
	tmpCacheFileName := tmpCacheFile.Name()
	tmpChecksumFileName := tmpCacheFileName + ".sha256"
	// Clean up on errors and panics too. After a successful rename, the
	// temporary paths no longer exist, so removing them is harmless.
	defer func() {
		_ = tmpCacheFile.Close()
		_ = os.Remove(tmpCacheFileName)
		_ = os.Remove(tmpChecksumFileName)
	}()

	if err := writeCacheResponse(tmpCacheFile, res); err != nil {
		return err
	}
	if err := tmpCacheFile.Close(); err != nil {
		return err
	}
	if err := writeCacheChecksum(tmpCacheFileName, tmpChecksumFileName); err != nil {
		return err
	}

	// Publish the checksum first so readers never observe an entry without
	// one. An entry without a checksum is treated as corrupt on read.
	if err := renameCacheFile(tmpChecksumFileName, fs.checksumPath(key)); err != nil {
		return err
	}
	if err := renameCacheFile(tmpCacheFileName, cacheFilePath); err != nil {
		// Roll back the just-published checksum so a failed publication
		// leaves no trace behind.
		_ = os.Remove(fs.checksumPath(key))
		return err
	}
	return nil
}

// writeCacheChecksum records a SHA-256 digest of a freshly written cache entry
// in a sidecar file, so readers can detect entries corrupted after publication
// (for example by pre-atomic-publish races, crashes, or external modification).
func writeCacheChecksum(tmpCacheFileName, tmpChecksumFileName string) error {
	digest, err := fileSHA256(tmpCacheFileName)
	if err != nil {
		return err
	}
	return os.WriteFile(tmpChecksumFileName, []byte(digest), 0644)
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// verifyChecksum reports whether content matches the digest recorded when the
// entry was published. A missing digest (entries written before checksums
// existed) fails verification so such entries are refreshed instead of served.
func (fs *fileStorage) verifyChecksum(key string, content []byte) error {
	expected, err := os.ReadFile(fs.checksumPath(key))
	if err != nil {
		return fmt.Errorf("cache checksum missing: %w", err)
	}
	digest := sha256.Sum256(content)
	if fmt.Sprintf("%x", digest) != string(expected) {
		return errors.New("cache checksum mismatch")
	}
	return nil
}

func writeCacheResponse(w io.Writer, res *http.Response) error {
	if res.Body == nil {
		// Serialize the HTTP response headers only, since there is no body.
		return res.Write(w)
	}

	// Buffer the bytes consumed during serialization so the caller can replay
	// them. Restore the replay reader even if writing fails or panics.
	buffer := &bytes.Buffer{}
	recorder := &errorRecordingReader{Reader: io.TeeReader(res.Body, buffer)}
	source := &readCloser{Reader: recorder, Closer: res.Body}
	res.Body = source
	defer source.Close()
	defer func() {
		res.Body = io.NopCloser(&errorReplayingReader{Reader: buffer, err: recorder.err})
	}()

	return res.Write(w)
}

type errorRecordingReader struct {
	io.Reader
	err error
}

func (r *errorRecordingReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err != nil && err != io.EOF {
		r.err = err
	}
	return n, err
}

type errorReplayingReader struct {
	io.Reader
	err error
}

func (r *errorReplayingReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err == io.EOF && r.err != nil {
		err = r.err
		r.err = nil
	}
	return n, err
}

func copyStream(body io.ReadCloser) (replay, source io.ReadCloser) {
	buffer := &bytes.Buffer{}
	return io.NopCloser(buffer), &readCloser{
		Reader: io.TeeReader(body, buffer),
		Closer: body,
	}
}
