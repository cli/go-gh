package color

import (
	"os"

	"github.com/cli/go-gh/v2/pkg/config"
)

const (
	// AccessibleColorsEnv is the name of the environment variable to enable accessibile color features.
	AccessibleColorsEnv = "GH_ACCESSIBLE_COLORS"

	// AccessibleColorsSetting is the name of the `gh config` setting to enable accessibile color features.
	AccessibleColorsSetting = "accessible_colors"
)

// IsAccessibleColorsEnabled returns true if accessible colors are enabled via environment variable
// or configuration setting with the environment variable having higher precedence.
//
// If the environment variable is set, then any value other than the following will enable accessible colors:
// empty, "0", "false", or "no".
func IsAccessibleColorsEnabled() bool {
	// Environment variable only has the highest precedence if actually set.
	if envVar, set := os.LookupEnv(AccessibleColorsEnv); set {
		switch envVar {
		case "", "0", "false", "no":
			return false
		default:
			return true
		}
	}

	// We are not handling errors because we don't want to fail if the config is not
	// read. Instead, we assume an empty configuration is equivalent to "disabled".
	cfg, _ := config.Read(nil)
	accessibleConfigValue, _ := cfg.Get([]string{AccessibleColorsSetting})

	return accessibleConfigValue == "enabled"
}
