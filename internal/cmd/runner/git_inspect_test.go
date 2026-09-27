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

func TestRunLsFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		outMode      string
		paths        []string
		lsFilesArgs  parse.LsFilesArgs
		showAbsPath  bool
		globalAbs    bool
		runnerOutput string
		wantStdout   string
		validateJSON bool
	}{
		{
			name:         "text mode with rel path prefix",
			outMode:      "text",
			paths:        []string{"/repos/org/repo1"},
			lsFilesArgs:  parse.LsFilesArgs{LsFilesFlags: []string{"-c"}},
			showAbsPath:  false,
			globalAbs:    false,
			runnerOutput: "main.go\nREADME.md\n",
			wantStdout:   "org/repo1/main.go\norg/repo1/README.md\n",
		},
		{
			name:         "text mode with showAbsPath",
			outMode:      "text",
			paths:        []string{"/repos/org/repo1"},
			lsFilesArgs:  parse.LsFilesArgs{LsFilesFlags: []string{"-c"}},
			showAbsPath:  true,
			globalAbs:    false,
			runnerOutput: "main.go\n",
			wantStdout:   "/repos/org/repo1/main.go\n",
		},
		{
			name:         "text mode with global abs-path",
			outMode:      "text",
			paths:        []string{"/repos/org/repo1"},
			lsFilesArgs:  parse.LsFilesArgs{LsFilesFlags: []string{"-c"}},
			showAbsPath:  false,
			globalAbs:    true,
			runnerOutput: "main.go\n",
			wantStdout:   "/repos/org/repo1/main.go\n",
		},
		{
			name:         "json mode",
			outMode:      "json",
			paths:        []string{"/repos/org/repo1"},
			lsFilesArgs:  parse.LsFilesArgs{LsFilesFlags: []string{"-c"}},
			runnerOutput: "main.go\n",
			validateJSON: true,
		},
		{
			name:         "markdown mode",
			outMode:      "markdown",
			paths:        []string{"/repos/org/repo1"},
			lsFilesArgs:  parse.LsFilesArgs{LsFilesFlags: []string{"-c"}},
			runnerOutput: "main.go\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &config.Config{
				MaxProcs:            1,
				ReposRoot:           "/repos",
				AbsPath:             tt.globalAbs,
				MarkdownHeaderLevel: 3,
			}
			finder := &testutil.FakeFinder{Paths: tt.paths}
			fakeRunner := &testutil.FakeRunner{Output: tt.runnerOutput}
			var stdout, stderr bytes.Buffer

			err := runner.RunLsFiles(
				context.Background(),
				cfg,
				tt.outMode,
				tt.lsFilesArgs,
				finder,
				fakeRunner,
				tt.showAbsPath,
				&stdout,
				&stderr,
			)
			require.NoError(t, err)

			if tt.validateJSON {
				line := strings.TrimSpace(stdout.String())
				var res output.FileResult
				require.NoError(t, json.Unmarshal([]byte(line), &res))
				assert.Equal(t, "/repos/org/repo1", res.RepoAbsPath)
				assert.Equal(t, "org/repo1", res.RepoRelPath)
				assert.Equal(t, "main.go", res.FileRelPath)
				assert.Equal(t, "/repos/org/repo1/main.go", res.FileAbsPath)
			} else if tt.outMode == "markdown" {
				out := stdout.String()
				assert.Contains(t, out, "## ls-files: root = /repos\n\n")
				assert.Contains(t, out, "### org/repo1\n\nTotal: 1\n\n<details>\n<summary>Files</summary>\n\n- main.go\n\n</details>\n\n")
				assert.Contains(t, out, "---\n\n## Summary\n\n")
				assert.Contains(t, out, "- **Total Repositories**: 1\n")
				assert.Contains(t, out, "- **Total Files**: 1\n")
			} else {
				assert.Equal(t, tt.wantStdout, stdout.String())
			}
		})
	}
}

func TestRunLsFiles_CustomGitCommand(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		MaxProcs:   1,
		ReposRoot:  "/repos",
		GitCommand: "/custom/bin/git",
	}
	finder := &testutil.FakeFinder{Paths: []string{"/repos/r1"}}
	fakeRunner := &testutil.FakeRunner{Output: "file.txt\n"}
	var stdout, stderr bytes.Buffer

	err := runner.RunLsFiles(
		context.Background(),
		cfg,
		"text",
		parse.LsFilesArgs{LsFilesFlags: []string{"-m"}},
		finder,
		fakeRunner,
		false,
		&stdout,
		&stderr,
	)
	require.NoError(t, err)
	require.Len(t, fakeRunner.Calls, 1)
	assert.Equal(t, []string{"/custom/bin/git", "ls-files", "-m"}, fakeRunner.Calls[0].Command)
}

func TestRunStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		outMode      string
		paths        []string
		statusArgs   parse.StatusArgs
		runnerOutput string
		wantStdout   string
		validateJSON bool
	}{
		{
			name:         "text mode default short status",
			outMode:      "text",
			paths:        []string{"/repos/repo1", "/repos/repo2"},
			statusArgs:   parse.StatusArgs{},
			runnerOutput: " M file.go\n",
			wantStdout:   " M file.go\n M file.go\n",
		},
		{
			name:         "json mode",
			outMode:      "json",
			paths:        []string{"/repos/repo1"},
			statusArgs:   parse.StatusArgs{StatusFlags: []string{"--porcelain"}},
			runnerOutput: "?? untracked.txt\n",
			validateJSON: true,
		},
		{
			name:         "markdown mode",
			outMode:      "markdown",
			paths:        []string{"/repos/repo1"},
			statusArgs:   parse.StatusArgs{StatusFlags: []string{"--short"}},
			runnerOutput: " M file.go\n",
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

			err := runner.RunStatus(
				context.Background(),
				cfg,
				tt.outMode,
				tt.statusArgs,
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
				assert.Equal(t, []string{"git", "status", "--porcelain"}, res.Command)
				assert.Equal(t, tt.runnerOutput, res.Stdout)
			}

			if tt.outMode == "markdown" {
				out := stdout.String()
				assert.Contains(t, out, "## status: root = /repos, command = git status --short\n\n")
				assert.Contains(t, out, "### repo1: exit code = 0\n\n")
				assert.Contains(t, out, "<details>\n<summary>stdout</summary>\n\n```\n M file.go\n```\n\n</details>")
				assert.Contains(t, out, "## Summary\n\n")
			}
		})
	}
}

func TestRunDiff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		outMode      string
		paths        []string
		diffArgs     parse.DiffArgs
		runnerOutput string
		wantStdout   string
		validateJSON bool
	}{
		{
			name:         "text mode streaming",
			outMode:      "text",
			paths:        []string{"/repos/repo1", "/repos/repo2"},
			diffArgs:     parse.DiffArgs{DiffFlags: []string{"--stat"}},
			runnerOutput: " 1 file changed, 1 insertion(+)\n",
			wantStdout:   " 1 file changed, 1 insertion(+)\n 1 file changed, 1 insertion(+)\n",
		},
		{
			name:         "json mode",
			outMode:      "json",
			paths:        []string{"/repos/repo1"},
			diffArgs:     parse.DiffArgs{DiffFlags: []string{"--cached"}},
			runnerOutput: "diff --git a/foo b/foo\n",
			validateJSON: true,
		},
		{
			name:         "json mode interactive",
			outMode:      "json",
			paths:        []string{"/repos/repo1"},
			diffArgs:     parse.DiffArgs{DiffFlags: []string{"--cached"}, Interactive: true},
			runnerOutput: "diff --git a/foo b/foo\n",
			validateJSON: true,
		},
		{
			name:         "markdown mode",
			outMode:      "markdown",
			paths:        []string{"/repos/repo1"},
			diffArgs:     parse.DiffArgs{DiffFlags: []string{"--stat"}},
			runnerOutput: " 1 file changed\n",
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

			err := runner.RunDiff(
				context.Background(),
				cfg,
				tt.outMode,
				tt.diffArgs,
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
				if tt.diffArgs.Interactive {
					assert.Equal(t, []string{"git", "diff", "--cached"}, res.Command)
				} else {
					assert.Equal(t, []string{"git", "--no-pager", "diff", "--cached"}, res.Command)
				}
				assert.Equal(t, tt.runnerOutput, res.Stdout)
			}

			if tt.outMode == "markdown" {
				out := stdout.String()
				assert.Contains(t, out, "## diff: root = /repos, command = git --no-pager diff --stat\n\n")
				assert.Contains(t, out, "### repo1: exit code = 0\n\n")
				assert.Contains(t, out, "<details>\n<summary>stdout</summary>\n\n```\n 1 file changed\n```\n\n</details>")
				assert.Contains(t, out, "## Summary\n\n")
			}
		})
	}
}
