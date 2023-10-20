package migration_test

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/config"
	"github.com/cli/go-gh/v2/pkg/config/migration"
	"github.com/stretchr/testify/require"
)

func TestMigration(t *testing.T) {
	cfg := config.ReadFromString(`
hosts:
  github.com:
    user: user1
    oauth_token: xxxxxxxxxxxxxxxxxxxx
    git_protocol: ssh
  enterprise.com:
    user: user2
    oauth_token: yyyyyyyyyyyyyyyyyyyy
    git_protocol: https
`)

	var m migration.MultiAccount
	required, err := m.Do(cfg)

	require.NoError(t, err)
	require.True(t, required, "migration should be required")

	// Do some simple checks here for depth and multiple migrations
	// but I don't really want to write a full tree traversal matcher.
	requireKeyWithValue(t, cfg, []string{"hosts", "github.com", "active_user"}, "user1")
	requireKeyWithValue(t, cfg, []string{"hosts", "github.com", "oauth_token"}, "xxxxxxxxxxxxxxxxxxxx")
	requireKeyWithValue(t, cfg, []string{"hosts", "github.com", "users", "user1", "git_protocol"}, "ssh")
	requireKeyWithValue(t, cfg, []string{"hosts", "github.com", "users", "user1", "oauth_token"}, "xxxxxxxxxxxxxxxxxxxx")

	requireKeyWithValue(t, cfg, []string{"hosts", "enterprise.com", "active_user"}, "user2")
	requireKeyWithValue(t, cfg, []string{"hosts", "enterprise.com", "oauth_token"}, "yyyyyyyyyyyyyyyyyyyy")
	requireKeyWithValue(t, cfg, []string{"hosts", "enterprise.com", "users", "user2", "git_protocol"}, "https")
	requireKeyWithValue(t, cfg, []string{"hosts", "enterprise.com", "users", "user2", "oauth_token"}, "yyyyyyyyyyyyyyyyyyyy")
}

func TestMigrationErrorsWithDeeplyNestedEntries(t *testing.T) {
	cfg := config.ReadFromString(`
hosts:
  github.com:
    user: user1
    nested:
      too: deep
`)

	var m migration.MultiAccount
	_, err := m.Do(cfg)

	require.ErrorContains(t, err, "hosts file has entries that are surprisingly deeply nested")
}

func TestMigrationReturnsNotRequiredWhenNoHostsEntry(t *testing.T) {
	cfg := config.ReadFromString(``)

	var m migration.MultiAccount
	required, err := m.Do(cfg)

	require.NoError(t, err)
	require.False(t, required, "migration should not be required when already migrated")
}

func TestMigrationReturnsNotRequredWhenEmptyHosts(t *testing.T) {
	cfg := config.ReadFromString(`
hosts:
`)

	var m migration.MultiAccount
	required, err := m.Do(cfg)

	require.NoError(t, err)
	require.False(t, required, "migration should not be required when already migrated")
}

func TestMigrationReturnsNotRequiredWhenAlreadyMigrated(t *testing.T) {
	cfg := config.ReadFromString(`
hosts:
  github.com:
    active_user: user1
    users:
      user1:
        oauth_token: xxxxxxxxxxxxxxxxxxxx
        git_protocol: ssh
`)

	var m migration.MultiAccount
	required, err := m.Do(cfg)

	require.NoError(t, err)
	require.False(t, required, "migration should not be required when already migrated")
}

func requireKeyWithValue(t *testing.T, cfg *config.Config, keys []string, value string) {
	t.Helper()

	actual, err := cfg.Get(keys)
	require.NoError(t, err)

	require.Equal(t, value, actual)
}
