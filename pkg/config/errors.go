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

// SetInvalidKeyError represents an error when trying to set a invalid key.
type SetInvalidKeyError struct {
	Host  string
	Key   string
	Value string
}

// Allow SetInvalidKeyError to satisfy error interface.
func (e SetInvalidKeyError) Error() string {
	if e.Host != "" {
		return fmt.Sprintf("can not set %q for %q because it is an invalid key", e.Key, e.Host)
	}
	return fmt.Sprintf("can not set %q because it is an invalid key", e.Key)
}

// SetInvalidValueError represents an error when trying to set a invalid value.
type SetInvalidValueError struct {
	Host  string
	Key   string
	Value string
}

// Allow SetInvalidValueError to satisfy error interface.
func (e SetInvalidValueError) Error() string {
	if e.Host != "" {
		return fmt.Sprintf("can not set %q to %q for %q because it is an invalid value", e.Key, e.Value, e.Host)
	}
	return fmt.Sprintf("can not set %q to %q because it is an invalid value", e.Key, e.Value)
}
