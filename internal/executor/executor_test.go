package executor_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/berquerant/git-iter-go/internal/executor"
	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTasks(dirs ...string) []executor.Task {
	tasks := make([]executor.Task, len(dirs))
	for i, d := range dirs {
		tasks[i] = executor.Task{Dir: d, Command: []string{"echo", d}}
	}
	return tasks
}

func TestExecutor_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		maxProcs  int
		tasks     []executor.Task
		runnerErr error
		wantErr   bool
		wantCalls int
	}{
		{
			name:      "no tasks",
			maxProcs:  1,
			tasks:     nil,
			wantCalls: 0,
		},
		{
			name:      "single task maxProcs 1",
			maxProcs:  1,
			tasks:     makeTasks("/repo"),
			wantCalls: 1,
		},
		{
			name:      "three tasks maxProcs 1",
			maxProcs:  1,
			tasks:     makeTasks("/a", "/b", "/c"),
			wantCalls: 3,
		},
		{
			name:      "three tasks maxProcs 3",
			maxProcs:  3,
			tasks:     makeTasks("/a", "/b", "/c"),
			wantCalls: 3,
		},
		{
			name:      "runner error is returned",
			maxProcs:  1,
			tasks:     makeTasks("/repo"),
			runnerErr: assert.AnError,
			wantErr:   true,
			wantCalls: 1,
		},
		{
			name:      "maxProcs 0 treated as 1",
			maxProcs:  0,
			tasks:     makeTasks("/repo"),
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fake := &testutil.FakeRunner{Err: tt.runnerErr}
			ex := &executor.Executor{
				Runner:   fake,
				MaxProcs: tt.maxProcs,
				Stdout:   &bytes.Buffer{},
				Stderr:   &bytes.Buffer{},
			}
			err := ex.Run(context.Background(), tt.tasks)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Len(t, fake.Calls, tt.wantCalls)
		})
	}
}

func TestExecutor_ContextCancellation(t *testing.T) {
	t.Parallel()

	var started atomic.Int64
	blockingRunner := &blockRunner{started: &started}

	ctx, cancel := context.WithCancel(context.Background())

	tasks := makeTasks("/a", "/b", "/c", "/d", "/e")
	ex := &executor.Executor{
		Runner:   blockingRunner,
		MaxProcs: 1,
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}

	done := make(chan error, 1)
	go func() {
		done <- ex.Run(ctx, tasks)
	}()

	require.Eventually(t, func() bool { return started.Load() > 0 }, time.Second, 10*time.Millisecond)
	cancel()

	err := <-done
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

// blockRunner implements runner.Runner; blocks until ctx is cancelled.
type blockRunner struct {
	started *atomic.Int64
}

func (r *blockRunner) Run(ctx context.Context, _ string, _ []string, _, _ io.Writer) error {
	r.started.Add(1)
	<-ctx.Done()
	return nil
}

// timeoutRunner implements runner.Runner; blocks until ctx is cancelled.
type timeoutRunner struct{}

func (timeoutRunner) Run(ctx context.Context, _ string, _ []string, _, _ io.Writer) error {
	<-ctx.Done()
	return ctx.Err()
}

func TestExecutor_Control(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		tasks        []executor.Task
		runner       runner.Runner
		timeout      time.Duration
		failFast     bool
		wantErr      error
		wantMaxCalls int // max allowed calls to Runner
	}{
		{
			name:         "fail fast aborts remaining tasks",
			tasks:        makeTasks("/a", "/b", "/c", "/d"),
			runner:       &testutil.FakeRunner{Err: errors.New("boom")},
			failFast:     true,
			wantErr:      errors.New("boom"),
			wantMaxCalls: 3, // out of 4 tasks, at most 3 (typically 1)
		},
		{
			name:         "timeout terminates hanging task with deadline exceeded",
			tasks:        makeTasks("/repo1"),
			runner:       timeoutRunner{},
			timeout:      10 * time.Millisecond,
			wantErr:      context.DeadlineExceeded,
			wantMaxCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ex := &executor.Executor{
				Runner:   tt.runner,
				MaxProcs: 1,
				FailFast: tt.failFast,
				Timeout:  tt.timeout,
				Stdout:   &bytes.Buffer{},
				Stderr:   &bytes.Buffer{},
			}
			err := ex.Run(context.Background(), tt.tasks)
			require.Error(t, err)
			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr) || strings.Contains(err.Error(), tt.wantErr.Error()))
			}
			if fake, ok := tt.runner.(*testutil.FakeRunner); ok {
				assert.LessOrEqual(t, len(fake.Calls), tt.wantMaxCalls)
			}
		})
	}
}
