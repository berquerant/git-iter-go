package runner

import (
	"fmt"
	"strings"
)

// TaskError annotates an execution error with the repository path,
// the executed command, and the process exit code.
type TaskError struct {
	Dir      string
	Command  []string
	ExitCode int
	Err      error
}

func (e *TaskError) Error() string {
	return fmt.Sprintf("failed in %s: %s (exit code %d): %v",
		e.Dir,
		strings.Join(e.Command, " "),
		e.ExitCode,
		e.Err,
	)
}

func (e *TaskError) Unwrap() error {
	return e.Err
}
