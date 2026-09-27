package common

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
	"uuid"

	"github.com/berquerant/git-iter-go/internal/runner"
	"golang.org/x/sync/semaphore"
)

// ExecuteTask runs a single task, emitting logs and returning the Result and error.
func ExecuteTask(ctx context.Context, task Task, timeout time.Duration, failFast bool) (Result, error) {
	taskCtx := ctx
	if timeout > 0 {
		var taskCancel context.CancelFunc
		taskCtx, taskCancel = context.WithTimeout(ctx, timeout)
		defer taskCancel()
	}

	r := task.Runner
	if r == nil {
		r = runner.ProcessRunner{}
	}

	taskID := uuid.New().String()
	slog.DebugContext(taskCtx, "exec start", "id", taskID, "dir", task.Dir, "command", task.Command)
	var stdout, stderr bytes.Buffer
	runErr := r.Run(taskCtx, task.Dir, task.Command, &stdout, &stderr)
	exitCode := ExitCodeFrom(runErr)

	var taskErr error
	if runErr != nil {
		taskErr = &runner.TaskError{
			Dir:      task.Dir,
			Command:  task.Command,
			ExitCode: exitCode,
			Err:      runErr,
		}
		if failFast {
			slog.ErrorContext(ctx, "exec failed (aborting)",
				"id", taskID, "dir", task.Dir, "command", task.Command,
				"exit_code", exitCode, "err", taskErr)
		} else {
			slog.WarnContext(ctx, "exec failed",
				"id", taskID, "dir", task.Dir, "command", task.Command,
				"exit_code", exitCode, "err", taskErr)
		}
	} else {
		slog.DebugContext(ctx, "exec success", "id", taskID, "dir", task.Dir, "exit_code", 0)
	}

	res := Result{
		RepoAbsPath: task.RepoAbsPath,
		RepoRelPath: RelPath(task.ReposRoot, task.RepoAbsPath),
		Command:     task.Command,
		ExitCode:    exitCode,
		Stdout:      stdout.String(),
		Stderr:      stderr.String(),
	}
	return res, taskErr
}

// ExecuteTasksCore coordinates parallel task execution with concurrency limit, timeout, and fail-fast.
func ExecuteTasksCore(
	ctx context.Context,
	tasks []Task,
	maxProcs int,
	timeout time.Duration,
	failFast bool,
	onResult func(Result) error,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sem := semaphore.NewWeighted(int64(max(maxProcs, 1)))
	var (
		mu   sync.Mutex
		wg   sync.WaitGroup
		errs []error
	)

	for _, task := range tasks {
		if err := sem.Acquire(ctx, 1); err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			break
		}
		wg.Go(func() {
			defer sem.Release(1)

			res, taskErr := ExecuteTask(ctx, task, timeout, failFast)
			if taskErr != nil {
				if failFast {
					mu.Lock()
					errs = append(errs, taskErr)
					mu.Unlock()
					cancel()
				}
			}

			mu.Lock()
			writeErr := onResult(res)
			if writeErr != nil {
				errs = append(errs, writeErr)
				if failFast {
					cancel()
				}
			}
			mu.Unlock()
		})
	}

	wg.Wait()
	return errors.Join(errs...)
}
