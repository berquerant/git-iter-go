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

	tests := []struct {
		name   string
		branch string
		want   string
	}{
		{
			name:   "simple branch name",
			branch: "feat-1",
			want:   filepath.Join(wtRoot, "github.com", "org", "repo", "feat-1"),
		},
		{
			name:   "nested branch name with slash",
			branch: "feat/sub",
			want:   filepath.Join(wtRoot, "github.com", "org", "repo", "feat", "sub"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := git.ResolveWorktreePath(reposRoot, wtRoot, repoDir, tt.branch)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveWorktreePath_Error(t *testing.T) {
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
	}{
		{name: "empty branch", reposRoot: reposRoot, worktreeRoot: wtRoot, repoDir: repoDir, branch: ""},
		{name: "leading dash branch injection", reposRoot: reposRoot, worktreeRoot: wtRoot, repoDir: repoDir, branch: "--detach"},
		{name: "path traversal branch", reposRoot: reposRoot, worktreeRoot: wtRoot, repoDir: repoDir, branch: "feat/../evil"},
		{name: "branch with leading or trailing whitespace", reposRoot: reposRoot, worktreeRoot: wtRoot, repoDir: repoDir, branch: " feat-1 "},
		{name: "branch with whitespace-only segment", reposRoot: reposRoot, worktreeRoot: wtRoot, repoDir: repoDir, branch: "feat/ /sub"},
		{name: "branch with whitespace-padded segment", reposRoot: reposRoot, worktreeRoot: wtRoot, repoDir: repoDir, branch: "feat/ sub/leaf"},
		{name: "empty worktreeRoot", reposRoot: reposRoot, worktreeRoot: "", repoDir: repoDir, branch: "feat-1"},
		{name: "empty reposRoot", reposRoot: "", worktreeRoot: wtRoot, repoDir: repoDir, branch: "feat-1"},
		{name: "repo outside reposRoot", reposRoot: reposRoot, worktreeRoot: wtRoot, repoDir: outsideRepo, branch: "feat-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := git.ResolveWorktreePath(tt.reposRoot, tt.worktreeRoot, tt.repoDir, tt.branch)
			require.Error(t, err)
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

func TestResolveMainRepoFromWorktree_LinkedGitWorktree(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	reposRoot := filepath.Join(tmp, "repos")
	wtRoot := filepath.Join(tmp, "worktrees")
	mainRepo := filepath.Join(reposRoot, "github.com", "org", "myrepo")
	mainGitDir := filepath.Join(mainRepo, ".git")
	require.NoError(t, os.MkdirAll(mainGitDir, 0o755))

	wtGitDir := filepath.Join(mainGitDir, "worktrees", "feat-wt")
	require.NoError(t, os.MkdirAll(wtGitDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(wtGitDir, "commondir"), []byte("../..\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(wtGitDir, "HEAD"), []byte("ref: refs/heads/feature/login\n"), 0o644))

	wtDir := filepath.Join(wtRoot, "org", "myrepo", "feat-wt")
	require.NoError(t, os.MkdirAll(wtDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(wtDir, ".git"), []byte("gitdir: "+wtGitDir+"\n"), 0o644))

	gotRepo, gotBranch, err := git.ResolveMainRepoFromWorktree(reposRoot, wtRoot, wtDir)
	require.NoError(t, err)
	assert.Equal(t, mainRepo, gotRepo)
	assert.Equal(t, "feature/login", gotBranch)
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

	t.Run("BranchDelete", func(t *testing.T) {
		t.Parallel()
		fake := newFakeRunner()
		c := git.New("git", fake)
		require.NoError(t, c.BranchDelete(context.Background(), "/repos/r", "feat", false, io.Discard))
		require.NoError(t, c.BranchDelete(context.Background(), "/repos/r", "feat2", true, io.Discard))
		require.Error(t, c.BranchDelete(context.Background(), "/repos/r", "-invalid", false, io.Discard))
		assert.Equal(t, [][]string{
			{"git", "branch", "-d", "--", "feat"},
			{"git", "branch", "-D", "--", "feat2"},
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

func TestFindMainRepo(t *testing.T) {
	t.Parallel()

	reposRoot := t.TempDir()
	wtRoot := filepath.Join(reposRoot, "git-worktree")
	mainRepo := filepath.Join(reposRoot, "org", "repo1")
	mainGitDir := filepath.Join(mainRepo, ".git")
	require.NoError(t, os.MkdirAll(mainGitDir, 0o755))

	wtRepo := filepath.Join(wtRoot, "org", "repo1", "feat-1")
	require.NoError(t, os.MkdirAll(wtRepo, 0o755))

	linkedWorktreeGitDir := filepath.Join(mainGitDir, "worktrees", "feat-1")
	require.NoError(t, os.MkdirAll(linkedWorktreeGitDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(wtRepo, ".git"), []byte("gitdir: "+linkedWorktreeGitDir+"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(linkedWorktreeGitDir, "commondir"), []byte("../../\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(linkedWorktreeGitDir, "HEAD"), []byte("ref: refs/heads/feat-1\n"), 0o644))

	t.Run("inside main repo", func(t *testing.T) {
		t.Parallel()
		got, err := git.FindMainRepo(reposRoot, wtRoot, mainRepo)
		require.NoError(t, err)
		assert.Equal(t, mainRepo, got)
	})

	t.Run("inside linked worktree", func(t *testing.T) {
		t.Parallel()
		got, err := git.FindMainRepo(reposRoot, wtRoot, wtRepo)
		require.NoError(t, err)
		assert.Equal(t, mainRepo, got)
	})

	t.Run("non git directory", func(t *testing.T) {
		t.Parallel()
		_, err := git.FindMainRepo(reposRoot, wtRoot, t.TempDir())
		require.Error(t, err)
	})
}
