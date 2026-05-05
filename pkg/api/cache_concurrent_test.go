package api

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestStore_AtomicWrites_NoTornReads verifies that concurrent writes to the
// same cache key never produce a partial / torn file that a reader might
// observe. It uses two independent fileStorage instances pointing at the same
// directory to simulate two processes (each has its own mutex; they share
// only the filesystem). With non-atomic writes (truncate-in-place), readers
// would observe partial bodies whose length disagrees with Content-Length, or
// bodies whose contents mix bytes from two different writes. With atomic
// rename, every successful read returns a complete response from one writer
// or the other, never a hybrid.
//
// Run this test with -race to also catch any in-process synchronization
// regressions that might be introduced if the file storage internals change.
func TestStore_AtomicWrites_NoTornReads(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "cache")
	fs1 := &fileStorage{dir: cacheDir, ttl: time.Hour, mu: &sync.RWMutex{}}
	fs2 := &fileStorage{dir: cacheDir, ttl: time.Hour, mu: &sync.RWMutex{}}

	const key = "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	const bodySize = 64 * 1024

	makeResponse := func(marker byte) *http.Response {
		body := bytes.Repeat([]byte{marker}, bodySize)
		return &http.Response{
			StatusCode:    200,
			Header:        http.Header{},
			Body:          io.NopCloser(bytes.NewReader(body)),
			ContentLength: int64(bodySize),
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
		}
	}

	require.NoError(t, fs1.store(key, makeResponse('A')), "initial store should succeed")

	var (
		stop      = make(chan struct{})
		wg        sync.WaitGroup
		failures  atomic.Int64
		readsOK   atomic.Int64
		writesOK  atomic.Int64
		readsErr  atomic.Int64
		writesErr atomic.Int64
	)

	writer := func(fs *fileStorage, marker byte) {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			if err := fs.store(key, makeResponse(marker)); err != nil {
				writesErr.Add(1)
			} else {
				writesOK.Add(1)
			}
		}
	}

	reader := func(fs *fileStorage) {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			res, err := fs.read(key)
			if err != nil {
				readsErr.Add(1)
				continue
			}
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			readsOK.Add(1)

			if int64(len(body)) != res.ContentLength {
				failures.Add(1)
				t.Errorf("torn read: body length %d != ContentLength %d", len(body), res.ContentLength)
				return
			}
			if len(body) > 0 {
				first := body[0]
				for _, b := range body {
					if b != first {
						failures.Add(1)
						t.Errorf("torn read: body contains mixed marker bytes (saw %q and %q)", first, b)
						return
					}
				}
			}
		}
	}

	wg.Add(1)
	go writer(fs1, 'B')
	wg.Add(1)
	go writer(fs2, 'C')
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go reader(fs1)
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go reader(fs2)
	}

	time.Sleep(300 * time.Millisecond)
	close(stop)
	wg.Wait()

	require.Zero(t, failures.Load(), "torn reads observed")
	require.Greater(t, writesOK.Load(), int64(0), "test did not exercise any successful writes")
	require.Greater(t, readsOK.Load(), int64(0), "test did not exercise any successful reads")
	t.Logf("successful: %d writes, %d reads; errors: %d writes, %d reads",
		writesOK.Load(), readsOK.Load(), writesErr.Load(), readsErr.Load())
}
