package testutils

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/config"
)

// StubConfig replaces the config.Read function with a function that returns a config object
// created from the given config string. It also sets up a cleanup function that restores the
// original config.Read function.
func StubConfig(t *testing.T, cfgStr string) {
	t.Helper()
	old := config.Read
	config.Read = func(_ *config.Config) (*config.Config, error) {
		return config.ReadFromString(cfgStr), nil
	}
	t.Cleanup(func() {
		config.Read = old
	})
}
