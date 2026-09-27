package common

import (
	"errors"
	"os/exec"
	"path/filepath"
)

// ExitCodeFrom extracts the OS exit code from err.
//   - nil             → 0
//   - *exec.ExitError → the process exit code
//   - any other error → 1
func ExitCodeFrom(err error) int {
	if err == nil {
		return 0
	}
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return exitErr.ExitCode()
	}
	return 1
}

// RelPath returns repoAbsPath relative to reposRoot.
// Falls back to repoAbsPath when reposRoot is empty or the computation fails.
func RelPath(reposRoot, repoAbsPath string) string {
	if reposRoot == "" {
		return repoAbsPath
	}
	rel, err := filepath.Rel(reposRoot, repoAbsPath)
	if err != nil {
		return repoAbsPath
	}
	return rel
}
