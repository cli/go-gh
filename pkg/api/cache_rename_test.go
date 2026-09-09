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
)

func TestCacheRenameRetriesTransientFailure(t *testing.T) {
	t.Parallel()

	// Given a complete replacement that cannot initially be published.
	dir := t.TempDir()
	oldPath, newPath := filepath.Join(dir, "temp"), filepath.Join(dir, "cache")
	require.NoError(t, os.WriteFile(oldPath, []byte("replacement"), 0600))
	require.NoError(t, os.WriteFile(newPath, []byte("previous"), 0600))
	synctest.Test(t, func(t *testing.T) {
		available := time.Now().Add(10 * time.Millisecond)

		// When the rename becomes possible within the retry budget.
		err := retryCacheRename(func() (error, bool) {
			if time.Now().Before(available) {
				return errors.New("file is temporarily in use"), true
			}
			return os.Rename(oldPath, newPath), false
		})

		// Then the completed replacement is published.
		require.NoError(t, err)
		data, err := os.ReadFile(newPath)
		require.NoError(t, err)
		assert.Equal(t, "replacement", string(data))
	})
}

func TestCacheRenameStopsRetryingAfterBudget(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		// Given a rename that remains blocked.
		wantErr := &os.LinkError{Op: "rename", Old: "temp", New: "cache", Err: errors.New("file remains in use")}
		start := time.Now()

		// When its retry budget is exhausted.
		err := retryCacheRename(func() (error, bool) { return wantErr, true })

		// Then the original error is returned within the cache's latency budget.
		require.ErrorIs(t, err, wantErr)
		assert.Equal(t, 100*time.Millisecond, time.Since(start))
	})
}

func TestCacheRenameDoesNotRetryPermanentFailure(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		// Given a rename failure that waiting cannot resolve.
		wantErr := &os.LinkError{Op: "rename", Old: "missing", New: "cache", Err: os.ErrNotExist}
		start := time.Now()

		// When the rename is attempted.
		err := retryCacheRename(func() (error, bool) { return wantErr, false })

		// Then the error is returned without delay.
		require.ErrorIs(t, err, wantErr)
		assert.Zero(t, time.Since(start))
	})
}
