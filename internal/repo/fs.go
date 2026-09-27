package repo

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

// FilesystemFinder walks Root recursively and returns every directory that
// contains a ".git" entry (file or directory).
type FilesystemFinder struct {
	Root string
}

func (f *FilesystemFinder) Find(ctx context.Context) ([]string, error) {
	var repos []string
	err := filepath.WalkDir(f.Root, func(path string, d fs.DirEntry, err error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		// Check for .git inside this directory.
		if _, statErr := os.Stat(filepath.Join(path, ".git")); statErr == nil {
			repos = append(repos, path)
			// Don't descend into nested git repos.
			return filepath.SkipDir
		}
		return nil
	})
	return repos, err
}

// CommandFinder runs an external command and treats each line of its stdout as
// an absolute repository path.
// CommandFunc can be replaced in tests to avoid real subprocess execution.
type CommandFinder struct {
	// Command is a shell command string executed via "sh -c".
	Command string
	// CommandFunc overrides command execution when set (useful in tests).
	CommandFunc func(ctx context.Context, command string) ([]byte, error)
}

func (f *CommandFinder) runCommand(ctx context.Context) ([]byte, error) {
	if f.CommandFunc != nil {
		return f.CommandFunc(ctx, f.Command)
	}
	//nolint:gosec // intentional command execution by configuration
	return exec.CommandContext(ctx, "sh", "-c", f.Command).Output()
}

func (f *CommandFinder) Find(ctx context.Context) ([]string, error) {
	out, err := f.runCommand(ctx)
	if err != nil {
		return nil, err
	}
	return splitLines(ctx, string(out))
}

// NewFinder constructs the appropriate Finder based on config values.
// If listRepos is non-empty, a CommandFinder is returned.
// Otherwise a FilesystemFinder rooted at reposRoot is returned.
func NewFinder(reposRoot, listRepos string) Finder {
	if listRepos != "" {
		return &CommandFinder{Command: listRepos}
	}
	return &FilesystemFinder{Root: reposRoot}
}
