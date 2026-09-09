package api

import (
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
