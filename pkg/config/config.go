// Package config is a set of types for interacting with the gh configuration files.
// Note: This package is intended for use only in gh, any other use cases are subject
// to breakage and non-backwards compatible updates.
package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/cli/go-gh/v2/internal/yamlmap"
)

const (
	appData       = "AppData"
	ghConfigDir   = "GH_CONFIG_DIR"
	localAppData  = "LocalAppData"
	xdgConfigHome = "XDG_CONFIG_HOME"
	xdgDataHome   = "XDG_DATA_HOME"
	xdgStateHome  = "XDG_STATE_HOME"
)

var (
	cfg     *Config
	once    sync.Once
	loadErr error
)

// Config is a in memory representation of the gh configuration files.
// It can be thought of as map where entries consist of a key that
// correspond to either a string value or a map value, allowing for
// multi-level maps.
type Config struct {
	entries *yamlmap.Map
	mu      sync.RWMutex
}

func (c *Config) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.entries.String()
}

// TODO: consider passing in something that actually handles the migration
// like a v1v2Migrator, and then we can also do v2v1Migrator to go backwards
// For now, I'm just spiking and gotta spike hard.
//
// This migration exists to take a hosts section of the following structure:
//
//	github.com:
//	  user: williammartin
//	  git_protocol: https
//	  editor: vim
//	github.localhost:
//	  user: monalisa
//	  git_protocol: https
//
// We want this to migrate to something like:
//
// ```
// github.com:
//
//	 active: williammartin
//		users:
//		  williammartin:
//		    active: true
//		    git_protocol: https
//		    editor: vim
//
// github.localhost:
//
//	 active: monalisa
//		  monalisa:
//		    active: true
//		    git_protocol: https
//
// ```
//
// The reason for this is that we can then add new users under a host.
// For some reason gofmt is messing up the structure of that yaml data above which is awful.
func (c *Config) Migrate() (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hostsVersionAfterMigration := "2"

	// Get the entry in the yamlmap for hosts, which is a map that has a structure like:
	//
	// 	github.com:
	//    user: williammartin
	//    git_protocol: https
	//    editor: vim
	//  github.localhost:
	//    user: monalisa
	//    git_protocol: https
	hostsEntry, err := c.entries.FindEntry("hosts")
	if err != nil {
		return false, &KeyNotFoundError{"hosts"}
	}

	// Let's check to see whether our migration has already been applied.
	// The only error that can be returned here is ErrNotFound.
	if versionEntry, err := hostsEntry.FindEntry("version"); err == nil {
		if versionEntry.Value == hostsVersionAfterMigration {
			return false, nil
		}
	}

	// Create a new map in which to store our migrated data
	migratedHostsEntry := yamlmap.MapValue()

	// Iterate over the keys of the host entry, which are the hostnames like
	// [github.com, ghe.io]
	// TODO: Consider creating a .Entries() method that returns a struct
	// containing Key and Value, so that we don't have to .FindEntry()
	// for each key. I did this and it works, but I wanted to avoid extraneous
	// details in this spike.
	for _, hostKey := range hostsEntry.Keys() {

		// Find the entry for that host, which is a map that has a structure like:
		//
		// user: williammartin
		// git_protocol: https
		// editor: vim
		hostEntry, err := hostsEntry.FindEntry(hostKey)
		if err != nil {
			return true, &KeyNotFoundError{hostKey}
		}

		// Create new maps in which to store the migrated data. Note that we
		// are doing something special with the user entry, because we need
		// to add two new layers in our map that wasn't there before.
		migratedHostEntry := yamlmap.MapValue()
		migratedUsersEntry := yamlmap.MapValue()
		migratedHostEntry.SetEntry("users", migratedUsersEntry)

		migratedUserEntry := yamlmap.MapValue()
		var username string

		// Iterate over the keys of the host entry, which are config values like:
		// [user, editor, git_protocol]
		for _, hostCfgKey := range hostEntry.Keys() {

			// Find the entry for that config value, which should be string values like
			// williammartin, https or vim
			hostCfgEntry, err := hostEntry.FindEntry(hostCfgKey)
			if err != nil {
				return true, &KeyNotFoundError{hostCfgKey}
			}

			// If this is the user entry, then we'll store the username for our new layer
			// and then since this is a migration and we know we only had one user before this,
			// we'll add an "active" key with value "true".
			if hostCfgKey == "user" {
				username = hostCfgEntry.Value
				continue
			}

			// If this wasn't a user entry, then we'll take that configuration data and put
			// it under our migrated user entry
			migratedUserEntry.SetEntry(hostCfgKey, yamlmap.StringValue(hostCfgEntry.Value))
		}

		// Link the username key we stored earlier with our migrated user entry, on the migrated host entry
		// And set the active user to this username, since we know we only had one user before the migration
		migratedUsersEntry.AddEntry(username, migratedUserEntry)
		migratedHostEntry.SetEntry("active_user", yamlmap.StringValue(username))

		// And link our migrated host to the migrated hosts entry
		migratedHostsEntry.SetEntry(hostKey, migratedHostEntry)
	}

	// Link our migrated hosts to the top level hosts key
	c.entries.SetEntry("hosts", migratedHostsEntry)
	// Set a version so that we know we've applied this migration (kind of gross and we'll need to account for it,
	// by not treating it as a host later)
	migratedHostsEntry.SetEntry("version", yamlmap.StringValue(hostsVersionAfterMigration))

	// For this spike, let's add a revert field that we can use to easily revert this migration
	migratedHostsEntry.SetEntry("revert", hostsEntry)

	// Finally, let's write our hosts file and mark that entry as unmodified so it doesn't get
	// written again.
	if err = writeFile(hostsConfigFile(), []byte(migratedHostsEntry.String())); err != nil {
		return true, err
	}
	migratedHostsEntry.SetUnmodified()

	return true, nil
}

