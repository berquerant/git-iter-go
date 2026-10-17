package git_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindRepoRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	repoDir := filepath.Join(root, "github.com", "org", "repo")
	subDir := filepath.Join(repoDir, "pkg", "sub")
	require.NoError(t, os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755))
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	tests := []struct {
		name    string
		dir     string
		want    string
		wantErr bool
	}{
		{
			name: "at repo root",
			dir:  repoDir,
			want: repoDir,
		},
		{
			name: "inside subdirectory of repo",
			dir:  subDir,
			want: repoDir,
		},
		{
			name:    "empty dir",
			dir:     "",
			wantErr: true,
		},
		{
			name:    "not a git repo",
			dir:     t.TempDir(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := git.FindRepoRoot(tt.dir)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveWorktreePath(t *testing.T) {
	t.Parallel()

	reposRoot := t.TempDir()
	wtRoot := filepath.Join(reposRoot, "git-worktree")
	repoDir := filepath.Join(reposRoot, "github.com", "org", "repo")
	require.NoError(t, os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755))

	outsideRepo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(outsideRepo, ".git"), 0o755))

	tests := []struct {
		name         string
		reposRoot    string
		worktreeRoot string
		repoDir      string
		branch       string
		want         string
		wantErr      bool
	}{
		{
			name:         "simple branch name",
			reposRoot:    reposRoot,
			worktreeRoot: wtRoot,
			repoDir:      repoDir,
			branch:       "feat-1",
			want:         filepath.Join(wtRoot, "github.com", "org", "repo", "feat-1"),
		},
		{
			name:         "nested branch name with slash",
			reposRoot:    reposRoot,
			worktreeRoot: wtRoot,
			repoDir:      repoDir,
			branch:       "feat/sub",
			want:         filepath.Join(wtRoot, "github.com", "org", "repo", "feat", "sub"),
		},
		{
			name:         "empty branch",
			reposRoot:    reposRoot,
			worktreeRoot: wtRoot,
			repoDir:      repoDir,
			branch:       "",
			wantErr:      true,
		},
		{
			name:         "leading dash branch injection",
			reposRoot:    reposRoot,
			worktreeRoot: wtRoot,
			repoDir:      repoDir,
			branch:       "--detach",
			wantErr:      true,
		},
		{
			name:         "path traversal branch",
			reposRoot:    reposRoot,
			worktreeRoot: wtRoot,
			repoDir:      repoDir,
			branch:       "feat/../evil",
			wantErr:      true,
		},
		{
			name:         "branch with leading or trailing whitespace",
			reposRoot:    reposRoot,
			worktreeRoot: wtRoot,
			repoDir:      repoDir,
			branch:       " feat-1 ",
			wantErr:      true,
		},
		{
			name:         "branch with whitespace-only segment",
			reposRoot:    reposRoot,
			worktreeRoot: wtRoot,
			repoDir:      repoDir,
			branch:       "feat/ /sub",
			wantErr:      true,
		},
		{
			name:         "branch with whitespace-padded segment",
			reposRoot:    reposRoot,
			worktreeRoot: wtRoot,
			repoDir:      repoDir,
			branch:       "feat/ sub/leaf",
			wantErr:      true,
		},
		{
			name:         "empty worktreeRoot",
			reposRoot:    reposRoot,
			worktreeRoot: "",
			repoDir:      repoDir,
			branch:       "feat-1",
			wantErr:      true,
		},
		{
			name:         "empty reposRoot",
			reposRoot:    "",
			worktreeRoot: wtRoot,
			repoDir:      repoDir,
			branch:       "feat-1",
			wantErr:      true,
		},
		{
			name:         "repo outside reposRoot",
			reposRoot:    reposRoot,
			worktreeRoot: wtRoot,
			repoDir:      outsideRepo,
			branch:       "feat-1",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := git.ResolveWorktreePath(tt.reposRoot, tt.worktreeRoot, tt.repoDir, tt.branch)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveMainRepoFromWorktree(t *testing.T) {
	t.Parallel()

	reposRoot := t.TempDir()
	wtRoot := filepath.Join(reposRoot, "git-worktree")
	repoDir := filepath.Join(reposRoot, "github.com", "org", "repo")
	require.NoError(t, os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755))

	wtSimple := filepath.Join(wtRoot, "github.com", "org", "repo", "feat-1")
	wtNested := filepath.Join(wtRoot, "github.com", "org", "repo", "feat", "nested")
	wtOrphan := filepath.Join(wtRoot, "github.com", "unknown", "repo", "feat-1")

	tests := []struct {
		name         string
		reposRoot    string
		worktreeRoot string
		wtPath       string
		wantRepo     string
		wantBranch   string
		wantErr      bool
	}{
		{
			name:         "simple branch worktree",
			reposRoot:    reposRoot,
			worktreeRoot: wtRoot,
			wtPath:       wtSimple,
			wantRepo:     repoDir,
			wantBranch:   "feat-1",
		},
		{
			name:         "nested branch worktree",
			reposRoot:    reposRoot,
			worktreeRoot: wtRoot,
			wtPath:       wtNested,
			wantRepo:     repoDir,
			wantBranch:   "feat/nested",
		},
		{
			name:         "worktree without matching main repo",
			reposRoot:    reposRoot,
			worktreeRoot: wtRoot,
			wtPath:       wtOrphan,
			wantErr:      true,
		},
		{
			name:         "path outside worktreeRoot",
			reposRoot:    reposRoot,
			worktreeRoot: wtRoot,
			wtPath:       repoDir,
			wantErr:      true,
		},
		{
			name:         "empty reposRoot",
			reposRoot:    "",
			worktreeRoot: wtRoot,
			wtPath:       wtSimple,
			wantErr:      true,
		},
		{
			name:         "empty worktreeRoot",
			reposRoot:    reposRoot,
			worktreeRoot: "",
			wtPath:       wtSimple,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotRepo, gotBranch, err := git.ResolveMainRepoFromWorktree(tt.reposRoot, tt.worktreeRoot, tt.wtPath)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantRepo, gotRepo)
			assert.Equal(t, tt.wantBranch, gotBranch)
		})
	}
}

