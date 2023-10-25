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
					t.Setenv(k, v)
				}
			}
			assert.Equal(t, tt.output, ConfigDir())
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
					t.Setenv(k, v)
				}
			}
			assert.Equal(t, tt.output, StateDir())
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
					t.Setenv(k, v)
				}
			}
			assert.Equal(t, tt.output, DataDir())
		})
	}
}

func TestCacheDir(t *testing.T) {
	expected := filepath.Join(os.TempDir(), "gh-cli-cache")
	actual := CacheDir()
	assert.Equal(t, expected, actual)
}

func TestLoad(t *testing.T) {
	tempDir := t.TempDir()
	globalFilePath := filepath.Join(tempDir, "config.yml")
	invalidGlobalFilePath := filepath.Join(tempDir, "invalid_config.yml")
	hostsFilePath := filepath.Join(tempDir, "hosts.yml")
	invalidHostsFilePath := filepath.Join(tempDir, "invalid_hosts.yml")
	err := os.WriteFile(globalFilePath, []byte(testGlobalData()), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(invalidGlobalFilePath, []byte("invalid"), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(hostsFilePath, []byte(testHostsData()), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(invalidHostsFilePath, []byte("invalid"), 0755)
	assert.NoError(t, err)

	tests := []struct {
		name             string
		globalConfigPath string
		hostsConfigPath  string
		fallback         *Config
		wantGitProtocol  string
		wantToken        string
		wantErr          bool
		wantErrMsg       string
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
			wantGitProtocol:  "",
			wantToken:        "yyyyyyyyyyyyyyyyyyyy",
		},
		{
			name:             "global file exist and hosts file does not exist",
			globalConfigPath: globalFilePath,
			hostsConfigPath:  "",
			wantGitProtocol:  "ssh",
			wantToken:        "",
		},
		{
			name:             "global file does not exist and hosts file does not exist with no fallback",
			globalConfigPath: "",
			hostsConfigPath:  "",
			wantGitProtocol:  "",
			wantToken:        "",
		},
		{
			name:             "global file does not exist and hosts file does not exist with fallback",
			globalConfigPath: "",
			hostsConfigPath:  "",
			fallback:         ReadFromString(testFullConfig()),
			wantGitProtocol:  "ssh",
			wantToken:        "yyyyyyyyyyyyyyyyyyyy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := load(tt.globalConfigPath, tt.hostsConfigPath, tt.fallback)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
			assert.NoError(t, err)

			if tt.wantGitProtocol == "" {
				assertNoKey(t, cfg, []string{"git_protocol"})
			} else {
				assertKeyWithValue(t, cfg, []string{"git_protocol"}, tt.wantGitProtocol)
			}

			if tt.wantToken == "" {
				assertNoKey(t, cfg, []string{"hosts", "enterprise.com", "oauth_token"})
			} else {
				assertKeyWithValue(t, cfg, []string{"hosts", "enterprise.com", "oauth_token"}, tt.wantToken)
			}

			if tt.fallback != nil {
				// Assert that load returns an equivalent copy of fallvback.
				assert.Equal(t, tt.fallback.entries.String(), cfg.entries.String())
				assert.False(t, tt.fallback == cfg)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name         string
		createConfig func() *Config
		wantConfig   func() *Config
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name: "writes config and hosts files",
			createConfig: func() *Config {
				cfg := ReadFromString(testFullConfig())
				cfg.Set([]string{"editor"}, "vim")
				cfg.Set([]string{"hosts", "github.com", "git_protocol"}, "https")
				return cfg
			},
			wantConfig: func() *Config {
				// Same as created config as both a global property and host property has
				// been edited.
				cfg := ReadFromString(testFullConfig())
				cfg.Set([]string{"editor"}, "vim")
				cfg.Set([]string{"hosts", "github.com", "git_protocol"}, "https")
				return cfg
			},
		},
		{
			name: "only writes hosts file",
			createConfig: func() *Config {
				cfg := ReadFromString(testFullConfig())
				cfg.Set([]string{"hosts", "enterprise.com", "git_protocol"}, "ssh")
				return cfg
			},
			wantConfig: func() *Config {
				// The hosts file is writen but not the global config file.
				cfg := ReadFromString("")
				cfg.Set([]string{"hosts", "github.com", "user"}, "user1")
				cfg.Set([]string{"hosts", "github.com", "oauth_token"}, "xxxxxxxxxxxxxxxxxxxx")
				cfg.Set([]string{"hosts", "github.com", "git_protocol"}, "ssh")
				cfg.Set([]string{"hosts", "enterprise.com", "user"}, "user2")
				cfg.Set([]string{"hosts", "enterprise.com", "oauth_token"}, "yyyyyyyyyyyyyyyyyyyy")
				cfg.Set([]string{"hosts", "enterprise.com", "git_protocol"}, "ssh")
				return cfg
			},
		},
		{
			name: "only writes global config file",
			createConfig: func() *Config {
				cfg := ReadFromString(testFullConfig())
				cfg.Set([]string{"editor"}, "vim")
				return cfg
			},
			wantConfig: func() *Config {
				// The global config file is written but not the hosts config file.
				cfg := ReadFromString(testGlobalData())
				cfg.Set([]string{"editor"}, "vim")
				return cfg
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			t.Setenv("GH_CONFIG_DIR", tempDir)
			cfg := tt.createConfig()
			err := Write(cfg)
			assert.NoError(t, err)
			loadedCfg, err := load(generalConfigFile(), hostsConfigFile(), nil)
			assert.NoError(t, err)
			wantCfg := tt.wantConfig()
			assert.Equal(t, wantCfg.entries.String(), loadedCfg.entries.String())
		})
	}
}

func TestWriteEmptyValues(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("GH_CONFIG_DIR", tempDir)
	cfg := ReadFromString(testFullConfig())
	cfg.Set([]string{"editor"}, "")
	err := Write(cfg)
	assert.NoError(t, err)
	data, err := os.ReadFile(generalConfigFile())
	assert.NoError(t, err)
	assert.Equal(t, "git_protocol: ssh\neditor:\nprompt: enabled\npager: less\n", string(data))
}

func TestGet(t *testing.T) {
	tests := []struct {
		name      string
		keys      []string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "get git_protocol value",
			keys:      []string{"git_protocol"},
			wantValue: "ssh",
		},
		{
			name:      "get editor value",
			keys:      []string{"editor"},
			wantValue: "",
		},
		{
			name:      "get prompt value",
			keys:      []string{"prompt"},
			wantValue: "enabled",
		},
		{
			name:      "get pager value",
			keys:      []string{"pager"},
			wantValue: "less",
		},
		{
			name:    "non-existant key",
			keys:    []string{"unknown"},
			wantErr: true,
		},
		{
			name:      "nested key",
			keys:      []string{"nested", "key"},
			wantValue: "value",
		},
		{
			name:      "nested key with same name",
			keys:      []string{"nested", "pager"},
			wantValue: "more",
		},
		{
			name:    "nested non-existant key",
			keys:    []string{"nested", "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testConfig()
			if tt.wantErr {
				assertNoKey(t, cfg, tt.keys)
			} else {
				assertKeyWithValue(t, cfg, tt.keys, tt.wantValue)
			}
			assert.False(t, cfg.entries.IsModified())
		})
	}
}

func TestKeys(t *testing.T) {
	tests := []struct {
		name       string
		findKeys   []string
		wantKeys   []string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "top level keys",
			findKeys: nil,
			wantKeys: []string{"git_protocol", "editor", "prompt", "pager", "nested"},
		},
		{
			name:     "nested keys",
			findKeys: []string{"nested"},
			wantKeys: []string{"key", "pager"},
		},
		{
			name:       "keys for non-existant nested key",
			findKeys:   []string{"unknown"},
			wantKeys:   nil,
			wantErr:    true,
			wantErrMsg: `could not find key "unknown"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testConfig()
			ks, err := cfg.Keys(tt.findKeys)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantKeys, ks)
			assert.False(t, cfg.entries.IsModified())
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name       string
		keys       []string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "remove top level key",
			keys: []string{"pager"},
		},
		{
			name: "remove nested key",
			keys: []string{"nested", "pager"},
		},
		{
			name: "remove top level map",
			keys: []string{"nested"},
		},
		{
			name:       "remove non-existant top level key",
			keys:       []string{"unknown"},
			wantErr:    true,
			wantErrMsg: `could not find key "unknown"`,
		},
		{
			name:       "remove non-existant nested key",
			keys:       []string{"nested", "invalid"},
			wantErr:    true,
			wantErrMsg: `could not find key "invalid"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testConfig()
			err := cfg.Remove(tt.keys)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
				assert.False(t, cfg.entries.IsModified())
			} else {
				assert.NoError(t, err)
				assert.True(t, cfg.entries.IsModified())
			}
			assertNoKey(t, cfg, tt.keys)
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name  string
		keys  []string
		value string
	}{
		{
			name:  "set top level existing key",
			keys:  []string{"pager"},
			value: "test pager",
		},
		{
			name:  "set nested existing key",
			keys:  []string{"nested", "pager"},
			value: "new pager",
		},
		{
			name:  "set top level map",
			keys:  []string{"nested"},
			value: "override",
		},
		{
			name:  "set non-existant top level key",
			keys:  []string{"unknown"},
			value: "why not",
		},
		{
			name:  "set non-existant nested key",
			keys:  []string{"nested", "invalid"},
			value: "sure",
		},
		{
			name:  "set non-existant nest",
			keys:  []string{"johnny", "test"},
			value: "dukey",
		},
		{
			name:  "set empty value",
			keys:  []string{"empty"},
			value: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testConfig()
			cfg.Set(tt.keys, tt.value)
			assert.True(t, cfg.entries.IsModified())
			assertKeyWithValue(t, cfg, tt.keys, tt.value)
		})
	}
}

func testConfig() *Config {
	var data = `
git_protocol: ssh
editor:
prompt: enabled
pager: less
nested:
  key: value
  pager: more
`
	return ReadFromString(data)
}

func testGlobalData() string {
	var data = `
git_protocol: ssh
editor:
prompt: enabled
pager: less
`
	return data
}

func testHostsData() string {
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

func testFullConfig() string {
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
	return data
}

func assertNoKey(t *testing.T, cfg *Config, keys []string) {
	t.Helper()
	_, err := cfg.Get(keys)
	var keyNotFoundError *KeyNotFoundError
	assert.ErrorAs(t, err, &keyNotFoundError)
}

func assertKeyWithValue(t *testing.T, cfg *Config, keys []string, value string) {
	t.Helper()
	actual, err := cfg.Get(keys)
	assert.NoError(t, err)
	assert.Equal(t, value, actual)
}
