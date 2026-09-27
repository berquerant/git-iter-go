package runner_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/berquerant/git-iter-go/internal/cmd/runner"
	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/output"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunPull(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		outMode      string
		paths        []string
		pullArgs     parse.PullArgs
		runnerOutput string
		wantStdout   string
		validateJSON bool
	}{
		{
			name:         "text mode streaming",
			outMode:      "text",
			paths:        []string{"/repos/repo1", "/repos/repo2"},
			pullArgs:     parse.PullArgs{PullFlags: []string{"--rebase"}},
			runnerOutput: "Already up to date.\n",
			wantStdout:   "Already up to date.\nAlready up to date.\n",
		},
		{
			name:         "json mode",
			outMode:      "json",
			paths:        []string{"/repos/repo1"},
			pullArgs:     parse.PullArgs{PullFlags: []string{"--ff-only"}},
			runnerOutput: "Fast-forward\n",
			validateJSON: true,
		},
		{
			name:         "markdown mode",
			outMode:      "markdown",
			paths:        []string{"/repos/repo1"},
			pullArgs:     parse.PullArgs{PullFlags: []string{"--ff-only"}},
			runnerOutput: "Fast-forward\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				MaxProcs:            1,
				ReposRoot:           "/repos",
				MarkdownHeaderLevel: 3,
			}
			finder := &testutil.FakeFinder{Paths: tt.paths}
			fakeRunner := &testutil.FakeRunner{Output: tt.runnerOutput}
			var stdout, stderr bytes.Buffer

			err := runner.RunPull(
				context.Background(),
				cfg,
				tt.outMode,
				tt.pullArgs,
				finder,
				fakeRunner,
				&stdout,
				&stderr,
			)
			require.NoError(t, err)

			if tt.wantStdout != "" {
				assert.Equal(t, tt.wantStdout, stdout.String())
			}

			if tt.validateJSON {
				dec := json.NewDecoder(&stdout)
				var res output.Result
				require.NoError(t, dec.Decode(&res))
				assert.Equal(t, tt.paths[0], res.RepoAbsPath)
				assert.Equal(t, []string{"git", "pull", "--ff-only"}, res.Command)
				assert.Equal(t, tt.runnerOutput, res.Stdout)
			}

			if tt.outMode == "markdown" {
				out := stdout.String()
				assert.Contains(t, out, "## pull: root = /repos, command = git pull --ff-only\n\n")
				assert.Contains(t, out, "### repo1: exit code = 0\n\n")
				assert.Contains(t, out, "<details>\n<summary>stdout</summary>\n\n```\nFast-forward\n```\n\n</details>")
				assert.Contains(t, out, "## Summary\n\n")
			}
		})
	}
}

func TestRunFetch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		outMode      string
		paths        []string
		fetchArgs    parse.FetchArgs
		runnerOutput string
		wantStdout   string
		validateJSON bool
	}{
		{
			name:         "text mode streaming",
			outMode:      "text",
			paths:        []string{"/repos/repo1", "/repos/repo2"},
			fetchArgs:    parse.FetchArgs{FetchFlags: []string{"--prune"}},
			runnerOutput: "Fetching origin\n",
			wantStdout:   "Fetching origin\nFetching origin\n",
		},
		{
			name:         "json mode",
			outMode:      "json",
			paths:        []string{"/repos/repo1"},
			fetchArgs:    parse.FetchArgs{FetchFlags: []string{"--all"}},
			runnerOutput: "Fetching origin\n",
			validateJSON: true,
		},
		{
			name:         "markdown mode",
			outMode:      "markdown",
			paths:        []string{"/repos/repo1"},
			fetchArgs:    parse.FetchArgs{FetchFlags: []string{"--prune"}},
			runnerOutput: "Fetching origin\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				MaxProcs:            1,
				ReposRoot:           "/repos",
				MarkdownHeaderLevel: 3,
			}
			finder := &testutil.FakeFinder{Paths: tt.paths}
			fakeRunner := &testutil.FakeRunner{Output: tt.runnerOutput}
			var stdout, stderr bytes.Buffer

			err := runner.RunFetch(
				context.Background(),
				cfg,
				tt.outMode,
				tt.fetchArgs,
				finder,
				fakeRunner,
				&stdout,
				&stderr,
			)
			require.NoError(t, err)

			if tt.wantStdout != "" {
				assert.Equal(t, tt.wantStdout, stdout.String())
			}

			if tt.validateJSON {
				dec := json.NewDecoder(&stdout)
				var res output.Result
				require.NoError(t, dec.Decode(&res))
				assert.Equal(t, tt.paths[0], res.RepoAbsPath)
				assert.Equal(t, []string{"git", "fetch", "--all"}, res.Command)
				assert.Equal(t, tt.runnerOutput, res.Stdout)
			}

			if tt.outMode == "markdown" {
				out := stdout.String()
				assert.Contains(t, out, "## fetch: root = /repos, command = git fetch --prune\n\n")
				assert.Contains(t, out, "### repo1: exit code = 0\n\n")
				assert.Contains(t, out, "<details>\n<summary>stdout</summary>\n\n```\nFetching origin\n```\n\n</details>")
				assert.Contains(t, out, "## Summary\n\n")
			}
		})
	}
}
