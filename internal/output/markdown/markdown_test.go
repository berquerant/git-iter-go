package markdown_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/berquerant/git-iter-go/internal/output/common"
	"github.com/berquerant/git-iter-go/internal/output/markdown"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownExecutor_Run(t *testing.T) {
	t.Parallel()

	tasks := []common.Task{
		{
			RepoAbsPath: "/repos/a",
			ReposRoot:   "/repos",
			Dir:         "/repos/a",
			Command:     []string{"echo", "hi"},
			Runner:      &testutil.FakeRunner{Output: "hi\n"},
		},
		{
			RepoAbsPath: "/repos/b",
			ReposRoot:   "/repos",
			Dir:         "/repos/b",
			Command:     []string{"false"},
			Runner:      &testutil.FakeRunner{Err: assert.AnError},
		},
	}

	var buf bytes.Buffer
	ex := &markdown.MarkdownExecutor{
		Title:       "do: root = /repos, command = echo hi",
		MaxProcs:    1,
		HeaderLevel: 3,
		Out:         &buf,
	}

	err := ex.Run(context.Background(), tasks)
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "## do: root = /repos, command = echo hi\n\n")
	assert.Contains(t, out, "### a: exit code = 0\n\n")
	assert.Contains(t, out, "<details>\n<summary>stdout</summary>\n\n```\nhi\n```\n\n</details>")
	assert.Contains(t, out, "### b: exit code = 1\n\n")
	assert.Contains(t, out, "## Summary\n\n")
	assert.Contains(t, out, "- **Total Repositories**: 2\n")
	assert.Contains(t, out, "  - Exit Code `0`: 1\n")
	assert.Contains(t, out, "  - Exit Code `1`: 1\n")
	assert.Contains(t, out, "  - `b` (Exit Code `1`)\n")
}

func TestLsFilesMarkdownExecutor_Run(t *testing.T) {
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
	ex := &markdown.LsFilesMarkdownExecutor{
		Title:       "ls-files: root = /repos",
		MaxProcs:    1,
		HeaderLevel: 3,
		Out:         &buf,
	}

	err := ex.Run(context.Background(), tasks)
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "## ls-files: root = /repos\n\n")
	assert.Contains(t, out, "### org/repo1\n\nTotal: 2\n\n<details>\n<summary>Files</summary>\n\n- main.go\n- cmd/root.go\n\n</details>\n\n")
	assert.Contains(t, out, "### org/repo2\n\nTotal: 1\n\n<details>\n<summary>Files</summary>\n\n- README.md\n\n</details>\n\n")
	assert.Contains(t, out, "---\n\n## Summary\n\n")
	assert.Contains(t, out, "- **Total Repositories**: 2\n")
	assert.Contains(t, out, "- **Total Files**: 3\n")
	assert.Contains(t, out, "- **Exit Codes**:\n  - Exit Code `0`: 2\n")
	assert.Contains(t, out, "- **Failed Repositories**:\n  - None (All succeeded)\n")
}
