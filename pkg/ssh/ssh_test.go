package ssh

import (
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh/v2/internal/testutils"
	"github.com/cli/safeexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testutils.RunFakeSSHIfRequested()
	os.Exit(m.Run())
}

func TestTranslatorWithRealSSH(t *testing.T) {
	if _, err := safeexec.LookPath("ssh"); err != nil {
		t.Skip("no ssh found on system")
	}

	tests := []struct {
		name      string
		sshConfig string
		input     string
		want      string
	}{
		{
			name: "translates SSH URL",
			sshConfig: heredoc.Doc(`
				Host github-*
					Hostname github.com
			`),
			input: "ssh://git@github-foo/owner/repo.git",
			want:  "ssh://git@github.com/owner/repo.git",
		},
		{
			name: "does not translate HTTPS URL",
			sshConfig: heredoc.Doc(`
				Host github-*
					Hostname github.com
			`),
			input: "https://github-foo/owner/repo.git",
			want:  "https://github-foo/owner/repo.git",
		},
		{
			name: "treats ssh.github.com as github.com",
			sshConfig: heredoc.Doc(`
				Host github.com
					Hostname ssh.github.com
			`),
			input: "ssh://git@github.com/owner/repo.git",
			want:  "ssh://git@github.com/owner/repo.git",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given an ssh config
			configPath := filepath.Join(t.TempDir(), "ssh-config")
			require.NoError(t, os.WriteFile(configPath, []byte(tt.sshConfig), 0o600))
			tr := &Translator{
				newCommand: func(exe string, args ...string) *exec.Cmd {
					return exec.Command(exe, append([]string{"-F", configPath}, args...)...)
				},
			}

			// When a URL is translated
			got := tr.Translate(mustParseURL(t, tt.input))

			// Then the URL is translated as expected
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestTranslatorCachesHostname(t *testing.T) {
	// Given a translator that has already resolved an alias
	testutils.StubSSH(t, map[string]testutils.SSHResponse{"github-work": testutils.SSHReportsHostname("github.com")})
	tr := NewTranslator()
	tr.Translate(mustParseURL(t, "ssh://git@github-work/owner/repo.git"))

	// When ssh would now resolve the alias differently
	testutils.StubSSH(t, map[string]testutils.SSHResponse{"github-work": testutils.SSHReportsHostname("ghe.example")})
	got := tr.Translate(mustParseURL(t, "ssh://git@github-work/owner/repo.git"))

	// Then the first resolution is reused
	assert.Equal(t, "ssh://git@github.com/owner/repo.git", got.String())
}

func TestTranslatorResolvesEachHostSeparately(t *testing.T) {
	// Given a translator that has already resolved one alias
	testutils.StubSSH(t, map[string]testutils.SSHResponse{
		"github-work":       testutils.SSHReportsHostname("github.com"),
		"github-enterprise": testutils.SSHReportsHostname("ghe.example"),
	})
	tr := NewTranslator()
	tr.Translate(mustParseURL(t, "ssh://git@github-work/owner/repo.git"))

	// When a URL for a different alias is translated
	got := tr.Translate(mustParseURL(t, "ssh://git@github-enterprise/owner/repo.git"))

	// Then that alias is resolved on its own
	assert.Equal(t, "ssh://git@ghe.example/owner/repo.git", got.String())
}

func TestTranslatorKeepsHostWhenSSHReportsNoHostname(t *testing.T) {
	// Given ssh reports no hostname for a host
	testutils.StubSSH(t, map[string]testutils.SSHResponse{"github-work": testutils.SSHReportsNoHostname()})

	// When a URL for that host is translated
	got := NewTranslator().Translate(mustParseURL(t, "ssh://git@github-work/owner/repo.git"))

	// Then the URL is unchanged
	assert.Equal(t, "ssh://git@github-work/owner/repo.git", got.String())
}

func TestTranslatorKeepsHostWhenSSHFails(t *testing.T) {
	// Given ssh fails for a host
	testutils.StubSSH(t, map[string]testutils.SSHResponse{"github-work": testutils.SSHFails()})

	// When a URL for that host is translated
	got := NewTranslator().Translate(mustParseURL(t, "ssh://git@github-work/owner/repo.git"))

	// Then the URL is unchanged
	assert.Equal(t, "ssh://git@github-work/owner/repo.git", got.String())
}

func TestTranslatorKeepsHostWhenSSHIsNotInstalled(t *testing.T) {
	// Given ssh is not on PATH
	t.Setenv("PATH", t.TempDir())

	// When an SSH URL is translated
	got := NewTranslator().Translate(mustParseURL(t, "ssh://git@github-work/owner/repo.git"))

	// Then the URL is unchanged
	assert.Equal(t, "ssh://git@github-work/owner/repo.git", got.String())
}

func mustParseURL(t *testing.T, s string) *url.URL {
	t.Helper()
	u, err := url.Parse(s)
	require.NoError(t, err)
	return u
}
