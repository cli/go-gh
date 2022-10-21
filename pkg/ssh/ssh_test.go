package ssh

import (
	"net/url"
	"os"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/safeexec"
)

func TestTranslator(t *testing.T) {
	if _, err := safeexec.LookPath("ssh"); err != nil {
		t.Skip("no ssh found on system")
	}

	tests := []struct {
		name      string
		sshConfig string
		arg       string
		want      string
	}{
		{
			name: "translate SSH URL",
			sshConfig: heredoc.Doc(`
				Host github-*
					Hostname github.com
			`),
			arg:  "ssh://git@github-foo/owner/repo.git",
			want: "ssh://git@github.com/owner/repo.git",
		},
		{
			name: "does not translate HTTPS URL",
			sshConfig: heredoc.Doc(`
				Host github-*
					Hostname github.com
			`),
			arg:  "https://github-foo/owner/repo.git",
			want: "https://github-foo/owner/repo.git",
		},
		{
			name: "treats ssh.github.com as github.com",
			sshConfig: heredoc.Doc(`
				Host github.com
					Hostname ssh.github.com
			`),
			arg:  "ssh://git@github.com/owner/repo.git",
			want: "ssh://git@github.com/owner/repo.git",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "ssh-config.*")
			if err != nil {
				t.Fatalf("error creating file: %v", err)
			}
			_, err = f.WriteString(tt.sshConfig)
			_ = f.Close()
			if err != nil {
				t.Fatalf("error writing ssh config: %v", err)
			}

			tr := &Translator{sshConfig: f.Name()}
			u, err := url.Parse(tt.arg)
			if err != nil {
				t.Fatalf("error parsing URL: %v", err)
			}
			res := tr.Translate(u)
			if got := res.String(); got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
