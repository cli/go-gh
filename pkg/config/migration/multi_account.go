package migration

import (
	"errors"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/config"
)

type CowardlyRefusalError struct {
	Reason string
}

func (e CowardlyRefusalError) Error() string {
	// Consider whether we should add a call to action here like "open an issue with the contents of your redacted hosts.yml"
	return fmt.Sprintf("cowardly refusing to continue with multi account migration: %s", e.Reason)
}

var hostsKey = []string{"hosts"}

// This migration exists to take a hosts section of the following structure:
//
//	github.com:
//	  user: williammartin
//	  git_protocol: https
//	  editor: vim
//	github.localhost:
//	  user: monalisa
//	  git_protocol: https
//    oauth_token: xyz
//
// We want this to migrate to something like:
//
// github.com:
//   active_user: williammartin
//	 users:
//	   williammartin:
//	     git_protocol: https
//	     editor: vim
//
// github.localhost:
//	 active_user: monalisa
//   oauth_token: xyz
//   users:
//	   monalisa:
//	     git_protocol: https
//	     oauth_token: xyz
//
// The reason for this is that we can then add new users under a host.
// Note that it's important we duplicate and hoist the oauth_token with an unchanged key
// so that existing users of the go-gh auth package don't break. In practice
// it represents the "active_oauth_token" if there is one (due to insecure storage).

type MultiAccount struct{}

func (m MultiAccount) Do(c *config.Config) (bool, error) {
	hostnames, err := c.Keys(hostsKey)
	// [github.com, github.localhost]
	var keyNotFoundError *config.KeyNotFoundError
	if errors.As(err, &keyNotFoundError) {
		return false, nil
	}
	if err != nil {
		return false, CowardlyRefusalError{"couldn't get hosts configuration"}
	}

	// If there are no hosts then it doesn't matter whether we migrate or not,
	// so lets avoid any confusion and say there's no migration required.
	if len(hostnames) == 0 {
		return false, nil
	}

	// Check to see if we are on the old version of the hosts file by getting the user key
	_, err = c.Get(append(hostsKey, hostnames[0], "user"))
	hasPreMigrationData := err == nil

	// And check to see whether we have a new version of the hosts file by getting the active_user key
	_, err = c.Get(append(hostsKey, hostnames[0], "active_user"))
	hasPostMigrationData := err == nil

	// If we have both then ther user has been naughty. They must have gone back and used an older
	// version of gh with a newer hosts file and now they are going to be in trouble!
	if hasPreMigrationData && hasPostMigrationData {
		return false, CowardlyRefusalError{"hosts file has a mix of old and new formats, did you use an older version of gh?"}
	}

	// Otherwise if they have just the new version, then we don't need to migrate
	if hasPostMigrationData {
		return false, nil
	}

	// Otherwise let's get to the business of migrating!
	for _, hostname := range hostnames {
		configEntryKeys, err := c.Keys(append(hostsKey, hostname))
		// e.g. [user, git_protocol, editor, ouath_token]
		if err != nil {
			return false, CowardlyRefusalError{fmt.Sprintf("couldn't get host configuration despite %q existing", hostname)}
		}

		// Get the user so that we can nest under it in future
		username, err := c.Get(append(hostsKey, hostname, "user"))
		if err != nil {
			return false, CowardlyRefusalError{fmt.Sprintf("couldn't get user name for %q", hostname)}
		}

		for _, configEntryKey := range configEntryKeys {
			// We would expect that these keys map directly to values
			// e.g. [williammartin, https, vim, gho_xyz...] but it's possible that a manually
			// edited config file might nest further but we don't support that.
			//
			// We could consider throwing away the nested values, but I suppose
			// I'd rather make the user take a destructive action even if we have a backup.
			// If they have configuration here, it's probably for a reason.
			keys, err := c.Keys(append(hostsKey, hostname, configEntryKey))
			if err == nil && len(keys) > 0 {
				return false, CowardlyRefusalError{"hosts file has entries that are surprisingly deeply nested"}
			}

			configEntryValue, err := c.Get(append(hostsKey, hostname, configEntryKey))
			if err != nil {
				return false, CowardlyRefusalError{fmt.Sprintf("couldn't get configuration entry value despite %q / %q existing", hostname, configEntryKey)}
			}

			// Remove all these entries, because we are going to move
			if err := c.Remove(append(hostsKey, hostname, configEntryKey)); err != nil {
				return false, CowardlyRefusalError{fmt.Sprintf("couldn't remove configuration entry %q despite %q / %q existing", configEntryKey, hostname, configEntryKey)}
			}

			// If this is the user key, we don't need to do anything with it because it's
			// now part of the final structure.
			if configEntryKey == "user" {
				continue
			}

			// And if it's the oauth_token, we want to duplicate it up a layer to ensure the go-gh auth
			// package continues to work.
			if configEntryKey == "oauth_token" {
				c.Set(append(hostsKey, hostname, "oauth_token"), configEntryValue)
			}

			// Set these entries in their new location under the user
			c.Set(append(hostsKey, hostname, "users", username, configEntryKey), configEntryValue)
		}

		// And after migrating all the existing values, we'll add one new "active" key to indicate the user
		// that is active for this host after the migration.
		c.Set(append(hostsKey, hostname, "active_user"), username)
	}

	return true, nil
}
