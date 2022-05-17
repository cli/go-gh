package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigDir(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		onlyWindows bool
		env         map[string]string
		output      string
	}{
		{
			name: "HOME/USERPROFILE specified",
			env: map[string]string{
				"GH_CONFIG_DIR":   "",
				"XDG_CONFIG_HOME": "",
				"AppData":         "",
				"USERPROFILE":     tempDir,
				"HOME":            tempDir,
			},
			output: filepath.Join(tempDir, ".config", "gh"),
		},
		{
			name: "GH_CONFIG_DIR specified",
			env: map[string]string{
				"GH_CONFIG_DIR": filepath.Join(tempDir, "gh_config_dir"),
			},
			output: filepath.Join(tempDir, "gh_config_dir"),
		},
		{
			name: "XDG_CONFIG_HOME specified",
			env: map[string]string{
				"XDG_CONFIG_HOME": tempDir,
			},
			output: filepath.Join(tempDir, "gh"),
		},
		{
			name: "GH_CONFIG_DIR and XDG_CONFIG_HOME specified",
			env: map[string]string{
				"GH_CONFIG_DIR":   filepath.Join(tempDir, "gh_config_dir"),
				"XDG_CONFIG_HOME": tempDir,
			},
			output: filepath.Join(tempDir, "gh_config_dir"),
		},
		{
			name:        "AppData specified",
			onlyWindows: true,
			env: map[string]string{
				"AppData": tempDir,
			},
			output: filepath.Join(tempDir, "GitHub CLI"),
		},
		{
			name:        "GH_CONFIG_DIR and AppData specified",
			onlyWindows: true,
			env: map[string]string{
				"GH_CONFIG_DIR": filepath.Join(tempDir, "gh_config_dir"),
				"AppData":       tempDir,
			},
			output: filepath.Join(tempDir, "gh_config_dir"),
		},
		{
			name:        "XDG_CONFIG_HOME and AppData specified",
			onlyWindows: true,
			env: map[string]string{
				"XDG_CONFIG_HOME": tempDir,
				"AppData":         tempDir,
			},
			output: filepath.Join(tempDir, "gh"),
		},
	}

	for _, tt := range tests {
		if tt.onlyWindows && runtime.GOOS != "windows" {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != nil {
				for k, v := range tt.env {
					old := os.Getenv(k)
					os.Setenv(k, v)
					defer os.Setenv(k, old)
				}
			}

			assert.Equal(t, tt.output, configDir())
		})
	}
}

func TestStateDir(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		onlyWindows bool
		env         map[string]string
		output      string
	}{
		{
			name: "HOME/USERPROFILE specified",
			env: map[string]string{
				"XDG_STATE_HOME":  "",
				"GH_CONFIG_DIR":   "",
				"XDG_CONFIG_HOME": "",
				"LocalAppData":    "",
				"USERPROFILE":     tempDir,
				"HOME":            tempDir,
			},
			output: filepath.Join(tempDir, ".local", "state", "gh"),
		},
		{
			name: "XDG_STATE_HOME specified",
			env: map[string]string{
				"XDG_STATE_HOME": tempDir,
			},
			output: filepath.Join(tempDir, "gh"),
		},
		{
			name:        "LocalAppData specified",
			onlyWindows: true,
			env: map[string]string{
				"LocalAppData": tempDir,
			},
			output: filepath.Join(tempDir, "GitHub CLI"),
		},
		{
			name:        "XDG_STATE_HOME and LocalAppData specified",
			onlyWindows: true,
			env: map[string]string{
				"XDG_STATE_HOME": tempDir,
				"LocalAppData":   tempDir,
			},
			output: filepath.Join(tempDir, "gh"),
		},
	}

	for _, tt := range tests {
		if tt.onlyWindows && runtime.GOOS != "windows" {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != nil {
				for k, v := range tt.env {
					old := os.Getenv(k)
					os.Setenv(k, v)
					defer os.Setenv(k, old)
				}
			}

			assert.Equal(t, tt.output, stateDir())
		})
	}
}