// The opposite of Migrate
func (c *Config) Revert() (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get the entry in the yamlmap for hosts, which is a map that has a structure like:
	hostsEntry, err := c.entries.FindEntry("hosts")
	if err != nil {
		return false, &KeyNotFoundError{"hosts"}
	}

	// Find the "revert" key, if it doesn't exist then the migration hasn't been applied (probably)
	// The only error that can be returned here is ErrNotFound
	revertEntry, err := hostsEntry.FindEntry("revert")
	if err != nil {
		return false, nil
	}

	// Link the value of the revert entry to the top level hosts (overwriting the migrated hosts)
	c.entries.SetEntry("hosts", revertEntry)

	// Finally, let's write our hosts file and mark that entry as unmodified so it doesn't get
	// written again.
	if err = writeFile(hostsConfigFile(), []byte(revertEntry.String())); err != nil {
		return true, err
	}
	revertEntry.SetUnmodified()

	return false, nil
}

// Get a string value from a Config.
// The keys argument is a sequence of key values so that nested
// entries can be retrieved. A undefined string will be returned
// if trying to retrieve a key that corresponds to a map value.
// Returns "", KeyNotFoundError if any of the keys can not be found.
func (c *Config) Get(keys []string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m := c.entries
	for _, key := range keys {
		var err error
		m, err = m.FindEntry(key)
		if err != nil {
			return "", &KeyNotFoundError{key}
		}
	}
	return m.Value, nil
}

// Keys enumerates a Config's keys.
// The keys argument is a sequence of key values so that nested
// map values can be have their keys enumerated.
// Returns nil, KeyNotFoundError if any of the keys can not be found.
func (c *Config) Keys(keys []string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m := c.entries
	for _, key := range keys {
		var err error
		m, err = m.FindEntry(key)
		if err != nil {
			return nil, &KeyNotFoundError{key}
		}
	}
	return m.Keys(), nil
}

// Remove an entry from a Config.
// The keys argument is a sequence of key values so that nested
// entries can be removed. Removing an entry that has nested
// entries removes those also.
// Returns KeyNotFoundError if any of the keys can not be found.
func (c *Config) Remove(keys []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	m := c.entries
	for i := 0; i < len(keys)-1; i++ {
		var err error
		key := keys[i]
		m, err = m.FindEntry(key)
		if err != nil {
			return &KeyNotFoundError{key}
		}
	}
	err := m.RemoveEntry(keys[len(keys)-1])
	if err != nil {
		return &KeyNotFoundError{keys[len(keys)-1]}
	}
	return nil
}

// Set a string value in a Config.
// The keys argument is a sequence of key values so that nested
// entries can be set. If any of the keys do not exist they will
// be created.
func (c *Config) Set(keys []string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	m := c.entries
	for i := 0; i < len(keys)-1; i++ {
		key := keys[i]
		entry, err := m.FindEntry(key)
		if err != nil {
			entry = yamlmap.MapValue()
			m.AddEntry(key, entry)
		}
		m = entry
	}
	m.SetEntry(keys[len(keys)-1], yamlmap.StringValue(value))
}

// Read gh configuration files from the local file system and
// return a Config.
var Read = func() (*Config, error) {
	once.Do(func() {
		cfg, loadErr = load(generalConfigFile(), hostsConfigFile())
	})
	return cfg, loadErr
}

