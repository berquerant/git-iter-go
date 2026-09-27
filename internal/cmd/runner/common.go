package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/executor"
	"github.com/berquerant/git-iter-go/internal/output"
	iterRunner "github.com/berquerant/git-iter-go/internal/runner"
	"golang.org/x/sync/semaphore"
)

const (
	outModeMarkdown = "markdown"
	outModeMD       = "md"
	outModeJSON     = "json"
)

func formatCommandTitle(subcmd, reposRoot string, command []string) string {
	var parts []string
	if reposRoot != "" {
		parts = append(parts, fmt.Sprintf("root = %s", reposRoot))
	}
	if len(command) > 0 {
		parts = append(parts, fmt.Sprintf("command = %s", strings.Join(command, " ")))
	}
	if len(parts) > 0 {
		return fmt.Sprintf("%s: %s", subcmd, strings.Join(parts, ", "))
	}
	return subcmd
}

func buildOutputTasks(paths []string, reposRoot string, command []string, r iterRunner.Runner) []output.Task {
	return buildOutputTasksWithRunner(paths, reposRoot, command, func(_ string) iterRunner.Runner {
		return r
	})
}

func buildOutputTasksWithRunner(
	paths []string,
	reposRoot string,
	command []string,
	runnerFunc func(path string) iterRunner.Runner,
) []output.Task {
	tasks := make([]output.Task, len(paths))
	for i, p := range paths {
		tasks[i] = output.Task{
			RepoAbsPath: p,
			ReposRoot:   reposRoot,
			Dir:         p,
			Command:     command,
			Runner:      runnerFunc(p),
		}
	}
	return tasks
}

func runCommandOnPaths(
	ctx context.Context,
	cfg *config.Config,
	outMode, subcmd string,
	paths, command []string,
	r iterRunner.Runner,
	stdout, stderr io.Writer,
) error {
	switch outMode {
	case outModeMarkdown, outModeMD:
		tasks := buildOutputTasks(paths, cfg.ReposRoot, command, r)
		title := formatCommandTitle(subcmd, cfg.ReposRoot, command)
		return (&output.MarkdownExecutor{
			Title:       title,
			MaxProcs:    cfg.MaxProcs,
			FailFast:    cfg.FailFast,
			Timeout:     cfg.Timeout,
			HeaderLevel: cfg.MarkdownHeaderLevel,
			Out:         stdout,
		}).Run(ctx, tasks)

	case outModeJSON:
		tasks := buildOutputTasks(paths, cfg.ReposRoot, command, r)
		return (&output.JSONLExecutor{
			MaxProcs: cfg.MaxProcs,
			FailFast: cfg.FailFast,
			Timeout:  cfg.Timeout,
			Out:      stdout,
		}).Run(ctx, tasks)

	default: // "text"
		tasks := make([]executor.Task, len(paths))
		for i, p := range paths {
			tasks[i] = executor.Task{Dir: p, Command: command}
		}
		err := (&executor.Executor{
			Runner:   r,
			MaxProcs: cfg.MaxProcs,
			FailFast: cfg.FailFast,
			Timeout:  cfg.Timeout,
			Stdout:   stdout,
			Stderr:   stderr,
		}).Run(ctx, tasks)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if cfg.FailFast {
			return err
		}
		return nil
	}
}

// runPathPrefixText is the common text-mode execution logic for grep and ls-files.
func runPathPrefixText(
	ctx context.Context,
	maxProcs int,
	failFast, absPath bool,
	reposRoot string,
	paths, cmd []string,
	innerRunner iterRunner.Runner,
	stdout, stderr io.Writer,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if maxProcs < 1 {
		maxProcs = 1
	}
	sem := semaphore.NewWeighted(int64(maxProcs))
	var (
		wg   sync.WaitGroup
		errs []error
		mu   sync.Mutex
	)
	for _, p := range paths {
		if err := sem.Acquire(ctx, 1); err != nil {
			break
		}
		wg.Go(func() {
			defer sem.Release(1)
			pr := &iterRunner.PathPrefixRunner{
				Inner:    innerRunner,
				RepoPath: p,
				AbsPath:  absPath,
				Root:     reposRoot,
			}
			ex := &executor.Executor{
				Runner:   pr,
				MaxProcs: 1,
				FailFast: failFast,
				Stdout:   stdout,
				Stderr:   stderr,
			}
			if err := ex.Run(ctx, []executor.Task{{Dir: p, Command: cmd}}); err != nil {
				// Process errors are collected; if failFast is true, cancel remaining.
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				if failFast {
					cancel()
				}
			}
		})
	}
	wg.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if failFast && len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
