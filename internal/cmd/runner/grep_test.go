package runner_test

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestRunGrep(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		outMode      string
		paths        []string
		grepArgs     parse.GrepArgs
		runnerOutput string
		wantStdout   string
		validateJSON bool
	}{
		{
			name:         "text mode with path prefix",
			outMode:      "text",
			paths:        []string{"/repos/org/repo1"},
			grepArgs:     parse.GrepArgs{GitGrepArgs: []string{"TODO"}},
			runnerOutput: "main.go:1:TODO\n",
			wantStdout:   "org/repo1/main.go:1:TODO\n",
		},
		{
			name:         "json mode",
			outMode:      "json",
			paths:        []string{"/repos/org/repo1"},
			grepArgs:     parse.GrepArgs{GitGrepArgs: []string{"TODO"}},
			runnerOutput: "main.go:1:TODO\n",
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

			err := runner.RunGrep(context.Background(), cfg, tt.outMode, tt.grepArgs, finder, fakeRunner, &stdout, &stderr)
			require.NoError(t, err)

			if tt.validateJSON {
				line := strings.TrimSpace(stdout.String())
				var res output.Result
				require.NoError(t, json.Unmarshal([]byte(line), &res))
				assert.Equal(t, "/repos/org/repo1", res.RepoAbsPath)
				assert.Equal(t, "org/repo1/main.go:1:TODO\n", res.Stdout)
			} else {
				assert.Equal(t, tt.wantStdout, stdout.String())
			}
		})
	}
}

func TestRunGrep_CustomGitCommand(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		MaxProcs:   1,
		ReposRoot:  "/repos",
		GitCommand: "/custom/bin/git",
	}
	finder := &testutil.FakeFinder{Paths: []string{"/repos/org/repo1"}}
	fakeRunner := &testutil.FakeRunner{}
	var stdout, stderr bytes.Buffer

	err := runner.RunGrep(context.Background(), cfg, "text", parse.GrepArgs{GitGrepArgs: []string{"TODO"}}, finder, fakeRunner, &stdout, &stderr)
	require.NoError(t, err)

	require.Len(t, fakeRunner.Calls, 1)
	assert.Equal(t, []string{"/custom/bin/git", "grep", "-H", "TODO"}, fakeRunner.Calls[0].Command)
}

func TestRunGrep_ContextCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := &config.Config{
		MaxProcs:  1,
		ReposRoot: "/repos",
	}
	finder := &testutil.FakeFinder{Paths: []string{"/repos/a", "/repos/b"}}
	fakeRunner := &testutil.FakeRunner{}
	var stdout, stderr bytes.Buffer

	err := runner.RunGrep(ctx, cfg, "text", parse.GrepArgs{GitGrepArgs: []string{"TODO"}}, finder, fakeRunner, &stdout, &stderr)
	require.ErrorIs(t, err, context.Canceled)
}

func TestRunGrep_Markdown(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		MaxProcs:            1,
		ReposRoot:           "/repos",
		MarkdownHeaderLevel: 3,
	}
	finder := &testutil.FakeFinder{Paths: []string{"/repos/org/repo1"}}
	fakeRunner := &testutil.FakeRunner{Output: "foo.go:1:match\n"}
	var stdout, stderr bytes.Buffer

	err := runner.RunGrep(context.Background(), cfg, "md", parse.GrepArgs{GitGrepArgs: []string{"match"}}, finder, fakeRunner, &stdout, &stderr)
	require.NoError(t, err)

	out := stdout.String()
	assert.Contains(t, out, "## grep: root = /repos, command = git grep -H match\n\n")
	assert.Contains(t, out, "### org/repo1: exit code = 0\n\n")
	assert.Contains(t, out, "<details>\n<summary>stdout</summary>\n\n```\norg/repo1/foo.go:1:match\n```\n\n</details>")
	assert.Contains(t, out, "## Summary\n\n")
}
