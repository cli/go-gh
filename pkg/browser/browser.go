// Package browser facilitates opening of URLs in a web browser.
package browser

import (
	"io"
	"os"
	"os/exec"

	cliBrowser "github.com/cli/browser"
	"github.com/cli/go-gh/v2/pkg/config"
	"github.com/cli/safeexec"
	"github.com/google/shlex"
)

// Browser represents a web browser that can be used to open up URLs.
type Browser struct {
	launcher string
	stderr   io.Writer
	stdout   io.Writer
}

// New initializes a Browser. If a launcher is not specified
// one is determined based on environment variables or from the
// configuration file.
// The order of precedence for determining a launcher is:
// - Specified launcher;
// - GH_BROWSER environment variable;
// - browser option from configuration file;
// - BROWSER environment variable.
func New(launcher string, stdout, stderr io.Writer) *Browser {
	if launcher == "" {
		launcher = resolveLauncher()
	}
	b := &Browser{
		launcher: launcher,
		stderr:   stderr,
		stdout:   stdout,
	}
	return b
}

// Browse opens the launcher and navigates to the specified URL.
func (b *Browser) Browse(url string) error {
	return b.browse(url, nil)
}

func (b *Browser) browse(url string, env []string) error {
	if b.launcher == "" {
		return cliBrowser.OpenURL(url)
	}
	launcherArgs, err := shlex.Split(b.launcher)
	if err != nil {
		return err
	}
	launcherExe, err := safeexec.LookPath(launcherArgs[0])
	if err != nil {
		return err
	}
	args := append(launcherArgs[1:], url)
	cmd := exec.Command(launcherExe, args...)
	cmd.Stdout = b.stdout
	cmd.Stderr = b.stderr
	if env != nil {
		cmd.Env = env
	}
	return cmd.Run()
}

func resolveLauncher() string {
	if ghBrowser := os.Getenv("GH_BROWSER"); ghBrowser != "" {
		return ghBrowser
	}
	cfg, err := config.Read(nil)
	if err == nil {
		if cfgBrowser, _ := cfg.Get([]string{"browser"}); cfgBrowser != "" {
			return cfgBrowser
		}
	}
	return os.Getenv("BROWSER")
}
