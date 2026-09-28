package cmd_test

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/berquerant/git-iter-go/internal/cmd"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestNoRepoSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "neither flag set causes error",
			args:    []string{"list"},
			wantErr: true,
		},
		{
			name:    "repos-root set is OK",
			args:    []string{"--repos-root", t.TempDir(), "list"},
			wantErr: false,
		},
		{
			name:    "list-repos set is OK",
			args:    []string{"--list-repos", "echo /tmp", "list"},
			wantErr: false,
		},
		{
			name:    "version does not require repo source",
			args:    []string{"version"},
			wantErr: false,
		},
		{
			name:    "help does not require repo source",
			args:    []string{"help"},
			wantErr: false,
		},
		{
			name:    "read does not require repo source",
			args:    []string{"read", "echo"},
			wantErr: false,
		},
		{
			name:    "mcp does not require repo source",
			args:    []string{"mcp", "--help"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := cmd.NewRootCmd()
			c.SetOut(&bytes.Buffer{})
			c.SetErr(&bytes.Buffer{})
			c.SetArgs(tt.args)
			err := c.ExecuteContext(context.Background())
			if tt.wantErr {
				require.ErrorIs(t, err, cmd.ErrNoRepoSource)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestListCmd_Output(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	testutil.MakeGitRepo(t, root, "repo-b")
	testutil.MakeGitRepo(t, root, "repo-a")
	testutil.MakeGitRepo(t, root, "repo-c")

	repoA := filepath.Join(root, "repo-a")
	repoB := filepath.Join(root, "repo-b")
	repoC := filepath.Join(root, "repo-c")

	tests := []struct {
		name      string
		args      []string
		wantLines []string
	}{
		{
			name:      "sort ascending",
			args:      []string{"--repos-root", root, "list", "--sort"},
			wantLines: []string{repoA, repoB, repoC},
		},
		{
			name:      "sort descending (reverse)",
			args:      []string{"--repos-root", root, "list", "--sort-reverse"},
			wantLines: []string{repoC, repoB, repoA},
		},
		{
			name:      "limit after sort ascending",
			args:      []string{"--repos-root", root, "-n", "2", "list", "--sort"},
			wantLines: []string{repoA, repoB},
		},
		{
			name:      "limit after sort descending (reverse)",
			args:      []string{"--repos-root", root, "-n", "1", "list", "--sort-reverse"},
			wantLines: []string{repoC},
		},
		{
			name: "markdown output",
			args: []string{"--repos-root", root, "-o", "md", "list", "--sort"},
			wantLines: []string{
				"## List: root = " + root,
				"",
				"- repo-a",
				"- repo-b",
				"- repo-c",
				"",
				"---",
				"",
				"### Summary",
				"",
				"- **Total Repositories**: 3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var stdout bytes.Buffer
			c := cmd.NewRootCmd()
			c.SetOut(&stdout)
			c.SetErr(&bytes.Buffer{})
			c.SetArgs(tt.args)
			err := c.ExecuteContext(context.Background())
			require.NoError(t, err)

			got := strings.Split(strings.TrimSpace(stdout.String()), "\n")
			assert.Equal(t, tt.wantLines, got)
		})
	}
}

func TestCmd_RequiresArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "read requires COMMAND arg",
			args: []string{"read"},
		},
		{
			name: "do requires command after pattern",
			args: []string{"--repos-root", t.TempDir(), "do", "somepattern"},
		},
		{
			name: "grep requires GIT_GREP_ARGS",
			args: []string{"--repos-root", t.TempDir(), "grep"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := cmd.NewRootCmd()
			c.SetOut(&bytes.Buffer{})
			c.SetErr(&bytes.Buffer{})
			c.SetArgs(tt.args)
			err := c.ExecuteContext(context.Background())
			require.Error(t, err)
		})
	}
}

