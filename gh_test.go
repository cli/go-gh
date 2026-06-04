package gh

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

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

func TestHelperProcessLongRunning(t *testing.T) {
	if os.Getenv("GH_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args[3:]
	fmt.Fprintf(os.Stdout, "%v", args)
	fmt.Fprint(os.Stderr, "going to sleep...")
	time.Sleep(10 * time.Second)
	fmt.Fprint(os.Stderr, "...going to exit")
	os.Exit(0)
}

func TestRun(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run(context.TODO(), os.Args[0], []string{"GH_WANT_HELPER_PROCESS=1"}, nil, &stdout, &stderr,
		[]string{"-test.run=TestHelperProcess", "--", "gh", "issue", "list"})
	assert.NoError(t, err)
	assert.Equal(t, "[gh issue list]", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestRunError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run(context.TODO(), os.Args[0], []string{"GH_WANT_HELPER_PROCESS=1"}, nil, &stdout, &stderr,
		[]string{"-test.run=TestHelperProcess", "--", "gh", "error"})
	assert.EqualError(t, err, "gh execution failed: exit status 1")
	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "process exited with error", stderr.String())
}

func TestRunInteractiveContextCanceled(t *testing.T) {
	// pass current time to ensure that deadline has already passed
	ctx, cancel := context.WithDeadline(context.Background(), time.Now())
	cancel()
	err := run(ctx, os.Args[0], []string{"GH_WANT_HELPER_PROCESS=1"}, nil, nil, nil,
		[]string{"-test.run=TestHelperProcessLongRunning", "--", "gh", "issue", "list"})
	assert.EqualError(t, err, "gh execution failed: context deadline exceeded")
}
