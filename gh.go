// Package gh is a library for CLI Go applications to help interface with the gh CLI tool,
// and the GitHub API.
//
// Note that the examples in this package assume gh and git are installed. They do not run in
// the Go Playground used by pkg.go.dev.
package gh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/cli/safeexec"
)

// Exec invokes a gh command in a subprocess and captures the output and error streams.
func Exec(args ...string) (stdout, stderr bytes.Buffer, err error) {
	ghExe, err := Path()
	if err != nil {
		return
	}
	err = run(context.Background(), ghExe, nil, nil, &stdout, &stderr, args)
	return
}

// ExecContext invokes a gh command in a subprocess and captures the output and error streams.
func ExecContext(ctx context.Context, args ...string) (stdout, stderr bytes.Buffer, err error) {
	ghExe, err := Path()
	if err != nil {
		return
	}
	err = run(ctx, ghExe, nil, nil, &stdout, &stderr, args)
	return
}

// Exec invokes a gh command in a subprocess with its stdin, stdout, and stderr streams connected to
// those of the parent process. This is suitable for running gh commands with interactive prompts.
func ExecInteractive(ctx context.Context, args ...string) error {
	ghExe, err := Path()
	if err != nil {
		return err
	}
	return run(ctx, ghExe, nil, os.Stdin, os.Stdout, os.Stderr, args)
}

// Path searches for an executable named "gh" in the directories named by the PATH environment variable.
// If the executable is found the result is an absolute path.
func Path() (string, error) {
	if ghExe := os.Getenv("GH_PATH"); ghExe != "" {
		return ghExe, nil
	}
	return safeexec.LookPath("gh")
}

func run(ctx context.Context, ghExe string, env []string, stdin io.Reader, stdout, stderr io.Writer, args []string) error {
	cmd := exec.CommandContext(ctx, ghExe, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if env != nil {
		cmd.Env = env
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh execution failed: %w", err)
	}
	return nil
}
