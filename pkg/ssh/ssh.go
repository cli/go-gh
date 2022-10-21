// Package ssh resolves local SSH hostname aliases.
package ssh

import (
	"bufio"
	"net/url"
	"os/exec"
	"strings"
	"sync"

	"github.com/cli/safeexec"
)

type Translator struct {
	cacheMap   map[string]string
	cacheMu    sync.RWMutex
	sshPath    string
	sshPathErr error
	sshPathMu  sync.Mutex
	sshConfig  string
}

// NewTranslator initializes a new Translator instance.
func NewTranslator() *Translator {
	return &Translator{}
}

// Translate applies applicable SSH hostname aliases to the specified URL and returns the resulting URL.
func (t *Translator) Translate(u *url.URL) *url.URL {
	if u.Scheme != "ssh" {
		return u
	}
	resolvedHost, err := t.resolve(u.Hostname())
	if err != nil {
		return u
	}
	if strings.EqualFold(resolvedHost, "ssh.github.com") {
		resolvedHost = "github.com"
	}
	newURL, _ := url.Parse(u.String())
	newURL.Host = resolvedHost
	return newURL
}

func (t *Translator) resolve(hostname string) (string, error) {
	t.cacheMu.RLock()
	cached, cacheFound := t.cacheMap[strings.ToLower(hostname)]
	t.cacheMu.RUnlock()
	if cacheFound {
		return cached, nil
	}

	var sshPath string
	t.sshPathMu.Lock()
	if t.sshPath == "" && t.sshPathErr == nil {
		t.sshPath, t.sshPathErr = safeexec.LookPath("ssh")
	}
	if t.sshPathErr != nil {
		defer t.sshPathMu.Unlock()
		return t.sshPath, t.sshPathErr
	}
	sshPath = t.sshPath
	t.sshPathMu.Unlock()

	t.cacheMu.Lock()
	defer t.cacheMu.Unlock()

	sshArgs := []string{"-G", hostname}
	if t.sshConfig != "" {
		sshArgs = append(sshArgs, "-F", t.sshConfig)
	}
	sshCmd := exec.Command(sshPath, sshArgs...)
	stdout, err := sshCmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := sshCmd.Start(); err != nil {
		return "", err
	}

	var resolvedHost string
	s := bufio.NewScanner(stdout)
	for s.Scan() {
		line := s.Text()
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 && parts[0] == "hostname" {
			resolvedHost = parts[1]
		}
	}

	_ = sshCmd.Wait()
	return resolvedHost, nil
}
