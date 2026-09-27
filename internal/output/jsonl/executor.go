package jsonl

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/berquerant/git-iter-go/internal/output/common"
)

// JSONLExecutor runs Tasks concurrently (bounded by MaxProcs), captures each
// task's stdout and stderr, and writes one JSONL Result line to Out per task.
//
// Non-zero exit codes from the executed process are recorded in Result.ExitCode
// and do NOT cause Run to return an error — they are treated as data.
// Run returns an error when context is cancelled, writing to Out fails,
// or when FailFast is true and a task fails.
type JSONLExecutor struct {
	MaxProcs int
	FailFast bool
	Timeout  time.Duration
	Out      io.Writer
}

// Run executes all tasks and blocks until they complete.
// When FailFast is true, the first error cancels context and halts remaining tasks.
func (e *JSONLExecutor) Run(ctx context.Context, tasks []common.Task) error {
	return common.ExecuteTasksCore(ctx, tasks, e.MaxProcs, e.Timeout, e.FailFast, func(r common.Result) error {
		return WriteJSONL(e.Out, r)
	})
}

// FileJSONLExecutor runs Tasks concurrently (bounded by MaxProcs), parses each
// line of stdout as a relative file path, and writes one JSONL FileResult line per file.
type FileJSONLExecutor struct {
	MaxProcs int
	FailFast bool
	Timeout  time.Duration
	Out      io.Writer
}

// Run executes all tasks and emits a FileResult per output line.
func (e *FileJSONLExecutor) Run(ctx context.Context, tasks []common.Task) error {
	return common.ExecuteTasksCore(ctx, tasks, e.MaxProcs, e.Timeout, e.FailFast, func(r common.Result) error {
		for line := range strings.SplitSeq(r.Stdout, "\n") {
			if line == "" {
				continue
			}
			fileRelPath := line
			fileAbsPath := filepath.Join(r.RepoAbsPath, fileRelPath)
			fr := common.FileResult{
				RepoAbsPath: r.RepoAbsPath,
				RepoRelPath: r.RepoRelPath,
				FileRelPath: fileRelPath,
				FileAbsPath: fileAbsPath,
			}
			if err := WriteFileJSONL(e.Out, fr); err != nil {
				return err
			}
		}
		return nil
	})
}
