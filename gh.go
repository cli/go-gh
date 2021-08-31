package gh

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/cli/safeexec"
)

func Path() (string, error) {
	return safeexec.LookPath("gh")
}

func Exec(args ...string) (stdOut, stdErr bytes.Buffer, err error) {
	path, err := Path()
	if err != nil {
		err = fmt.Errorf("could not find gh executable in PATH. error: %w", err)
		return
	}
	return run(path, nil, args...)
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
