package api

import (
	"os"
	"time"
)

func renameCacheFile(oldPath, newPath string) error {
	return retryCacheRename(func() (error, bool) {
		err := os.Rename(oldPath, newPath)
		return err, isRetryableCacheRenameError(err)
	})
}

func retryCacheRename(rename func() (err error, retryable bool)) error {
	// Caching is best-effort: allow brief Windows sharing conflicts to clear
	// without holding up an API response indefinitely.
	deadline := time.Now().Add(100 * time.Millisecond)
	for delay := time.Millisecond; ; delay *= 2 {
		err, retryable := rename()
		if err == nil || !retryable {
			return err
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return err
		}
		time.Sleep(min(delay, remaining))
	}
}
