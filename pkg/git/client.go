package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/cli/safeexec"
)

var remoteRE = regexp.MustCompile(`(.+)\s+(.+)\s+\((push|fetch)\)`)

type Client struct {
	command  func(ctx context.Context, name string, args ...string) *exec.Cmd
	dir      string
	exeErr   error
	exePath  string
	stderr   io.Writer
	stdin    io.Reader
	stdout   io.Writer
	workTree string
}

type ClientOptions struct {
	Command  func(ctx context.Context, name string, args ...string) *exec.Cmd
	Dir      string    // will be passed to commands as --git-dir
	ExePath  string    // will be resolved if none passed
	Stderr   io.Writer // will use stderr if none specified
	Stdin    io.Reader // will use stdin if none specified
	Stdout   io.Writer // will use stdout if none specified
	WorkTree string    // will be passed to commands as --work-tree
}

func NewClient(opts *ClientOptions) Client {
	if opts == nil {
		opts = &ClientOptions{}
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.Stdin == nil {
		opts.Stdin = os.Stdin
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	return Client{
		command:  opts.Command,
		dir:      opts.Dir,
		exePath:  opts.ExePath,
		stderr:   opts.Stderr,
		stdin:    opts.Stdin,
		stdout:   opts.Stdout,
		workTree: opts.WorkTree,
	}
}

func (c *Client) Command(ctx context.Context, args ...string) (*exec.Cmd, error) {
	if c.dir != "" {
		args = append([]string{"--git-dir", c.dir}, args...)
	}
	if c.workTree != "" {
		args = append([]string{"--work-tree", c.workTree}, args...)
	}
	exe, err := c.exe()
	if err != nil {
		return nil, err
	}
	var cmd *exec.Cmd
	if c.command != nil {
		cmd = c.command(ctx, exe, args...)
	} else {
		cmd = exec.CommandContext(ctx, exe, args...)
	}
	cmd.Stderr = c.stderr
	cmd.Stdin = c.stdin
	cmd.Stdout = c.stdout
	return cmd, nil
}

func (c *Client) exe() (string, error) {
	if c.exePath == "" && c.exeErr == nil {
		c.exePath, c.exeErr = safeexec.LookPath("git")
	}
	return c.exePath, c.exeErr
}

func (c *Client) Remotes(ctx context.Context) (RemoteSet, error) {
	remoteArgs := []string{"remote", "-v"}
	remoteCmd, err := c.Command(ctx, remoteArgs...)
	if err != nil {
		return nil, err
	}
	remoteOutBuf, remoteErrBuf := bytes.Buffer{}, bytes.Buffer{}
	remoteCmd.Stderr, remoteCmd.Stdout = &remoteErrBuf, &remoteOutBuf
	if err := remoteCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run git: %s. error: %w", remoteErrBuf.String(), err)
	}

	configArgs := []string{"config", "--get-regexp", `^remote\..*\.gh-resolved$`}
	configCmd, err := c.Command(ctx, configArgs...)
	if err != nil {
		return nil, err
	}
	configOutBuf, configErrBuf := bytes.Buffer{}, bytes.Buffer{}
	configCmd.Stderr, configCmd.Stdout = &configErrBuf, &configOutBuf
	if err := configCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run git: %s. error: %w", configErrBuf.String(), err)
	}

	remotes := parseRemotes(toLines(remoteOutBuf.String()))
	populateResolvedRemotes(remotes, toLines(configOutBuf.String()))
	sort.Sort(remotes)
	return remotes, nil
}

func parseRemotes(remotesStr []string) RemoteSet {
	remotes := RemoteSet{}
	for _, r := range remotesStr {
		match := remoteRE.FindStringSubmatch(r)
		if match == nil {
			continue
		}
		name := strings.TrimSpace(match[1])
		urlStr := strings.TrimSpace(match[2])
		urlType := strings.TrimSpace(match[3])

		url, err := ParseURL(urlStr)
		if err != nil {
			continue
		}
		host, owner, repo, _ := RepoInfoFromURL(url)

		var rem *Remote
		if len(remotes) > 0 {
			rem = remotes[len(remotes)-1]
			if name != rem.Name {
				rem = nil
			}
		}
		if rem == nil {
			rem = &Remote{Name: name}
			remotes = append(remotes, rem)
		}

		switch urlType {
		case "fetch":
			rem.FetchURL = url
			rem.Host = host
			rem.Owner = owner
			rem.Repo = repo
		case "push":
			rem.PushURL = url
			if rem.Host == "" {
				rem.Host = host
			}
			if rem.Owner == "" {
				rem.Owner = owner
			}
			if rem.Repo == "" {
				rem.Repo = repo
			}
		}
	}
	return remotes
}

func populateResolvedRemotes(remotes RemoteSet, resolved []string) {
	for _, l := range resolved {
		parts := strings.SplitN(l, " ", 2)
		if len(parts) < 2 {
			continue
		}
		rp := strings.SplitN(parts[0], ".", 3)
		if len(rp) < 2 {
			continue
		}
		name := rp[1]
		for _, r := range remotes {
			if r.Name == name {
				r.Resolved = parts[1]
				break
			}
		}
	}
}

func toLines(output string) []string {
	lines := strings.TrimSuffix(output, "\n")
	return strings.Split(lines, "\n")
}
