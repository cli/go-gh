// Package gh is a library for CLI Go applications to help interface with the gh CLI tool,
// and the GitHub API.
//
// Note that the examples in this package assume gh and git are installed. They do not run in
// the Go Playground used by pkg.go.dev.
package gh

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/cli/safeexec"
)

// Exec gh command with provided arguments.
func Exec(args ...string) (stdOut, stdErr bytes.Buffer, err error) {
	path, err := path()
	if err != nil {
		err = fmt.Errorf("could not find gh executable in PATH. error: %w", err)
		return
	}
	return run(path, nil, args...)
}

func path() (string, error) {
	return safeexec.LookPath("gh")
}

func run(path string, env []string, args ...string) (stdOut, stdErr bytes.Buffer, err error) {
	cmd := exec.Command(path, args...)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	if env != nil {
		cmd.Env = env
	}
	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("failed to run gh: %s. error: %w", stdErr.String(), err)
		return
	}
	return
}
