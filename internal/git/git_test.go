package git

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GH_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if err := func(args []string) error {
		if args[len(args)-1] == "error" {
			return fmt.Errorf("process exited with error")
		}
		fmt.Fprintf(os.Stdout, "%v", args)
		return nil
	}(os.Args[3:]); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func TestRun(t *testing.T) {
	stdOut, stdErr, err := run(os.Args[0],
		[]string{"GH_WANT_HELPER_PROCESS=1"},
		"-test.run=TestHelperProcess", "--", "git", "status")
	assert.NoError(t, err)
	assert.Equal(t, "[git status]", stdOut.String())
	assert.Equal(t, "", stdErr.String())
}

func TestRunError(t *testing.T) {
	stdOut, stdErr, err := run(os.Args[0],
		[]string{"GH_WANT_HELPER_PROCESS=1"},
		"-test.run=TestHelperProcess", "--", "git", "status", "error")
	assert.EqualError(t, err, "failed to run git: process exited with error. error: exit status 1")
	assert.Equal(t, "", stdOut.String())
	assert.Equal(t, "process exited with error", stdErr.String())
}
