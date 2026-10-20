package testutils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type sshResponseKind string

const (
	sshReportsHostname   sshResponseKind = "hostname"
	sshReportsNoHostname sshResponseKind = "no-hostname"
	sshFails             sshResponseKind = "fails"
)

// SSHResponse is how the fake ssh installed by StubSSH answers `ssh -G` for a host.
// Build one with SSHReportsHostname, SSHReportsNoHostname, or SSHFails. The zero value is invalid.
type SSHResponse struct {
	kind     sshResponseKind
	hostname string
}

// SSHReportsHostname makes the fake ssh report hostname for a host.
func SSHReportsHostname(hostname string) SSHResponse {
	return SSHResponse{kind: sshReportsHostname, hostname: hostname}
}

// SSHReportsNoHostname makes the fake ssh succeed without reporting a hostname for a host.
func SSHReportsNoHostname() SSHResponse {
	return SSHResponse{kind: sshReportsNoHostname}
}

// SSHFails makes the fake ssh exit with an error for a host. It still prints a hostname first,
// so callers must honour the exit status rather than the output.
func SSHFails() SSHResponse {
	return SSHResponse{kind: sshFails}
}

type wireSSHResponse struct {
	Kind     sshResponseKind `json:"kind"`
	Hostname string          `json:"hostname,omitempty"`
}

const failingSSHHostname = "failed-ssh.invalid"

const fakeSSHResponsesEnv = "GO_GH_TEST_FAKE_SSH_RESPONSES"

// StubSSH puts a fake ssh first on PATH. Its `ssh -G HOST` answers with responses[HOST], or reports
// HOST itself when unmapped, like ssh without a matching config. The fake is the running test binary,
// so the test package's TestMain must call RunFakeSSHIfRequested.
func StubSSH(t *testing.T, responses map[string]SSHResponse) {
	t.Helper()
	wire := make(map[string]wireSSHResponse, len(responses))
	for host, r := range responses {
		require.NotEmpty(t, r.kind, "fake ssh response for %q must be built with an SSH* constructor", host)
		wire[host] = wireSSHResponse{Kind: r.kind, Hostname: r.hostname}
	}
	encoded, err := json.Marshal(wire)
	require.NoError(t, err, "encoding fake ssh responses")
	dir := t.TempDir()
	name := "ssh"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	copyTestBinary(t, filepath.Join(dir, name))
	t.Setenv(fakeSSHResponsesEnv, string(encoded))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// RunFakeSSHIfRequested acts as `ssh -G` and exits when this test binary was started as the fake ssh
// installed by StubSSH, that is, when it runs under the name ssh with the fake's responses set.
// Otherwise it returns, so other re-execs of the test binary are unaffected. Call it first in TestMain.
func RunFakeSSHIfRequested() {
	if strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe") != "ssh" {
		return
	}
	encoded, ok := os.LookupEnv(fakeSSHResponsesEnv)
	if !ok {
		return
	}
	var responses map[string]wireSSHResponse
	if err := json.Unmarshal([]byte(encoded), &responses); err != nil {
		fmt.Fprintf(os.Stderr, "fake ssh: decoding responses: %v\n", err)
		os.Exit(1)
	}
	args := os.Args[1:]
	if len(args) != 2 || args[0] != "-G" {
		fmt.Fprintf(os.Stderr, "fake ssh: unexpected arguments %q\n", args)
		os.Exit(1)
	}
	host := args[1]
	response, mapped := responses[host]
	if !mapped {
		fmt.Printf("hostname %s\n", host)
		os.Exit(0)
	}
	switch response.Kind {
	case sshReportsHostname:
		fmt.Printf("hostname %s\n", response.Hostname)
	case sshReportsNoHostname:
	case sshFails:
		fmt.Printf("hostname %s\n", failingSSHHostname)
		fmt.Fprintf(os.Stderr, "fake ssh: failing for %s\n", host)
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "fake ssh: unknown response kind %q for %s\n", response.Kind, host)
		os.Exit(1)
	}
	os.Exit(0)
}

func copyTestBinary(t *testing.T, dst string) {
	t.Helper()
	src, err := os.Executable()
	require.NoError(t, err, "locating test binary")
	in, err := os.Open(src)
	require.NoError(t, err, "opening test binary")
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0o755)
	require.NoError(t, err, "creating fake ssh")
	_, err = io.Copy(out, in)
	closeErr := out.Close()
	require.NoError(t, err, "copying test binary")
	require.NoError(t, closeErr, "closing fake ssh")
}
