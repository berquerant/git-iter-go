package jsonl_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/berquerant/git-iter-go/internal/output/common"
	"github.com/berquerant/git-iter-go/internal/output/jsonl"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseJSONLResults(t *testing.T, s string) []common.Result {
	t.Helper()
	var results []common.Result
	for line := range strings.SplitSeq(strings.TrimSpace(s), "\n") {
		if line == "" {
			continue
		}
		var r common.Result
		require.NoError(t, json.Unmarshal([]byte(line), &r))
		results = append(results, r)
	}
	return results
}

func TestWriteJSONL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		r    common.Result
	}{
		{
			name: "zero value",
			r:    common.Result{},
		},
		{
			name: "full result",
			r: common.Result{
				RepoAbsPath: "/repos/myorg/myrepo",
				RepoRelPath: "myorg/myrepo",
				Command:     []string{"git", "status"},
				ExitCode:    0,
				Stdout:      "On branch main\n",
				Stderr:      "",
			},
		},
		{
			name: "non-zero exit",
			r: common.Result{
				RepoAbsPath: "/repos/x",
				RepoRelPath: "x",
				Command:     []string{"false"},
				ExitCode:    1,
				Stdout:      "",
				Stderr:      "error text",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			require.NoError(t, jsonl.WriteJSONL(&buf, tt.r))

			line := buf.String()
			// Must end with exactly one newline.
			assert.Equal(t, byte('\n'), line[len(line)-1])

			// Must be valid JSON that round-trips correctly.
			var got common.Result
			require.NoError(t, json.Unmarshal([]byte(line), &got))
			assert.Equal(t, tt.r, got)
		})
	}
}

func TestWriteFileJSONL(t *testing.T) {
	t.Parallel()
	r := common.FileResult{
		RepoAbsPath: "/repos/org/repo1",
		RepoRelPath: "org/repo1",
		FileRelPath: "cmd/main.go",
		FileAbsPath: "/repos/org/repo1/cmd/main.go",
	}

	var buf bytes.Buffer
	require.NoError(t, jsonl.WriteFileJSONL(&buf, r))

	line := buf.String()
	assert.Equal(t, byte('\n'), line[len(line)-1])

	var got common.FileResult
	require.NoError(t, json.Unmarshal([]byte(line), &got))
	assert.Equal(t, r, got)
}

func TestJSONLExecutor_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		maxProcs  int
		tasks     []common.Task
		wantLines int
		// wantExitCode is the expected exit_code for all results (when uniform).
		wantExitCode *int
	}{
		{
			name:      "no tasks produces no output",
			maxProcs:  1,
			tasks:     nil,
			wantLines: 0,
		},
		{
			name:     "single task success",
			maxProcs: 1,
			tasks: []common.Task{
				{RepoAbsPath: "/repos/a", ReposRoot: "/repos", Dir: "/repos/a",
					Command: []string{"echo"}, Runner: &testutil.FakeRunner{Output: "hello\n"}},
			},
			wantLines:    1,
			wantExitCode: new(0),
		},
		{
			name:     "three tasks maxProcs 1",
			maxProcs: 1,
			tasks: []common.Task{
				{RepoAbsPath: "/repos/a", ReposRoot: "/repos", Dir: "/repos/a", Command: []string{"x"}, Runner: &testutil.FakeRunner{}},
				{RepoAbsPath: "/repos/b", ReposRoot: "/repos", Dir: "/repos/b", Command: []string{"x"}, Runner: &testutil.FakeRunner{}},
				{RepoAbsPath: "/repos/c", ReposRoot: "/repos", Dir: "/repos/c", Command: []string{"x"}, Runner: &testutil.FakeRunner{}},
			},
			wantLines:    3,
			wantExitCode: new(0),
		},
		{
			name:     "three tasks maxProcs 3",
			maxProcs: 3,
			tasks: []common.Task{
				{RepoAbsPath: "/repos/a", Dir: "/repos/a", Command: []string{"x"}, Runner: &testutil.FakeRunner{}},
				{RepoAbsPath: "/repos/b", Dir: "/repos/b", Command: []string{"x"}, Runner: &testutil.FakeRunner{}},
				{RepoAbsPath: "/repos/c", Dir: "/repos/c", Command: []string{"x"}, Runner: &testutil.FakeRunner{}},
			},
			wantLines:    3,
			wantExitCode: new(0),
		},
		{
			name:     "runner error captured in exit_code not propagated",
			maxProcs: 1,
			tasks: []common.Task{
				{RepoAbsPath: "/repos/a", Dir: "/repos/a", Command: []string{"x"},
					Runner: &testutil.FakeRunner{Err: assert.AnError}},
			},
			wantLines:    1,
			wantExitCode: new(1), // generic error → exit code 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			ex := &jsonl.JSONLExecutor{MaxProcs: tt.maxProcs, Out: &buf}
			err := ex.Run(context.Background(), tt.tasks)
			require.NoError(t, err)

			results := parseJSONLResults(t, buf.String())
			assert.Len(t, results, tt.wantLines)

			if tt.wantExitCode != nil {
				for _, r := range results {
					assert.Equal(t, *tt.wantExitCode, r.ExitCode)
				}
			}
		})
	}
}

