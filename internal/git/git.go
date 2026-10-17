// Package git provides centralized utilities and command builders for invoking git.
package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// Runner executes a command in dir, writing its stdout/stderr to the provided writers.
type Runner interface {
	Run(ctx context.Context, dir string, command []string, stdout, stderr io.Writer) error
}

// processRunner executes commands as real OS processes.
type processRunner struct{}

func (processRunner) Run(ctx context.Context, dir string, command []string, stdout, stderr io.Writer) error {
	if len(command) == 0 {
		return fmt.Errorf("runner: empty command")
	}
	//nolint:gosec // intentional command execution
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// DefaultCommand is the default binary name for git.
const DefaultCommand = "git"

// Command represents a git command configuration.
type Command struct {
	// Bin is the path or name of the git binary. Defaults to DefaultCommand ("git") if empty.
	Bin string
	// Runner executes the command.
	Runner Runner
}

// New creates a new Command with the given binary and runner.
func New(bin string, r Runner) *Command {
	if bin == "" {
		bin = DefaultCommand
	}
	return &Command{
		Bin:    bin,
		Runner: r,
	}
}

// GitBin returns the git binary name/path.
func (c *Command) GitBin() string {
	if c != nil && c.Bin != "" {
		return c.Bin
	}
	return DefaultCommand
}

func (c *Command) runner() Runner {
	if c != nil && c.Runner != nil {
		return c.Runner
	}
	return processRunner{}
}

// Run executes a git command in dir with sub-arguments args.
func (c *Command) Run(ctx context.Context, dir string, args []string, stdout, stderr io.Writer) error {
	fullCmd := append([]string{c.GitBin()}, args...)
	return c.runner().Run(ctx, dir, fullCmd, stdout, stderr)
}

// Output executes a git command in dir and returns its trimmed stdout string.
func (c *Command) Output(ctx context.Context, dir string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	if err := c.Run(ctx, dir, args, &stdout, &stderr); err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout.String()), nil
}

// GrepCmd constructs argv for "git grep -H" with the given arguments.
func (c *Command) GrepCmd(args ...string) []string {
	return append([]string{c.GitBin(), "grep", "-H"}, args...)
}

// PullCmd constructs argv for "git pull" with the given flags.
func (c *Command) PullCmd(flags ...string) []string {
	return append([]string{c.GitBin(), "pull"}, flags...)
}

// FetchCmd constructs argv for "git fetch" with the given flags.
func (c *Command) FetchCmd(flags ...string) []string {
	return append([]string{c.GitBin(), "fetch"}, flags...)
}

// LsFilesCmd constructs argv for "git ls-files" with the given flags.
func (c *Command) LsFilesCmd(flags ...string) []string {
	return append([]string{c.GitBin(), "ls-files"}, flags...)
}

// DiffCmd constructs argv for "git diff" with the given flags.
// If interactive is false, it injects "--no-pager" before "diff".
func (c *Command) DiffCmd(interactive bool, flags ...string) []string {
	if interactive {
		return append([]string{c.GitBin(), "diff"}, flags...)
	}
	return append([]string{c.GitBin(), "--no-pager", "diff"}, flags...)
}

// StatusCmd constructs argv for "git status" with the given flags.
// If flags is empty, it defaults to []string{"--short"}.
func (c *Command) StatusCmd(flags ...string) []string {
	if len(flags) == 0 {
		return []string{c.GitBin(), "status", "--short"}
	}
	return append([]string{c.GitBin(), "status"}, flags...)
}

// IsDirty checks if the git repository in dir has uncommitted changes or untracked files.
func (c *Command) IsDirty(ctx context.Context, dir string) (bool, error) {
	out, err := c.Output(ctx, dir, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("git is dirty: %w", err)
	}
	return len(strings.TrimSpace(out)) > 0, nil
}

// CurrentBranch returns the currently checked-out branch in dir.
func (c *Command) CurrentBranch(ctx context.Context, dir string) (string, error) {
	out, err := c.Output(ctx, dir, "branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("git current branch: %w", err)
	}
	return out, nil
}

// Checkout switches the branch in dir to target.
func (c *Command) Checkout(ctx context.Context, dir, target string, stderr io.Writer) error {
	return c.Run(ctx, dir, []string{"checkout", target}, io.Discard, stderr)
}

