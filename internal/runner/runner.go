// Package runner provides types for executing commands in a specific directory.
package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

// Runner executes a command in dir, writing its stdout/stderr to the provided writers.
type Runner interface {
	Run(ctx context.Context, dir string, command []string, stdout, stderr io.Writer) error
}

// ProcessRunner executes commands as real OS processes.
type ProcessRunner struct{}

func (r ProcessRunner) Run(ctx context.Context, dir string, command []string, stdout, stderr io.Writer) error {
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

// PathPrefixRunner wraps another Runner to execute a command and prefix each output
// line with the repository path.
//
// When AbsPath is true the prefix is the absolute RepoPath.
// When AbsPath is false the prefix is RepoPath relative to Root.
type PathPrefixRunner struct {
	Inner Runner
	// RepoPath is the absolute path of the repository.
	RepoPath string
	// AbsPath controls whether output paths are absolute or relative to Root.
	AbsPath bool
	// Root is the repos root used to compute a relative prefix when AbsPath is false.
	Root string
}

// GrepRunner is an alias for PathPrefixRunner.
type GrepRunner = PathPrefixRunner

// prefixPath returns the path prefix to prepend to each matched line.
func (g *PathPrefixRunner) prefixPath() (string, error) {
	if g.AbsPath {
		return g.RepoPath, nil
	}
	if g.Root == "" {
		return g.RepoPath, nil
	}
	rel, err := filepath.Rel(g.Root, g.RepoPath)
	if err != nil {
		return "", fmt.Errorf("runner: rel path: %w", err)
	}
	return rel, nil
}

func (g *PathPrefixRunner) Run(ctx context.Context, dir string, command []string, stdout, stderr io.Writer) error {
	prefix, err := g.prefixPath()
	if err != nil {
		return err
	}

	// Capture git grep output so we can prepend the repo path to every line.
	var buf bytes.Buffer
	if err := g.Inner.Run(ctx, dir, command, &buf, stderr); err != nil {
		return err
	}

	for line := range strings.SplitSeq(buf.String(), "\n") {
		if line == "" {
			continue
		}
		if _, err := fmt.Fprintf(stdout, "%s/%s\n", prefix, line); err != nil {
			return err
		}
	}
	return nil
}