func TestJSONLExecutor_FieldsPopulated(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	task := common.Task{
		RepoAbsPath: "/repos/org/myrepo",
		ReposRoot:   "/repos",
		Dir:         "/repos/org/myrepo",
		Command:     []string{"git", "status"},
		Runner:      &testutil.FakeRunner{Output: "stdout content"},
	}
	ex := &jsonl.JSONLExecutor{MaxProcs: 1, Out: &buf}
	require.NoError(t, ex.Run(context.Background(), []common.Task{task}))

	results := parseJSONLResults(t, buf.String())
	require.Len(t, results, 1)
	r := results[0]
	assert.Equal(t, "/repos/org/myrepo", r.RepoAbsPath)
	assert.Equal(t, "org/myrepo", r.RepoRelPath)
	assert.Equal(t, []string{"git", "status"}, r.Command)
	assert.Equal(t, 0, r.ExitCode)
	assert.Equal(t, "stdout content", r.Stdout)
	assert.Equal(t, "", r.Stderr)
}

func TestJSONLExecutor_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	tasks := make([]common.Task, 5)
	for i := range tasks {
		tasks[i] = common.Task{
			RepoAbsPath: "/r",
			Dir:         "/r",
			Command:     []string{"x"},
			Runner:      &testutil.FakeRunner{},
		}
	}

	var buf bytes.Buffer
	ex := &jsonl.JSONLExecutor{MaxProcs: 1, Out: &buf}
	err := ex.Run(ctx, tasks)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

// blockingRunner blocks until ctx is done; used in liveness tests.
type blockingRunner struct{}

func (blockingRunner) Run(ctx context.Context, _ string, _ []string, _, _ io.Writer) error {
	<-ctx.Done()
	return ctx.Err()
}

func TestJSONLExecutor_Control(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		tasks         []common.Task
		timeout       time.Duration
		failFast      bool
		wantErr       bool
		wantResultLen int
		wantExitCode  int
	}{
		{
			name: "fail fast stops after first failure",
			tasks: []common.Task{
				{RepoAbsPath: "/a", Dir: "/a", Command: []string{"x"}, Runner: &testutil.FakeRunner{Err: errors.New("boom")}},
				{RepoAbsPath: "/b", Dir: "/b", Command: []string{"x"}, Runner: &testutil.FakeRunner{Err: errors.New("boom")}},
				{RepoAbsPath: "/c", Dir: "/c", Command: []string{"x"}, Runner: &testutil.FakeRunner{Err: errors.New("boom")}},
			},
			failFast:      true,
			wantErr:       true,
			wantResultLen: 1,
			wantExitCode:  1,
		},
		{
			name: "timeout terminates hanging task and records exit code 1",
			tasks: []common.Task{
				{
					RepoAbsPath: "/repo1",
					Dir:         "/repo1",
					Command:     []string{"sleep"},
					Runner:      blockingRunner{},
				},
			},
			timeout:       10 * time.Millisecond,
			wantErr:       false, // JSONLExecutor records runner error in Result.ExitCode
			wantResultLen: 1,
			wantExitCode:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			ex := &jsonl.JSONLExecutor{
				MaxProcs: 1,
				FailFast: tt.failFast,
				Timeout:  tt.timeout,
				Out:      &buf,
			}

			start := time.Now()
			err := ex.Run(context.Background(), tt.tasks)
			elapsed := time.Since(start)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.timeout > 0 {
				assert.Less(t, elapsed, 1*time.Second)
			}

			results := parseJSONLResults(t, buf.String())
			assert.Equal(t, tt.wantResultLen, len(results))
			if len(results) > 0 {
				assert.Equal(t, tt.wantExitCode, results[0].ExitCode)
			}
		})
	}
}

func TestFileJSONLExecutor_Run(t *testing.T) {
	t.Parallel()

	tasks := []common.Task{
		{
			RepoAbsPath: "/repos/org/repo1",
			ReposRoot:   "/repos",
			Dir:         "/repos/org/repo1",
			Command:     []string{"git", "ls-files"},
			Runner:      &testutil.FakeRunner{Output: "main.go\ncmd/root.go\n"},
		},
		{
			RepoAbsPath: "/repos/org/repo2",
			ReposRoot:   "/repos",
			Dir:         "/repos/org/repo2",
			Command:     []string{"git", "ls-files"},
			Runner:      &testutil.FakeRunner{Output: "README.md\n"},
		},
	}

	var buf bytes.Buffer
	ex := &jsonl.FileJSONLExecutor{
		MaxProcs: 1,
		Out:      &buf,
	}

	err := ex.Run(context.Background(), tasks)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 3)

	var res0 common.FileResult
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &res0))
	assert.Equal(t, "/repos/org/repo1", res0.RepoAbsPath)
	assert.Equal(t, "org/repo1", res0.RepoRelPath)
	assert.Equal(t, "main.go", res0.FileRelPath)
	assert.Equal(t, "/repos/org/repo1/main.go", res0.FileAbsPath)

	var res1 common.FileResult
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &res1))
	assert.Equal(t, "/repos/org/repo1", res1.RepoAbsPath)
	assert.Equal(t, "org/repo1", res1.RepoRelPath)
	assert.Equal(t, "cmd/root.go", res1.FileRelPath)
	assert.Equal(t, "/repos/org/repo1/cmd/root.go", res1.FileAbsPath)

	var res2 common.FileResult
	require.NoError(t, json.Unmarshal([]byte(lines[2]), &res2))
	assert.Equal(t, "/repos/org/repo2", res2.RepoAbsPath)
	assert.Equal(t, "org/repo2", res2.RepoRelPath)
	assert.Equal(t, "README.md", res2.FileRelPath)
	assert.Equal(t, "/repos/org/repo2/README.md", res2.FileAbsPath)
}
