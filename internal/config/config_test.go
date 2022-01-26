package config

import (
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

func TestConfigGet(t *testing.T) {
	cfg := testLoadedConfig()

	tests := []struct {
		name       string
		key        string
		wantValue  string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:      "get git_protocol value",
			key:       "git_protocol",
			wantValue: "ssh",
		},
		{
			name:      "get editor value",
			key:       "editor",
			wantValue: "",
		},
		{
			name:      "get prompt value",
			key:       "prompt",
			wantValue: "enabled",
		},
		{
			name:      "get pager value",
			key:       "pager",
			wantValue: "less",
		},
		{
			name:       "unknown key",
			key:        "unknown",
			wantErr:    true,
			wantErrMsg: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := cfg.Get(tt.key)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}

func TestConfigGetForHost(t *testing.T) {
	cfg := testLoadedConfig()

	tests := []struct {
		name       string
		host       string
		key        string
		wantValue  string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:      "get github user value",
			host:      "github.com",
			key:       "user",
			wantValue: "user1",
		},
		{
			name:      "get github oauth_token value",
			host:      "github.com",
			key:       "oauth_token",
			wantValue: "xxxxxxxxxxxxxxxxxxxx",
		},
		{
			name:      "get github git_protocol value",
			host:      "github.com",
			key:       "git_protocol",
			wantValue: "ssh",
		},
		{
			name:      "get enterprise user value",
			host:      "enterprise.com",
			key:       "user",
			wantValue: "user2",
		},
		{
			name:      "get enterprise oauth_token value",
			host:      "enterprise.com",
			key:       "oauth_token",
			wantValue: "yyyyyyyyyyyyyyyyyyyy",
		},
		{
			name:      "get enterprise git_protocol value",
			host:      "enterprise.com",
			key:       "git_protocol",
			wantValue: "https",
		},
		{
			name:       "unknown host",
			host:       "unknown",
			key:        "user",
			wantErr:    true,
			wantErrMsg: "not found",
		},
		{
			name:       "unknown key",
			host:       "github.com",
			key:        "unknown",
			wantErr:    true,
			wantErrMsg: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := cfg.GetForHost(tt.host, tt.key)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}

func TestConfigHost(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		ghHost   string
		wantHost string
	}{
		{
			name:     "GH_HOST if set",
			cfg:      testLoadedNoHostConfig(),
			ghHost:   "test.com",
			wantHost: "test.com",
		},
		{
			name:     "authenticated host if only one",
			cfg:      testLoadedSingleHostConfig(),
			wantHost: "enterprise.com",
		},
		{
			name:     "default host if more than one authenticated host",
			cfg:      testLoadedConfig(),
			wantHost: "github.com",
		},
		{
			name:     "default host if no authenticated host",
			cfg:      testLoadedNoHostConfig(),
			wantHost: "github.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ghHost != "" {
				k := "GH_HOST"
				old := os.Getenv(k)
				os.Setenv(k, tt.ghHost)
				defer os.Setenv(k, old)
			}
			host := tt.cfg.Host()
			assert.Equal(t, tt.wantHost, host)
		})
	}
}