// ResolveDefaultBranch identifies the default branch for the repository in dir:
// 1. refs/remotes/origin/HEAD symbolic ref (e.g. "origin/main" -> "main")
// 2. init.defaultBranch config
// 3. refs/heads/main exists
// 4. refs/heads/master exists
func (c *Command) ResolveDefaultBranch(ctx context.Context, dir string) (string, error) {
	// 1. refs/remotes/origin/HEAD
	if ref, err := c.Output(ctx, dir, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
		if after, ok := strings.CutPrefix(ref, "origin/"); ok && after != "" {
			return after, nil
		}
	}

	// 2. init.defaultBranch
	if branch, err := c.Output(ctx, dir, "config", "--get", "init.defaultBranch"); err == nil && branch != "" {
		return branch, nil
	}

	// 3. refs/heads/main
	mainCmd := []string{"show-ref", "--verify", "--quiet", "refs/heads/main"}
	if err := c.Run(ctx, dir, mainCmd, io.Discard, io.Discard); err == nil {
		return "main", nil
	}

	// 4. refs/heads/master
	masterCmd := []string{"show-ref", "--verify", "--quiet", "refs/heads/master"}
	if err := c.Run(ctx, dir, masterCmd, io.Discard, io.Discard); err == nil {
		return "master", nil
	}

	return "", fmt.Errorf("unable to determine default branch")
}

// CloneCmd constructs argv for "git clone -- <uri> <dir>" with the given URI and destination directory.
func (c *Command) CloneCmd(uri, dir string) []string {
	return []string{c.GitBin(), "clone", "--", uri, dir}
}

// Clone clones uri into targetDir, routing command output to stderr.
func (c *Command) Clone(ctx context.Context, uri, targetDir string, stderr io.Writer) error {
	return c.Run(ctx, "", c.CloneCmd(uri, targetDir)[1:], stderr, stderr)
}

// ResolveCloneDir resolves the absolute local directory path under reposRoot for uri.
func ResolveCloneDir(reposRoot, uri string) (string, error) {
	if reposRoot == "" {
		return "", fmt.Errorf("--repos-root (-r) is required when URI is specified")
	}
	absRoot, err := filepath.Abs(reposRoot)
	if err != nil {
		return "", fmt.Errorf("resolve repos-root: %w", err)
	}
	relPath, err := URI2RelPath(uri)
	if err != nil {
		return "", err
	}
	return filepath.Join(absRoot, relPath), nil
}

// RemoteURL returns the remote URL for origin in dir, or an empty string if not found.
func (c *Command) RemoteURL(ctx context.Context, dir string) (string, error) {
	out, err := c.Output(ctx, dir, "remote", "get-url", "origin")
	if err != nil {
		return "", nil
	}
	return out, nil
}

// URI2RelPath converts a repository URI (HTTPS, HTTP, SSH, git://, SCP-like git@host:path,
// or host/path) into a relative directory path (<host>/<path> without trailing .git).
func URI2RelPath(uri string) (string, error) {
	raw, err := extractHostAndPath(strings.TrimSpace(uri))
	if err != nil {
		return "", err
	}

	raw = strings.TrimRight(raw, "/")
	raw = strings.TrimSuffix(raw, ".git")
	raw = strings.TrimRight(raw, "/")

	parts := strings.Split(raw, "/")
	if len(parts) < 2 ||
		strings.HasPrefix(parts[0], "-") ||
		slices.Contains(parts, "") ||
		slices.Contains(parts, ".") ||
		slices.Contains(parts, "..") {
		return "", fmt.Errorf("invalid repository URI: %q", uri)
	}

	return filepath.FromSlash(strings.Join(parts, "/")), nil
}

func extractHostAndPath(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("empty repository URI")
	}
	if strings.HasPrefix(s, "-") {
		return "", fmt.Errorf("invalid repository URI: %q", s)
	}

	if strings.Contains(s, "://") {
		u, err := url.Parse(s)
		if err != nil || u.Hostname() == "" {
			return "", fmt.Errorf("invalid repository URI: %q", s)
		}
		return u.Hostname() + "/" + strings.TrimPrefix(u.Path, "/"), nil
	}

	// Handle SCP-like syntax: [user@]host:path
	colon := strings.IndexByte(s, ':')
	slash := strings.IndexByte(s, '/')
	if colon > 0 && (slash < 0 || colon < slash) {
		host := s[:colon]
		if at := strings.IndexByte(host, '@'); at >= 0 {
			host = host[at+1:]
		}
		return host + "/" + strings.TrimPrefix(s[colon+1:], "/"), nil
	}

	// Handle scheme-less [user@]host/path
	if slash > 0 {
		host := s[:slash]
		if at := strings.IndexByte(host, '@'); at >= 0 {
			host = host[at+1:]
		}
		return host + "/" + s[slash+1:], nil
	}

	return "", fmt.Errorf("invalid repository URI: %q", s)
}

