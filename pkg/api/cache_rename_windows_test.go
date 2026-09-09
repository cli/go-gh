package api

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestCacheRenameStopsRetryingAfterBudget(t *testing.T) {
	t.Parallel()

	// Given a destination whose reader prevents replacement for the entire retry budget.
	dir := t.TempDir()
	oldPath, newPath := filepath.Join(dir, "temp"), filepath.Join(dir, "cache")
	require.NoError(t, os.WriteFile(oldPath, []byte("replacement"), 0600))
	require.NoError(t, os.WriteFile(newPath, []byte("previous"), 0600))
	reader, err := os.Open(newPath)
	require.NoError(t, err)
	defer reader.Close()

	synctest.Test(t, func(t *testing.T) {
		start := time.Now()

		// When publication is attempted while the reader remains open.
		err := renameCacheFile(oldPath, newPath)

		// Then it returns the sharing error within budget, leaving both files intact.
		require.Error(t, err)
		assert.True(t, errors.Is(err, windows.ERROR_SHARING_VIOLATION) || errors.Is(err, windows.ERROR_ACCESS_DENIED),
			"expected a Windows sharing conflict, got %v", err)
		elapsed := time.Since(start)
		assert.Greater(t, elapsed, time.Duration(0), "expected retries before giving up")
		assert.LessOrEqual(t, elapsed, 100*time.Millisecond, "cache retries must stay within the latency budget")
		previous, err := os.ReadFile(newPath)
		require.NoError(t, err)
		assert.Equal(t, "previous", string(previous))
		replacement, err := os.ReadFile(oldPath)
		require.NoError(t, err)
		assert.Equal(t, "replacement", string(replacement))
	})
}

func TestCacheRenameWaitsForWindowsReader(t *testing.T) {
	t.Parallel()

	// Given a destination held open without delete sharing by another reader.
	dir := t.TempDir()
	oldPath, newPath := filepath.Join(dir, "temp"), filepath.Join(dir, "cache")
	require.NoError(t, os.WriteFile(oldPath, []byte("replacement"), 0600))
	require.NoError(t, os.WriteFile(newPath, []byte("previous"), 0600))
	reader, err := os.Open(newPath)
	require.NoError(t, err)
	defer reader.Close()

	synctest.Test(t, func(t *testing.T) {
		// When the reader closes while the rename is waiting to retry.
		done := make(chan error, 1)
		go func() { done <- renameCacheFile(oldPath, newPath) }()
		synctest.Wait()
		require.NoError(t, reader.Close())

		// Then the replacement succeeds.
		require.NoError(t, <-done)
		data, err := os.ReadFile(newPath)
		require.NoError(t, err)
		assert.Equal(t, "replacement", string(data))
	})
}

func TestCacheRenameReturnsMissingSourceImmediately(t *testing.T) {
	t.Parallel()

	// Given a source file that does not exist.
	dir := t.TempDir()
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()

		// When it is renamed.
		err := renameCacheFile(filepath.Join(dir, "missing"), filepath.Join(dir, "cache"))

		// Then the non-transient Windows error is returned without retrying.
		require.ErrorIs(t, err, os.ErrNotExist)
		assert.Zero(t, time.Since(start))
	})
}