func TestDataDir(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		onlyWindows bool
		env         map[string]string
		output      string
	}{
		{
			name: "HOME/USERPROFILE specified",
			env: map[string]string{
				"XDG_DATA_HOME":   "",
				"GH_CONFIG_DIR":   "",
				"XDG_CONFIG_HOME": "",
				"LocalAppData":    "",
				"USERPROFILE":     tempDir,
				"HOME":            tempDir,
			},
			output: filepath.Join(tempDir, ".local", "share", "gh"),
		},
		{
			name: "XDG_DATA_HOME specified",
			env: map[string]string{
				"XDG_DATA_HOME": tempDir,
			},
			output: filepath.Join(tempDir, "gh"),
		},
		{
			name:        "LocalAppData specified",
			onlyWindows: true,
			env: map[string]string{
				"LocalAppData": tempDir,
			},
			output: filepath.Join(tempDir, "GitHub CLI"),
		},
		{
			name:        "XDG_DATA_HOME and LocalAppData specified",
			onlyWindows: true,
			env: map[string]string{
				"XDG_DATA_HOME": tempDir,
				"LocalAppData":  tempDir,
			},
			output: filepath.Join(tempDir, "gh"),
		},
	}

	for _, tt := range tests {
		if tt.onlyWindows && runtime.GOOS != "windows" {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != nil {
				for k, v := range tt.env {
					old := os.Getenv(k)
					os.Setenv(k, v)
					defer os.Setenv(k, old)
				}
			}

			assert.Equal(t, tt.output, dataDir())
		})
	}
}

func TestLoad(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	assert.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	globalFilePath := filepath.Join(tempDir, "config.yml")
	invalidGlobalFilePath := filepath.Join(tempDir, "invalid_config.yml")
	hostsFilePath := filepath.Join(tempDir, "hosts.yml")
	invalidHostsFilePath := filepath.Join(tempDir, "invalid_hosts.yml")
	err := os.WriteFile(globalFilePath, []byte(testGlobalConfig()), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(invalidGlobalFilePath, []byte("invalid"), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(hostsFilePath, []byte(testHostsConfig()), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(invalidHostsFilePath, []byte("invalid"), 0755)
	assert.NoError(t, err)

	tests := []struct {
		name             string
		globalConfigPath string
		hostsConfigPath  string
		wantGitProtocol  string
		wantToken        string
		wantErr          bool
		wantErrMsg       string
		wantGetErr       bool
	}{
		{
			name:             "global and hosts files exist",
			globalConfigPath: globalFilePath,
			hostsConfigPath:  hostsFilePath,
			wantGitProtocol:  "ssh",
			wantToken:        "yyyyyyyyyyyyyyyyyyyy",
		},
		{
			name:             "invalid global file",
			globalConfigPath: invalidGlobalFilePath,
			wantErr:          true,
			wantErrMsg:       fmt.Sprintf("invalid config file %s: invalid format", filepath.Join(tempDir, "invalid_config.yml")),
		},
		{
			name:             "invalid hosts file",
			globalConfigPath: globalFilePath,
			hostsConfigPath:  invalidHostsFilePath,
			wantErr:          true,
			wantErrMsg:       fmt.Sprintf("invalid config file %s: invalid format", filepath.Join(tempDir, "invalid_hosts.yml")),
		},
		{
			name:             "global file does not exist and hosts file exist",
			globalConfigPath: "",
			hostsConfigPath:  hostsFilePath,
			wantGitProtocol:  "https",
			wantToken:        "yyyyyyyyyyyyyyyyyyyy",
		},
		{
			name:             "global file exist and hosts file does not exist",
			globalConfigPath: globalFilePath,
			hostsConfigPath:  "",
			wantGitProtocol:  "ssh",
			wantGetErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := load(tt.globalConfigPath, tt.hostsConfigPath)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
			assert.NoError(t, err)

			git_protocol := cfg.General().Get("git_protocol").Value()
			assert.Equal(t, tt.wantGitProtocol, git_protocol)

			tokenValue := cfg.Hosts().Get("enterprise.com", "oauth_token")
			if tt.wantGetErr {
				assert.True(t, tokenValue.NotFound())
				return
			}
			assert.Equal(t, tt.wantToken, tokenValue.Value())
		})
	}
}

func TestWrite(t *testing.T) {
	//TODO: Write tests.
}

func testGlobalConfig() string {
	var data = `
git_protocol: ssh
editor:
prompt: enabled
pager: less
`
	return data
}

func testHostsConfig() string {
	var data = `
github.com:
  user: user1
  oauth_token: xxxxxxxxxxxxxxxxxxxx
  git_protocol: ssh
enterprise.com:
  user: user2
  oauth_token: yyyyyyyyyyyyyyyyyyyy
  git_protocol: https
`
	return data
}
