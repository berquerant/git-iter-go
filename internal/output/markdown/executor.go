package markdown

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/berquerant/git-iter-go/internal/markdown"
	"github.com/berquerant/git-iter-go/internal/output/common"
)

// MarkdownExecutor runs Tasks concurrently (bounded by MaxProcs), captures each
// task's stdout and stderr, writes each result formatted as Markdown to Out,
// and writes an aggregate summary at the end.
type MarkdownExecutor struct {
	Title       string
	MaxProcs    int
	FailFast    bool
	Timeout     time.Duration
	HeaderLevel int
	Out         io.Writer
}

// Run executes all tasks and blocks until they complete.
func (e *MarkdownExecutor) Run(ctx context.Context, tasks []common.Task) error {
	if e.Title != "" {
		if err := markdown.WriteHeader(e.Out, e.Title, e.HeaderLevel); err != nil {
			return err
		}
	}

	var (
		resMu   sync.Mutex
		results []common.Result
	)
	err := common.ExecuteTasksCore(ctx, tasks, e.MaxProcs, e.Timeout, e.FailFast, func(r common.Result) error {
		writeErr := markdown.WriteResult(e.Out, r, e.HeaderLevel)
		resMu.Lock()
		results = append(results, r)
		resMu.Unlock()
		return writeErr
	})

	if len(results) > 0 {
		_ = markdown.WriteSummary(e.Out, results, e.HeaderLevel)
	}
	return err
}

// LsFilesMarkdownExecutor runs Tasks concurrently (bounded by MaxProcs),
// writes each repository's file list formatted as Markdown to Out,
// and writes an aggregate summary at the end.
type LsFilesMarkdownExecutor struct {
	Title       string
	MaxProcs    int
	FailFast    bool
	Timeout     time.Duration
	HeaderLevel int
	Out         io.Writer
}

// Run executes all tasks and writes Markdown output for ls-files.
func (e *LsFilesMarkdownExecutor) Run(ctx context.Context, tasks []common.Task) error {
	if e.Title != "" {
		if err := markdown.WriteHeader(e.Out, e.Title, e.HeaderLevel); err != nil {
			return err
		}
	}

	var (
		resMu   sync.Mutex
		results []common.Result
	)
	err := common.ExecuteTasksCore(ctx, tasks, e.MaxProcs, e.Timeout, e.FailFast, func(r common.Result) error {
		writeErr := markdown.WriteLsFiles(e.Out, r, e.HeaderLevel)
		resMu.Lock()
		results = append(results, r)
		resMu.Unlock()
		return writeErr
	})

	if len(results) > 0 {
		_ = markdown.WriteLsFilesSummary(e.Out, results, e.HeaderLevel)
	}
	return err
}
