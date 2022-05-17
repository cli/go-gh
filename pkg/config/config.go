// Package api is a set of types for interacting with the gh configuration.
package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cli/go-gh/internal/yamlmap"
)

const (
	appData       = "AppData"
	ghConfigDir   = "GH_CONFIG_DIR"
	localAppData  = "LocalAppData"
	xdgConfigHome = "XDG_CONFIG_HOME"
	xdgDataHome   = "XDG_DATA_HOME"
	xdgStateHome  = "XDG_STATE_HOME"

	defaultSource = "default"
)

type Config interface {
	General() GeneralEntries
	Hosts() HostEntries
	Aliases() AliasEntries
}

type Value interface {
	NotFound() bool
	Source() string
	Value() string
}

type value struct {
	notFound bool
	source   string
	value    string
}

func (v *value) NotFound() bool {
	return v.notFound
}

func (v *value) Source() string {
	if v.notFound {
		return defaultSource
	}
	return v.source
}

func (v *value) Value() string {
	return v.value
}

type cfg struct {
	general GeneralEntries
	hosts   HostEntries
	aliases AliasEntries
}

func (c *cfg) General() GeneralEntries {
	return c.general
}

func (c *cfg) Hosts() HostEntries {
	return c.hosts
}

func (c *cfg) Aliases() AliasEntries {
	return c.aliases
}

func Read() (Config, error) {
	return load(generalConfigFile(), hostsConfigFile())
}

// ReadFromString takes a yaml string and returns a Config.
// It assumes the yaml has a "hosts" and "aliases" key that
// are used as the alias and host entries.
// Note: This is only used in testing, and should not be
// relied upon in production.
func ReadFromString(str string) Config {
	ge, _ := mapFromString(str)
	he, _ := ge.FindEntry("hosts")
	ae, _ := ge.FindEntry("aliases")

	return &cfg{
		aliases: &aliasEntries{entries: ae},
		hosts:   &hostEntries{entries: he},
		general: &generalEntries{entries: ge},
	}
}

func Write(config Config) error {
	if config.General().Dirty() || config.Aliases().Dirty() {
		// Both general and aliases live in same file so if either
		// has been modified then we write the whole file. Additionally,
		// aliases is just a entry in general so writing general will
		// implicitly include aliases.
		err := writeFile(generalConfigFile(), []byte(config.General().String()))
		if err != nil {
			return err
		}
	}

	if config.Hosts().Dirty() {
		err := writeFile(hostsConfigFile(), []byte(config.Hosts().String()))
		if err != nil {
			return err
		}
	}

	return nil
}

func load(generalFilePath, hostsFilePath string) (Config, error) {
	generalData, err := mapFromFile(generalFilePath)
	if err != nil && !os.IsNotExist(err) {
		if errors.Is(err, yamlmap.ErrInvalidYaml) ||
			errors.Is(err, yamlmap.ErrInvalidFormat) {
			return nil, InvalidConfigFileError{Path: generalFilePath, Err: err}
		}
		return nil, err
	}

	if generalData.Empty() {
		generalData, _ = mapFromString(defaultGeneralEntries)
	}

	aliasesData, err := generalData.FindEntry("aliases")
	if err != nil {
		aliasesData = yamlmap.MapValue()
		generalData.AddEntry("aliases", aliasesData)
	}

	hostsData, err := mapFromFile(hostsFilePath)
	if err != nil && !os.IsNotExist(err) {
		if errors.Is(err, yamlmap.ErrInvalidYaml) ||
			errors.Is(err, yamlmap.ErrInvalidFormat) {
			return nil, InvalidConfigFileError{Path: hostsFilePath, Err: err}
		}
		return nil, err
	}

	return &cfg{
		aliases: &aliasEntries{entries: aliasesData},
		hosts:   &hostEntries{entries: hostsData},
		general: &generalEntries{entries: generalData},
	}, nil
}

func generalConfigFile() string {
	return filepath.Join(configDir(), "config.yml")
}

func hostsConfigFile() string {
	return filepath.Join(configDir(), "hosts.yml")
}

func mapFromFile(filename string) (yamlmap.Map, error) {
	data, err := readFile(filename)
	if err != nil {
		return yamlmap.NewMap(), err
	}
	return yamlmap.Unmarshal(data)
}

func mapFromString(str string) (yamlmap.Map, error) {
	return yamlmap.Unmarshal([]byte(str))
}

// Config path precedence: GH_CONFIG_DIR, XDG_CONFIG_HOME, AppData (windows only), HOME.
func configDir() string {
	var path string
	if a := os.Getenv(ghConfigDir); a != "" {
		path = a
	} else if b := os.Getenv(xdgConfigHome); b != "" {
		path = filepath.Join(b, "gh")
	} else if c := os.Getenv(appData); runtime.GOOS == "windows" && c != "" {
		path = filepath.Join(c, "GitHub CLI")
	} else {
		d, _ := os.UserHomeDir()
		path = filepath.Join(d, ".config", "gh")
	}
	return path
}

// State path precedence: XDG_STATE_HOME, LocalAppData (windows only), HOME.
func stateDir() string {
	var path string
	if a := os.Getenv(xdgStateHome); a != "" {
		path = filepath.Join(a, "gh")
	} else if b := os.Getenv(localAppData); runtime.GOOS == "windows" && b != "" {
		path = filepath.Join(b, "GitHub CLI")
	} else {
		c, _ := os.UserHomeDir()
		path = filepath.Join(c, ".local", "state", "gh")
	}
	return path
}

// Data path precedence: XDG_DATA_HOME, LocalAppData (windows only), HOME.
func dataDir() string {
	var path string
	if a := os.Getenv(xdgDataHome); a != "" {
		path = filepath.Join(a, "gh")
	} else if b := os.Getenv(localAppData); runtime.GOOS == "windows" && b != "" {
		path = filepath.Join(b, "GitHub CLI")
	} else {
		c, _ := os.UserHomeDir()
		path = filepath.Join(c, ".local", "share", "gh")
	}
	return path
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

func writeFile(filename string, data []byte) error {
	err := os.MkdirAll(filepath.Dir(filename), 0771)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}