func TestCommand_WorktreeCommands(t *testing.T) {
	t.Parallel()

	t.Run("WorktreeAdd existing branch", func(t *testing.T) {
		t.Parallel()
		fake := newFakeRunner()
		c := git.New("git", fake)
		err := c.WorktreeAdd(context.Background(), "/repos/r", "/wt/r/feat", "feat", "", io.Discard)
		require.NoError(t, err)
		assert.Equal(t, [][]string{{"git", "worktree", "add", "--", "/wt/r/feat", "feat"}}, fake.Calls())
	})

	t.Run("WorktreeAdd new branch", func(t *testing.T) {
		t.Parallel()
		fake := newFakeRunner()
		c := git.New("git", fake)
		err := c.WorktreeAdd(context.Background(), "/repos/r", "/wt/r/feat", "main", "feat", io.Discard)
		require.NoError(t, err)
		assert.Equal(t, [][]string{{"git", "worktree", "add", "-b", "feat", "--", "/wt/r/feat", "main"}}, fake.Calls())
	})

	t.Run("WorktreeAdd rejects invalid branch names", func(t *testing.T) {
		t.Parallel()
		fake := newFakeRunner()
		c := git.New("git", fake)
		require.Error(t, c.WorktreeAdd(context.Background(), "/repos/r", "/wt/r/feat", "-evil", "", io.Discard))
		require.Error(t, c.WorktreeAdd(context.Background(), "/repos/r", "/wt/r/feat", "main", "-evil", io.Discard))
	})

	t.Run("WorktreeRemove", func(t *testing.T) {
		t.Parallel()
		fake := newFakeRunner()
		c := git.New("git", fake)
		require.NoError(t, c.WorktreeRemove(context.Background(), "/repos/r", "/wt/r/feat", true, false, io.Discard))
		require.NoError(t, c.WorktreeRemove(context.Background(), "/repos/r", "/wt/r/feat2", false, true, io.Discard))
		assert.Equal(t, [][]string{
			{"git", "worktree", "remove", "--force", "--", "/wt/r/feat"},
			{"git", "worktree", "remove", "--no-force", "--", "/wt/r/feat2"},
		}, fake.Calls())
	})

	t.Run("WorktreePrune", func(t *testing.T) {
		t.Parallel()
		fake := newFakeRunner()
		c := git.New("git", fake)
		require.NoError(t, c.WorktreePrune(context.Background(), "/repos/r", false, io.Discard, io.Discard))
		require.NoError(t, c.WorktreePrune(context.Background(), "/repos/r", true, io.Discard, io.Discard))
		assert.Equal(t, [][]string{
			{"git", "worktree", "prune"},
			{"git", "worktree", "prune", "--dry-run"},
		}, fake.Calls())
	})
}

func TestCommand_IsMerged(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(f *fakeRunner)
		branch  string
		target  string
		want    bool
		wantErr bool
	}{
		{
			name: "ancestor merge returns true",
			setup: func(f *fakeRunner) {
				f.on("merge-base", func(_ []string, _, _ io.Writer) error {
					return nil
				})
			},
			branch: "feat",
			target: "main",
			want:   true,
		},
		{
			name: "squash merge with matching tree returns true",
			setup: func(f *fakeRunner) {
				f.on("merge-base", func(_ []string, _, _ io.Writer) error {
					return fmt.Errorf("not ancestor")
				})
				f.on("rev-parse", func(_ []string, stdout, _ io.Writer) error {
					_, _ = fmt.Fprintln(stdout, "tree-hash-123")
					return nil
				})
				f.on("merge-tree", func(_ []string, stdout, _ io.Writer) error {
					_, _ = fmt.Fprintln(stdout, "tree-hash-123")
					return nil
				})
			},
			branch: "feat",
			target: "main",
			want:   true,
		},
		{
			name: "unmerged branch with differing tree returns false",
			setup: func(f *fakeRunner) {
				f.on("merge-base", func(_ []string, _, _ io.Writer) error {
					return fmt.Errorf("not ancestor")
				})
				f.on("rev-parse", func(_ []string, stdout, _ io.Writer) error {
					_, _ = fmt.Fprintln(stdout, "tree-hash-123")
					return nil
				})
				f.on("merge-tree", func(_ []string, stdout, _ io.Writer) error {
					_, _ = fmt.Fprintln(stdout, "tree-hash-456")
					return nil
				})
			},
			branch: "feat",
			target: "main",
			want:   false,
		},
		{
			name:    "invalid branch name returns error",
			setup:   func(_ *fakeRunner) {},
			branch:  "-invalid",
			target:  "main",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fake := newFakeRunner()
			tt.setup(fake)
			c := git.New("git", fake)
			got, err := c.IsMerged(context.Background(), "/repos/r", tt.branch, tt.target)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