// ReadFromString takes a yaml string and returns a Config.
// Note: This is only used for testing, and should not be
// relied upon in production.
func ReadFromString(str string) *Config {
	m, _ := mapFromString(str)
	if m == nil {
		m = yamlmap.MapValue()
	}
	return &Config{entries: m}
}

// Write gh configuration files to the local file system.
// It will only write gh configuration files that have been modified
// since last being read.
func Write(c *Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	hosts, err := c.entries.FindEntry("hosts")
	if err == nil && hosts.IsModified() {
		err := writeFile(hostsConfigFile(), []byte(hosts.String()))
		if err != nil {
			return err
		}
		hosts.SetUnmodified()
	}

	if c.entries.IsModified() {
		// Hosts gets written to a different file above so remove it
		// before writing and add it back in after writing.
		hostsMap, hostsErr := c.entries.FindEntry("hosts")
		if hostsErr == nil {
			_ = c.entries.RemoveEntry("hosts")
		}
		err := writeFile(generalConfigFile(), []byte(c.entries.String()))
		if err != nil {
			return err
		}
		c.entries.SetUnmodified()
		if hostsErr == nil {
			c.entries.AddEntry("hosts", hostsMap)
		}
	}

	return nil
}

func load(generalFilePath, hostsFilePath string) (*Config, error) {
	generalMap, err := mapFromFile(generalFilePath)
	if err != nil && !os.IsNotExist(err) {
		if errors.Is(err, yamlmap.ErrInvalidYaml) ||
			errors.Is(err, yamlmap.ErrInvalidFormat) {
			return nil, &InvalidConfigFileError{Path: generalFilePath, Err: err}
		}
		return nil, err
	}

	if generalMap == nil || generalMap.Empty() {
		generalMap, _ = mapFromString(defaultGeneralEntries)
	}

	hostsMap, err := mapFromFile(hostsFilePath)
	if err != nil && !os.IsNotExist(err) {
		if errors.Is(err, yamlmap.ErrInvalidYaml) ||
			errors.Is(err, yamlmap.ErrInvalidFormat) {
			return nil, &InvalidConfigFileError{Path: hostsFilePath, Err: err}
		}
		return nil, err
	}

	if hostsMap != nil && !hostsMap.Empty() {
		generalMap.AddEntry("hosts", hostsMap)
	}

	return &Config{entries: generalMap}, nil
}

func generalConfigFile() string {
	return filepath.Join(ConfigDir(), "config.yml")
}

func hostsConfigFile() string {
	return filepath.Join(ConfigDir(), "hosts.yml")
}

func mapFromFile(filename string) (*yamlmap.Map, error) {
	data, err := readFile(filename)
	if err != nil {
		return nil, err
	}
	return yamlmap.Unmarshal(data)
}

func mapFromString(str string) (*yamlmap.Map, error) {
	return yamlmap.Unmarshal([]byte(str))
}

// Config path precedence: GH_CONFIG_DIR, XDG_CONFIG_HOME, AppData (windows only), HOME.
func ConfigDir() string {
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
func StateDir() string {
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
func DataDir() string {
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

// CacheDir returns the default path for gh cli cache.
func CacheDir() string {
	return filepath.Join(os.TempDir(), "gh-cli-cache")
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

func writeFile(filename string, data []byte) (writeErr error) {
	if writeErr = os.MkdirAll(filepath.Dir(filename), 0771); writeErr != nil {
		return
	}
	var file *os.File
	if file, writeErr = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600); writeErr != nil {
		return
	}
	defer func() {
		if err := file.Close(); writeErr == nil && err != nil {
			writeErr = err
		}
	}()
	_, writeErr = file.Write(data)
	return
}

var defaultGeneralEntries = `
# What protocol to use when performing git operations. Supported values: ssh, https
git_protocol: https
# What editor gh should run when creating issues, pull requests, etc. If blank, will refer to environment.
editor:
# When to interactively prompt. This is a global config that cannot be overridden by hostname. Supported values: enabled, disabled
prompt: enabled
# A pager program to send command output to, e.g. "less". Set the value to "cat" to disable the pager.
pager:
# Aliases allow you to create nicknames for gh commands
aliases:
  co: pr checkout
# The path to a unix socket through which send HTTP connections. If blank, HTTP traffic will be handled by net/http.DefaultTransport.
http_unix_socket:
# What web browser gh should use when opening URLs. If blank, will refer to environment.
browser:
`