func TestLogLevel_Debug(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "--log-level debug",
			args: []string{"--log-level", "debug"},
		},
		{
			name: "-l debug",
			args: []string{"-l", "debug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			c := cmd.NewRootCmd()
			c.SetOut(&stdout)
			c.SetErr(&stderr)
			c.SetArgs(append([]string{"--repos-root", t.TempDir()}, append(tt.args, "list")...))
			err := c.ExecuteContext(context.Background())
			require.NoError(t, err)
		})
	}
}

func TestSubcommands_Help(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		subcmd       string
		wantContains []string
	}{
		{
			name:         "do help",
			subcmd:       "do",
			wantContains: []string{"Output modes", "text", "json", "markdown"},
		},
		{
			name:         "grep help",
			subcmd:       "grep",
			wantContains: []string{"Output modes", "text", "json", "markdown"},
		},
		{
			name:         "list help",
			subcmd:       "list",
			wantContains: []string{"Output modes", "text", "json", "markdown"},
		},
		{
			name:         "read help",
			subcmd:       "read",
			wantContains: []string{"Output modes", "text", "json", "markdown"},
		},
		{
			name:         "pull help",
			subcmd:       "pull",
			wantContains: []string{"--switch-default", "--switch-back", "Output modes", "text", "json", "markdown"},
		},
		{
			name:         "fetch help",
			subcmd:       "fetch",
			wantContains: []string{"--default", "Output modes", "text", "json", "markdown"},
		},
		{
			name:         "ls-files help",
			subcmd:       "ls-files",
			wantContains: []string{"--show-abs-path", "lsf", "Output modes", "text", "json", "markdown"},
		},
		{
			name:         "status help",
			subcmd:       "status",
			wantContains: []string{"git status", "st", "--dirty-only", "--clean-only", "Output modes", "text", "json", "markdown"},
		},
		{
			name:         "diff help",
			subcmd:       "diff",
			wantContains: []string{"git diff", "--dirty-only", "--clean-only", "--interactive", "-i", "Output modes", "text", "json", "markdown"},
		},
		{
			name:         "mcp help",
			subcmd:       "mcp",
			wantContains: []string{"--enable-do", "git_iter_do"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			c := cmd.NewRootCmd()
			c.SetOut(&stdout)
			c.SetErr(&stderr)
			c.SetArgs([]string{tt.subcmd, "--help"})
			err := c.ExecuteContext(context.Background())
			require.NoError(t, err)

			out := stdout.String()
			for _, want := range tt.wantContains {
				assert.Contains(t, out, want)
			}
		})
	}
}

func TestConflictingFilterFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		flags   []string
		wantErr error
	}{
		{
			name:    "dirty and clean only conflict",
			flags:   []string{"--dirty-only", "--clean-only"},
			wantErr: cmd.ErrConflictingFilterFlags,
		},
		{
			name:    "default branch and not default branch only conflict",
			flags:   []string{"--default-branch-only", "--not-default-branch-only"},
			wantErr: cmd.ErrConflictingBranchFlags,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			c := cmd.NewRootCmd()
			c.SetOut(&stdout)
			c.SetErr(&stderr)
			args := append([]string{"--repos-root", t.TempDir()}, tt.flags...)
			args = append(args, "list")
			c.SetArgs(args)
			err := c.ExecuteContext(context.Background())
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestInvalidRegexFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		flags       []string
		errContains string
	}{
		{
			name:        "invalid remote-url regex",
			flags:       []string{"--remote-url", "[invalid("},
			errContains: "invalid remote-url regex",
		},
		{
			name:        "invalid branch regex",
			flags:       []string{"--branch", "[invalid("},
			errContains: "invalid branch regex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			c := cmd.NewRootCmd()
			c.SetOut(&stdout)
			c.SetErr(&stderr)
			args := append([]string{"--repos-root", t.TempDir()}, tt.flags...)
			args = append(args, "list")
			c.SetArgs(args)
			err := c.ExecuteContext(context.Background())
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}
