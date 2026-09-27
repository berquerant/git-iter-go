package markdown_test

import (
	"bytes"
	"testing"

	"github.com/berquerant/git-iter-go/internal/markdown"
	"github.com/berquerant/git-iter-go/internal/output/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteLsFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		result      common.Result
		headerLevel int
		contains    []string
	}{
		{
			name: "files and stderr with header level 3",
			result: common.Result{
				RepoAbsPath: "/repos/org/repo1",
				RepoRelPath: "org/repo1",
				ExitCode:    0,
				Stdout:      "file1.go\nfile2.go\n",
				Stderr:      "warning message\n",
			},
			headerLevel: 3,
			contains: []string{
				"### org/repo1\n\n",
				"Total: 2\n\n",
				"<details>\n<summary>Files</summary>\n\n- file1.go\n- file2.go\n\n</details>",
				"<details>\n<summary>stderr</summary>\n\n```\nwarning message\n```\n\n</details>",
			},
		},
		{
			name: "empty files",
			result: common.Result{
				RepoAbsPath: "/repos/repo2",
				RepoRelPath: "repo2",
				ExitCode:    0,
				Stdout:      "",
			},
			headerLevel: 2,
			contains: []string{
				"## repo2\n\n",
				"Total: 0\n\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			err := markdown.WriteLsFiles(&buf, tt.result, tt.headerLevel)
			require.NoError(t, err)

			out := buf.String()
			for _, exp := range tt.contains {
				assert.Contains(t, out, exp)
			}
		})
	}
}

func TestWriteLsFilesSummary(t *testing.T) {
	t.Parallel()

	results := []common.Result{
		{
			RepoRelPath: "repo1",
			ExitCode:    0,
			Stdout:      "a.go\nb.go\n",
		},
		{
			RepoRelPath: "repo2",
			ExitCode:    1,
			Stdout:      "c.go\n",
		},
	}

	var buf bytes.Buffer
	err := markdown.WriteLsFilesSummary(&buf, results, 3)
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "---\n\n## Summary\n\n")
	assert.Contains(t, out, "- **Total Repositories**: 2\n")
	assert.Contains(t, out, "- **Total Files**: 3\n")
	assert.Contains(t, out, "- **Exit Codes**:\n  - Exit Code `0`: 1\n  - Exit Code `1`: 1\n")
	assert.Contains(t, out, "- **Failed Repositories**:\n  - `repo2` (Exit Code `1`)\n")
}
