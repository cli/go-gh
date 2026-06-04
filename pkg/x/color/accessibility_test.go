package color

import (
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh/v2/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestIsAccessibleColorsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		cfgStr  string
		wantOut bool
	}{
		{
			name:    "When the accessibility configuration and env var are both unset, it should return false",
			cfgStr:  "",
			wantOut: false,
		},
		{
			name: "When the accessibility configuration is unset but the env var is set to something truthy (not '0' or 'false'), it should return true",
			env: map[string]string{
				"GH_ACCESSIBLE_COLORS": "1",
			},
			cfgStr:  "",
			wantOut: true,
		},
		{
			name: "When the accessibility configuration is unset and the env var returns '0', it should return false",
			env: map[string]string{
				"GH_ACCESSIBLE_COLORS": "0",
			},
			cfgStr:  "",
			wantOut: false,
		},
		{
			name: "When the accessibility configuration is unset and the env var returns 'false', it should return false",
			env: map[string]string{
				"GH_ACCESSIBLE_COLORS": "false",
			},
			cfgStr:  "",
			wantOut: false,
		},
		{
			name: "When the accessibility configuration is unset and the env var returns '', it should return false",
			env: map[string]string{
				"GH_ACCESSIBLE_COLORS": "",
			},
			cfgStr:  "",
			wantOut: false,
		},
		{
			name:    "When the accessibility configuration is set to enabled and the env var is unset, it should return true",
			cfgStr:  accessibilityEnabledConfig(),
			wantOut: true,
		},
		{
			name:    "When the accessibility configuration is set to disabled and the env var is unset, it should return false",
			cfgStr:  accessibilityDisabledConfig(),
			wantOut: false,
		},
		{
			name: "When the accessibility configuration is set to disabled and the env var is set to something truthy (not '0' or 'false'), it should return true",
			env: map[string]string{
				"GH_ACCESSIBLE_COLORS": "true",
			},
			cfgStr:  accessibilityDisabledConfig(),
			wantOut: true,
		},
		{
			name: "When the accessibility configuration is set to enabled and the env var is set to '0', it should return false",
			env: map[string]string{
				"GH_ACCESSIBLE_COLORS": "0",
			},
			cfgStr:  accessibilityEnabledConfig(),
			wantOut: false,
		},
		{
			name: "When the accessibility configuration is set to enabled and the env var is set to 'false', it should return false",
			env: map[string]string{
				"GH_ACCESSIBLE_COLORS": "false",
			},
			cfgStr:  accessibilityEnabledConfig(),
			wantOut: false,
		},
		{
			name: "When the accessibility configuration is set to enabled and the env var is set to '', it should return false",
			env: map[string]string{
				"GH_ACCESSIBLE_COLORS": "",
			},
			cfgStr:  accessibilityEnabledConfig(),
			wantOut: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			testutils.StubConfig(t, tt.cfgStr)
			assert.Equal(t, tt.wantOut, IsAccessibleColorsEnabled())
		})
	}
}

func accessibilityEnabledConfig() string {
	return heredoc.Doc(`
		accessible_colors: enabled
	`)
}

func accessibilityDisabledConfig() string {
	return heredoc.Doc(`
		accessible_colors: disabled
	`)
}
