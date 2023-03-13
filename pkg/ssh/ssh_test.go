package ssh

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
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

			tr := &Translator{
				newCommand: func(exe string, args ...string) *exec.Cmd {
					args = append([]string{"-F", f.Name()}, args...)
					return exec.Command(exe, args...)
				},
			}
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

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GH_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if err := func(args []string) error {
		if len(args) < 3 || args[2] == "error" {
			return errors.New("fatal")
		}
		if args[2] == "empty.io" {
			return nil
		}
		fmt.Fprintf(os.Stdout, "hostname %s\n", args[2])
		return nil
	}(os.Args[3:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func TestTranslator_caching(t *testing.T) {
	countLookPath := 0
	countNewCommand := 0
	tr := &Translator{
		lookPath: func(s string) (string, error) {
			countLookPath++
			return "/path/to/ssh", nil
		},
		newCommand: func(exe string, args ...string) *exec.Cmd {
			args = append([]string{"-test.run=TestHelperProcess", "--", exe}, args...)
			c := exec.Command(os.Args[0], args...)
			c.Env = []string{"GH_WANT_HELPER_PROCESS=1"}
			countNewCommand++
			return c
		},
	}

	tests := []struct {
		input  string
		result string
	}{
		{
			input:  "ssh://github1.com/owner/repo.git",
			result: "github1.com",
		},
		{
			input:  "ssh://github2.com/owner/repo.git",
			result: "github2.com",
		},
		{
			input:  "ssh://empty.io/owner/repo.git",
			result: "empty.io",
		},
		{
			input:  "ssh://error/owner/repo.git",
			result: "error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			u, err := url.Parse(tt.input)
			if err != nil {
				t.Fatalf("error parsing URL: %v", err)
			}
			if res := tr.Translate(u); res.Host != tt.result {
				t.Errorf("expected github.com, got: %q", res.Host)
			}
			if res := tr.Translate(u); res.Host != tt.result {
				t.Errorf("expected github.com, got: %q (second call)", res.Host)
			}
		})
	}

	if countLookPath != 1 {
		t.Errorf("expected lookPath to happen 1 time; actual: %d", countLookPath)
	}
	if countNewCommand != len(tests) {
		t.Errorf("expected ssh command to shell out %d times; actual: %d", len(tests), countNewCommand)
	}
}
