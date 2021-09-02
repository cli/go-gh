package config

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

const (
	GH_CONFIG_DIR   = "GH_CONFIG_DIR"
	XDG_CONFIG_HOME = "XDG_CONFIG_HOME"
	XDG_STATE_HOME  = "XDG_STATE_HOME"
	XDG_DATA_HOME   = "XDG_DATA_HOME"
	APP_DATA        = "AppData"
	LOCAL_APP_DATA  = "LocalAppData"
)

type Config interface {
	Get(key string) (string, error)
	GetForHost(host string, key string) (string, error)
}

type config struct {
	global configMap
	hosts  configMap
}

func (c config) Get(key string) (string, error) {
	return c.global.getStringValue(key)
}

func (c config) GetForHost(host, key string) (string, error) {
	hostEntry, err := c.hosts.findEntry(host)
	if err != nil {
		return "", err
	}
	hostMap := configMap{Root: hostEntry.ValueNode}
	return hostMap.getStringValue(key)
}

func FromString(str string) (Config, error) {
	root, err := parseData([]byte(str))
	if err != nil {
		return nil, err
	}
	cfg := config{}
	globalMap := configMap{Root: root}
	cfg.global = globalMap
	hostsEntry, err := globalMap.findEntry("hosts")
	if err == nil {
		cfg.hosts = configMap{Root: hostsEntry.ValueNode}
	}
	return cfg, nil
}

func DefaultConfig() Config {
	return config{global: configMap{Root: defaultGlobal().Content[0]}}
}

func Load() (Config, error) {
	return load(ConfigFile(), HostsConfigFile())
}

func load(globalFilePath, hostsFilePath string) (Config, error) {
	var readErr error
	var parseErr error
	globalData, readErr := readFile(globalFilePath)
	if readErr != nil && !errors.Is(readErr, fs.ErrNotExist) {
		return nil, readErr
	}

	//Use defaultGlobal node if globalFile does not exist or is empty
	global := defaultGlobal().Content[0]
	if len(globalData) > 0 {
		global, parseErr = parseData(globalData)
	}
	if parseErr != nil {
		return nil, parseErr
	}

	hostsData, readErr := readFile(hostsFilePath)
	if readErr != nil && !os.IsNotExist(readErr) {
		return nil, readErr
	}

	//Use nil if hostsFile does not exist or is empty
	var hosts *yaml.Node
	if len(hostsData) > 0 {
		hosts, parseErr = parseData(hostsData)
	}
	if parseErr != nil {
		return nil, parseErr
	}

	cfg := config{
		global: configMap{Root: global},
		hosts:  configMap{Root: hosts},
	}

	return cfg, nil
}

// Config path precedence
// 1. GH_CONFIG_DIR
// 2. XDG_CONFIG_HOME
// 3. AppData (windows only)
// 4. HOME
func ConfigDir() string {
	var path string
	if a := os.Getenv(GH_CONFIG_DIR); a != "" {
		path = a
	} else if b := os.Getenv(XDG_CONFIG_HOME); b != "" {
		path = filepath.Join(b, "gh")
	} else if c := os.Getenv(APP_DATA); runtime.GOOS == "windows" && c != "" {
		path = filepath.Join(c, "GitHub CLI")
	} else {
		d, _ := os.UserHomeDir()
		path = filepath.Join(d, ".config", "gh")
	}
	return path
}

// State path precedence
// 1. XDG_CONFIG_HOME
// 2. LocalAppData (windows only)
// 3. HOME
func StateDir() string {
	var path string
	if a := os.Getenv(XDG_STATE_HOME); a != "" {
		path = filepath.Join(a, "gh")
	} else if b := os.Getenv(LOCAL_APP_DATA); runtime.GOOS == "windows" && b != "" {
		path = filepath.Join(b, "GitHub CLI")
	} else {
		c, _ := os.UserHomeDir()
		path = filepath.Join(c, ".local", "state", "gh")
	}
	return path
}

// Data path precedence
// 1. XDG_DATA_HOME
// 2. LocalAppData (windows only)
// 3. HOME
func DataDir() string {
	var path string
	if a := os.Getenv(XDG_DATA_HOME); a != "" {
		path = filepath.Join(a, "gh")
	} else if b := os.Getenv(LOCAL_APP_DATA); runtime.GOOS == "windows" && b != "" {
		path = filepath.Join(b, "GitHub CLI")
	} else {
		c, _ := os.UserHomeDir()
		path = filepath.Join(c, ".local", "share", "gh")
	}
	return path
}

func ConfigFile() string {
	return filepath.Join(ConfigDir(), "config.yml")
}

func HostsConfigFile() string {
	return filepath.Join(ConfigDir(), "hosts.yml")
}

func readFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func parseData(data []byte) (*yaml.Node, error) {
	var root yaml.Node
	err := yaml.Unmarshal(data, &root)
	if err != nil {
		return nil, fmt.Errorf("invalid config file: %w", err)
	}
	if len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("invalid config file")
	}
	return root.Content[0], nil
}

func defaultGlobal() *yaml.Node {
	return &yaml.Node{
		Kind: yaml.DocumentNode,
		Content: []*yaml.Node{
			{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					{
						HeadComment: "What protocol to use when performing git operations. Supported values: ssh, https",
						Kind:        yaml.ScalarNode,
						Value:       "git_protocol",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "https",
					},
					{
						HeadComment: "What editor gh should run when creating issues, pull requests, etc. If blank, will refer to environment.",
						Kind:        yaml.ScalarNode,
						Value:       "editor",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "",
					},
					{
						HeadComment: "When to interactively prompt. This is a global config that cannot be overridden by hostname. Supported values: enabled, disabled",
						Kind:        yaml.ScalarNode,
						Value:       "prompt",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "enabled",
					},
					{
						HeadComment: "A pager program to send command output to, e.g. \"less\". Set the value to \"cat\" to disable the pager.",
						Kind:        yaml.ScalarNode,
						Value:       "pager",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "",
					},
					{
						HeadComment: "Aliases allow you to create nicknames for gh commands",
						Kind:        yaml.ScalarNode,
						Value:       "aliases",
					},
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{
								Kind:  yaml.ScalarNode,
								Value: "co",
							},
							{
								Kind:  yaml.ScalarNode,
								Value: "pr checkout",
							},
						},
					},
					{
						HeadComment: "The path to a unix socket through which send HTTP connections. If blank, HTTP traffic will be handled by net/http.DefaultTransport.",
						Kind:        yaml.ScalarNode,
						Value:       "http_unix_socket",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "",
					},
					{
						HeadComment: "What web browser gh should use when opening URLs. If blank, will refer to environment.",
						Kind:        yaml.ScalarNode,
						Value:       "browser",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "",
					},
				},
			},
		},
	}
}
