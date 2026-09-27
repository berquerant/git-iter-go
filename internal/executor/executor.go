// Package executor provides concurrent task execution over a set of repositories.
package executor

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"
	"uuid"

	"github.com/berquerant/git-iter-go/internal/output"
	"github.com/berquerant/git-iter-go/internal/runner"
	"golang.org/x/sync/semaphore"
)

// Task represents a single unit of work: run Command in Dir.
type Task struct {
	Dir     string
	Command []string
}

// Executor runs Tasks concurrently, limited to MaxProcs parallel goroutines.
// All task stdout/stderr are written to Stdout/Stderr respectively.
type Executor struct {
	Runner   runner.Runner
	MaxProcs int
	FailFast bool
	Timeout  time.Duration
	Stdout   io.Writer
	Stderr   io.Writer
}

// Run executes all tasks and waits for them to complete.
// When FailFast is true, the first error cancels the context and halts remaining tasks.
// All errors are collected and returned as a joined error.
func (e *Executor) Run(ctx context.Context, tasks []Task) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	maxProcs := max(e.MaxProcs, 1)
	sem := semaphore.NewWeighted(int64(maxProcs))

	var (
		mu   sync.Mutex
		errs []error
		wg   sync.WaitGroup
	)

	for _, task := range tasks {
		if err := sem.Acquire(ctx, 1); err != nil {
			// Context cancelled; stop scheduling new tasks.
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			break
		}
		wg.Go(func() {
			defer sem.Release(1)

			taskCtx := ctx
			if e.Timeout > 0 {
				var taskCancel context.CancelFunc
				taskCtx, taskCancel = context.WithTimeout(ctx, e.Timeout)
				defer taskCancel()
			}

			taskID := uuid.New().String()
			slog.DebugContext(taskCtx, "exec start", "id", taskID, "dir", task.Dir, "command", task.Command)
			runErr := e.Runner.Run(taskCtx, task.Dir, task.Command, e.Stdout, e.Stderr)
			exitCode := output.ExitCodeFrom(runErr)

			if runErr != nil {
				taskErr := &runner.TaskError{
					Dir:      task.Dir,
					Command:  task.Command,
					ExitCode: exitCode,
					Err:      runErr,
				}
				if e.FailFast {
					slog.ErrorContext(ctx, "exec failed (aborting)",
						"id", taskID, "dir", task.Dir, "command", task.Command,
						"exit_code", exitCode, "err", taskErr)
				} else {
					slog.WarnContext(ctx, "exec failed",
						"id", taskID, "dir", task.Dir, "command", task.Command,
						"exit_code", exitCode, "err", taskErr)
				}

				mu.Lock()
				errs = append(errs, taskErr)
				mu.Unlock()
				if e.FailFast {
					cancel()
				}
			} else {
				slog.DebugContext(ctx, "exec success", "id", taskID, "dir", task.Dir, "exit_code", 0)
			}
		})
	}

	wg.Wait()
	return errors.Join(errs...)
}
