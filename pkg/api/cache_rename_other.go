//go:build !windows

package api

func isRetryableCacheRenameError(error) bool {
	return false
}
