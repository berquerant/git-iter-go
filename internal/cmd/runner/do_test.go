package runner_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/berquerant/git-iter-go/internal/cmd/runner"
	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/output"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunDo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		outMode      string
		paths        []string
		doArgs       parse.DoArgs
		runnerOutput string
		wantStdout   string
		validateJSON bool
	}{
		{
			name:         "text mode streaming",
			outMode:      "text",
			paths:        []string{"/repos/repo1", "/repos/repo2"},
			doArgs:       parse.DoArgs{Command: []string{"git", "status"}},
			runnerOutput: "clean\n",
			wantStdout:   "clean\nclean\n",
		},
		{
			name:         "json mode",
			outMode:      "json",
			paths:        []string{"/repos/org/repo1"},
			doArgs:       parse.DoArgs{Command: []string{"echo", "hi"}},
			runnerOutput: "hi\n",
			validateJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &config.Config{
				MaxProcs:  1,
				ReposRoot: "/repos",
			}
			finder := &testutil.FakeFinder{Paths: tt.paths}
			fakeRunner := &testutil.FakeRunner{Output: tt.runnerOutput}
			var stdout, stderr bytes.Buffer

			err := runner.RunDo(context.Background(), cfg, tt.outMode, tt.doArgs, finder, fakeRunner, &stdout, &stderr)
			require.NoError(t, err)

			if tt.validateJSON {
				line := strings.TrimSpace(stdout.String())
				var res output.Result
				require.NoError(t, json.Unmarshal([]byte(line), &res))
				assert.Equal(t, "/repos/org/repo1", res.RepoAbsPath)
				assert.Equal(t, "org/repo1", res.RepoRelPath)
				assert.Equal(t, tt.doArgs.Command, res.Command)
				assert.Equal(t, tt.runnerOutput, res.Stdout)
				assert.Equal(t, 0, res.ExitCode)
			} else {
				assert.Equal(t, tt.wantStdout, stdout.String())
			}
		})
	}
}

func TestRunRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		outMode      string
		paths        []string
		readArgs     parse.ReadArgs
		runnerOutput string
		wantStdout   string
		validateJSON bool
	}{
		{
			name:         "text mode",
			outMode:      "text",
			paths:        []string{"/repos/r1"},
			readArgs:     parse.ReadArgs{Command: []string{"git", "status"}},
			runnerOutput: "ok\n",
			wantStdout:   "ok\n",
		},
		{
			name:         "json mode",
			outMode:      "json",
			paths:        []string{"/repos/r1"},
			readArgs:     parse.ReadArgs{Command: []string{"git", "status"}},
			runnerOutput: "ok\n",
			validateJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &config.Config{
				MaxProcs:  1,
				ReposRoot: "/repos",
			}
			finder := &testutil.FakeFinder{Paths: tt.paths}
			fakeRunner := &testutil.FakeRunner{Output: tt.runnerOutput}
			var stdout, stderr bytes.Buffer

			err := runner.RunRead(context.Background(), cfg, tt.outMode, tt.readArgs, finder, fakeRunner, &stdout, &stderr)
			require.NoError(t, err)

			if tt.validateJSON {
				line := strings.TrimSpace(stdout.String())
				var res output.Result
				require.NoError(t, json.Unmarshal([]byte(line), &res))
				assert.Equal(t, "/repos/r1", res.RepoAbsPath)
				assert.Equal(t, tt.readArgs.Command, res.Command)
				assert.Equal(t, tt.runnerOutput, res.Stdout)
			} else {
				assert.Equal(t, tt.wantStdout, stdout.String())
			}
		})
	}
}

func TestRunDo_FailFast(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		MaxProcs: 1,
		FailFast: true,
	}
	finder := &testutil.FakeFinder{Paths: []string{"/a", "/b", "/c"}}
	fakeRunner := &testutil.FakeRunner{Err: errors.New("fail")}
	var stdout, stderr bytes.Buffer

	err := runner.RunDo(context.Background(), cfg, "text", parse.DoArgs{Command: []string{"echo"}}, finder, fakeRunner, &stdout, &stderr)
	require.Error(t, err)
	assert.Less(t, len(fakeRunner.Calls), 3)
}

func TestRunDo_DefaultIgnoresErrors(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		MaxProcs: 1,
		FailFast: false,
	}
	finder := &testutil.FakeFinder{Paths: []string{"/a", "/b", "/c"}}
	fakeRunner := &testutil.FakeRunner{Err: errors.New("fail")}
	var stdout, stderr bytes.Buffer

	err := runner.RunDo(context.Background(), cfg, "text", parse.DoArgs{Command: []string{"echo"}}, finder, fakeRunner, &stdout, &stderr)
	require.NoError(t, err)
	assert.Len(t, fakeRunner.Calls, 3)
}

func TestRunDo_Markdown(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		MaxProcs:            1,
		ReposRoot:           "/repos",
		MarkdownHeaderLevel: 3,
	}
	finder := &testutil.FakeFinder{Paths: []string{"/repos/org/repo1"}}
	fakeRunner := &testutil.FakeRunner{Output: "output text\n"}
	var stdout, stderr bytes.Buffer

	err := runner.RunDo(context.Background(), cfg, "markdown", parse.DoArgs{Command: []string{"echo", "hi"}}, finder, fakeRunner, &stdout, &stderr)
	require.NoError(t, err)

	out := stdout.String()
	assert.Contains(t, out, "## do: root = /repos, command = echo hi\n\n")
	assert.Contains(t, out, "### org/repo1: exit code = 0\n\n")
	assert.Contains(t, out, "<details>\n<summary>stdout</summary>\n\n```\noutput text\n```\n\n</details>")
	assert.Contains(t, out, "## Summary\n\n")
	assert.Contains(t, out, "- **Total Repositories**: 1\n")
}
