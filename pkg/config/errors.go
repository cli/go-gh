package config

import (
	"fmt"
)

// InvalidConfigFileError represents an error when trying to read a config file.
type InvalidConfigFileError struct {
	Path string
	Err  error
}

// Allow InvalidConfigFileError to satisfy error interface.
func (e InvalidConfigFileError) Error() string {
	return fmt.Sprintf("invalid config file %s: %s", e.Path, e.Err)
}

// NotFoundError represents an error when trying to find a config key
// that does not exist.
type NotFoundError struct {
	Key string
}

// Allow NotFoundError to satisfy error interface.
func (e NotFoundError) Error() string {
	return fmt.Sprintf("could not find key %q", e.Key)
}