func TestConfigAuthToken(t *testing.T) {
	orig_GITHUB_TOKEN := os.Getenv("GITHUB_TOKEN")
	orig_GITHUB_ENTERPRISE_TOKEN := os.Getenv("GITHUB_ENTERPRISE_TOKEN")
	orig_GH_TOKEN := os.Getenv("GH_TOKEN")
	orig_GH_ENTERPRISE_TOKEN := os.Getenv("GH_ENTERPRISE_TOKEN")
	t.Cleanup(func() {
		os.Setenv("GITHUB_TOKEN", orig_GITHUB_TOKEN)
		os.Setenv("GITHUB_ENTERPRISE_TOKEN", orig_GITHUB_ENTERPRISE_TOKEN)
		os.Setenv("GH_TOKEN", orig_GH_TOKEN)
		os.Setenv("GH_ENTERPRISE_TOKEN", orig_GH_ENTERPRISE_TOKEN)
	})

	tests := []struct {
		name                    string
		host                    string
		GITHUB_TOKEN            string
		GITHUB_ENTERPRISE_TOKEN string
		GH_TOKEN                string
		GH_ENTERPRISE_TOKEN     string
		cfg                     Config
		wantToken               string
		wantErr                 bool
		wantErrMsg              string
	}{
		{
			name:       "token for github.com with no env tokens and no config token",
			host:       "github.com",
			cfg:        testLoadedNoHostConfig(),
			wantErr:    true,
			wantErrMsg: "not found",
		},
		{
			name:       "token for enterprise.com with no env tokens and no config token",
			host:       "enterprise.com",
			cfg:        testLoadedNoHostConfig(),
			wantErr:    true,
			wantErrMsg: "not found",
		},
		{
			name:         "token for github.com with GH_TOKEN, GITHUB_TOKEN, and config token",
			host:         "github.com",
			GH_TOKEN:     "GH_TOKEN",
			GITHUB_TOKEN: "GITHUB_TOKEN",
			cfg:          testLoadedConfig(),
			wantToken:    "GH_TOKEN",
		},
		{
			name:         "token for github.com with GITHUB_TOKEN, and config token",
			host:         "github.com",
			GITHUB_TOKEN: "GITHUB_TOKEN",
			cfg:          testLoadedConfig(),
			wantToken:    "GITHUB_TOKEN",
		},
		{
			name:      "token for github.com with config token",
			host:      "github.com",
			cfg:       testLoadedConfig(),
			wantToken: "xxxxxxxxxxxxxxxxxxxx",
		},
		{
			name:                    "token for enterprise.com with GH_ENTERPRISE_TOKEN, GITHUB_ENTERPRISE_TOKEN, and config token",
			host:                    "enterprise.com",
			GH_ENTERPRISE_TOKEN:     "GH_ENTERPRISE_TOKEN",
			GITHUB_ENTERPRISE_TOKEN: "GITHUB_ENTERPRISE_TOKEN",
			cfg:                     testLoadedConfig(),
			wantToken:               "GH_ENTERPRISE_TOKEN",
		},
		{
			name:                    "token for enterprise.com with GITHUB_ENTERPRISE_TOKEN, and config token",
			host:                    "enterprise.com",
			GITHUB_ENTERPRISE_TOKEN: "GITHUB_ENTERPRISE_TOKEN",
			cfg:                     testLoadedConfig(),
			wantToken:               "GITHUB_ENTERPRISE_TOKEN",
		},
		{
			name:      "token for enterprise.com with config token",
			host:      "enterprise.com",
			cfg:       testLoadedConfig(),
			wantToken: "yyyyyyyyyyyyyyyyyyyy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GITHUB_TOKEN", tt.GITHUB_TOKEN)
			os.Setenv("GITHUB_ENTERPRISE_TOKEN", tt.GITHUB_ENTERPRISE_TOKEN)
			os.Setenv("GH_TOKEN", tt.GH_TOKEN)
			os.Setenv("GH_ENTERPRISE_TOKEN", tt.GH_ENTERPRISE_TOKEN)
			token, err := tt.cfg.AuthToken(tt.host)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantToken, token)
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
		wantGetErrMsg    string
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
			wantErrMsg:       "invalid config file",
		},
		{
			name:             "invalid hosts file",
			globalConfigPath: globalFilePath,
			hostsConfigPath:  invalidHostsFilePath,
			wantErr:          true,
			wantErrMsg:       "invalid config file",
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
			wantGetErrMsg:    "not found",
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

			git_protocol, err := cfg.Get("git_protocol")
			assert.NoError(t, err)
			assert.Equal(t, tt.wantGitProtocol, git_protocol)

			token, err := cfg.GetForHost("enterprise.com", "oauth_token")
			if tt.wantGetErr {
				assert.EqualError(t, err, tt.wantGetErrMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantToken, token)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	git_protocol, err := cfg.Get("git_protocol")
	assert.NoError(t, err)
	assert.Equal(t, "https", git_protocol)

	editor, err := cfg.Get("editor")
	assert.NoError(t, err)
	assert.Equal(t, "", editor)

	prompt, err := cfg.Get("prompt")
	assert.NoError(t, err)
	assert.Equal(t, "enabled", prompt)

	pager, err := cfg.Get("pager")
	assert.NoError(t, err)
	assert.Equal(t, "", pager)

	unix_socket, err := cfg.Get("http_unix_socket")
	assert.NoError(t, err)
	assert.Equal(t, "", unix_socket)

	browser, err := cfg.Get("browser")
	assert.NoError(t, err)
	assert.Equal(t, "", browser)

	_, err = cfg.Get("unknown")
	assert.EqualError(t, err, "not found")
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

func testLoadedConfig() Config {
	var data = `
git_protocol: ssh
editor:
prompt: enabled
pager: less
hosts:
  github.com:
    user: user1
    oauth_token: xxxxxxxxxxxxxxxxxxxx
    git_protocol: ssh
  enterprise.com:
    user: user2
    oauth_token: yyyyyyyyyyyyyyyyyyyy
    git_protocol: https
`
	cfg, _ := FromString(data)
	return cfg
}

func testLoadedSingleHostConfig() Config {
	var data = `
git_protocol: ssh
editor:
prompt: enabled
pager: less
hosts:
  enterprise.com:
    user: user2
    oauth_token: yyyyyyyyyyyyyyyyyyyy
    git_protocol: https
`
	cfg, _ := FromString(data)
	return cfg
}

func testLoadedNoHostConfig() Config {
	cfg, _ := FromString(testGlobalConfig())
	return cfg
}
