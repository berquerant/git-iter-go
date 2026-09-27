package markdown_test

import (
	"bytes"
	"testing"

	"github.com/berquerant/git-iter-go/internal/markdown"
	"github.com/berquerant/git-iter-go/internal/output/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteResult(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		result      common.Result
		headerLevel int
		contains    []string
	}{
		{
			name: "header level 3 with stdout and stderr",
			result: common.Result{
				RepoAbsPath: "/repos/org/repo1",
				RepoRelPath: "org/repo1",
				Command:     []string{"git", "status"},
				ExitCode:    0,
				Stdout:      "On branch main\n",
				Stderr:      "some warning\n",
			},
			headerLevel: 3,
			contains: []string{
				"### org/repo1: exit code = 0\n\n",
				"<details>\n<summary>stdout</summary>\n\n```\nOn branch main\n```\n\n</details>",
				"<details>\n<summary>stderr</summary>\n\n```\nsome warning\n```\n\n</details>",
			},
		},
		{
			name: "header level 1 with empty output and error exit code",
			result: common.Result{
				RepoAbsPath: "/repos/org/repo2",
				RepoRelPath: "org/repo2",
				Command:     []string{"false"},
				ExitCode:    1,
			},
			headerLevel: 1,
			contains: []string{
				"# org/repo2: exit code = 1\n\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			err := markdown.WriteResult(&buf, tt.result, tt.headerLevel)
			require.NoError(t, err)

			out := buf.String()
			for _, exp := range tt.contains {
				assert.Contains(t, out, exp)
			}
		})
	}
}

func TestWriteSummary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		results     []common.Result
		headerLevel int
		contains    []string
	}{
		{
			name: "mixed exit codes with failures",
			results: []common.Result{
				{RepoRelPath: "r1", ExitCode: 0},
				{RepoRelPath: "r2", ExitCode: 1},
				{RepoRelPath: "r3", ExitCode: 0},
			},
			headerLevel: 3,
			contains: []string{
				"## Summary\n\n",
				"- **Total Repositories**: 3\n",
				"  - Exit Code `0`: 2\n",
				"  - Exit Code `1`: 1\n",
				"  - `r2` (Exit Code `1`)\n",
			},
		},
		{
			name: "all successful",
			results: []common.Result{
				{RepoRelPath: "r1", ExitCode: 0},
			},
			headerLevel: 2,
			contains: []string{
				"# Summary\n\n",
				"- **Total Repositories**: 1\n",
				"  - Exit Code `0`: 1\n",
				"  - None (All succeeded)\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			err := markdown.WriteSummary(&buf, tt.results, tt.headerLevel)
			require.NoError(t, err)

			out := buf.String()
			for _, exp := range tt.contains {
				assert.Contains(t, out, exp)
			}
		})
	}
}

func TestWriteRepo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		result      common.RepoResult
		headerLevel int
		want        string
	}{
		{
			name: "header level 2",
			result: common.RepoResult{
				RepoAbsPath: "/repos/org/repo1",
				RepoRelPath: "org/repo1",
			},
			headerLevel: 2,
			want:        "## org/repo1\n\n",
		},
		{
			name: "header level 4",
			result: common.RepoResult{
				RepoAbsPath: "/repos/org/repo2",
				RepoRelPath: "org/repo2",
			},
			headerLevel: 4,
			want:        "#### org/repo2\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			err := markdown.WriteRepo(&buf, tt.result, tt.headerLevel)
			require.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestWriteList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		results     []common.RepoResult
		reposRoot   string
		headerLevel int
		want        string
	}{
		{
			name: "with reposRoot",
			results: []common.RepoResult{
				{RepoAbsPath: "/repos/repo1", RepoRelPath: "repo1"},
				{RepoAbsPath: "/repos/repo2", RepoRelPath: "repo2"},
			},
			reposRoot:   "/repos",
			headerLevel: 3,
			want: `## List: root = /repos

- repo1
- repo2

---

### Summary

- **Total Repositories**: 2

`,
		},
		{
			name: "without reposRoot",
			results: []common.RepoResult{
				{RepoAbsPath: "/repos/repo1"},
			},
			reposRoot:   "",
			headerLevel: 2,
			want: `# List

- /repos/repo1

---

## Summary

- **Total Repositories**: 1

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			err := markdown.WriteList(&buf, tt.results, tt.reposRoot, tt.headerLevel)
			require.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}
