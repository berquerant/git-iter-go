package runner_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/berquerant/git-iter-go/internal/cmd/runner"
	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunWorktreeList(t *testing.T) {
	t.Parallel()

	reposRoot := t.TempDir()
	wtRoot := filepath.Join(reposRoot, "git-worktree")
	repo1 := filepath.Join(reposRoot, "org", "repo1")
	repo2 := filepath.Join(reposRoot, "org", "repo2")
	require.NoError(t, os.MkdirAll(filepath.Join(repo1, ".git"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(repo2, ".git"), 0o755))

	wt1 := filepath.Join(wtRoot, "org", "repo1", "feat-a")
	wt2 := filepath.Join(wtRoot, "org", "repo1", "feat-b")
	wt3 := filepath.Join(wtRoot, "org", "repo2", "feat-c")

	cfg := &config.Config{
		ReposRoot:    reposRoot,
		WorktreeRoot: wtRoot,
	}

	t.Run("lists worktree paths", func(t *testing.T) {
		t.Parallel()
		finder := &testutil.FakeFinder{Paths: []string{wt1, wt2, wt3}}
		var stdout bytes.Buffer
		err := runner.RunWorktreeList(context.Background(), cfg, "text", parse.WorktreeListArgs{}, finder, &stdout)
		require.NoError(t, err)
		assert.Equal(t, wt1+"\n"+wt2+"\n"+wt3+"\n", stdout.String())
	})

	t.Run("lists unique main repo paths when Repo is true", func(t *testing.T) {
		t.Parallel()
		finder := &testutil.FakeFinder{Paths: []string{wt1, wt2, wt3}}
		var stdout bytes.Buffer
		err := runner.RunWorktreeList(context.Background(), cfg, "text", parse.WorktreeListArgs{Repo: true}, finder, &stdout)
		require.NoError(t, err)
		assert.Equal(t, repo1+"\n"+repo2+"\n", stdout.String())
	})

	t.Run("errors when worktree root is empty", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		err := runner.RunWorktreeList(context.Background(), &config.Config{}, "text", parse.WorktreeListArgs{}, &testutil.FakeFinder{}, &stdout)
		require.Error(t, err)
	})
}

func TestRunWorktreeAddAndRemove(t *testing.T) {
	t.Parallel()

	reposRoot := t.TempDir()
	wtRoot := filepath.Join(reposRoot, "git-worktree")
	repoDir := filepath.Join(reposRoot, "org", "repo")
	require.NoError(t, os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755))

	cfg := &config.Config{
		ReposRoot:    reposRoot,
		WorktreeRoot: wtRoot,
		GitCommand:   "git",
		Timeout:      5 * time.Second,
	}

	t.Run("add new branch", func(t *testing.T) {
		t.Parallel()
		fake := &testutil.FakeRunner{}
		var stdout, stderr bytes.Buffer
		err := runner.RunWorktreeAdd(
			context.Background(),
			cfg,
			"text",
			parse.WorktreeAddArgs{Branch: "main", NewBranch: "feat-1"},
			repoDir,
			fake,
			&stdout,
			&stderr,
		)
		require.NoError(t, err)
		wantWT := filepath.Join(wtRoot, "org", "repo", "feat-1")
		assert.Equal(t, wantWT+"\n", stdout.String())
		require.Len(t, fake.Calls, 1)
		assert.Equal(t, []string{"git", "worktree", "add", "-b", "feat-1", "--", wantWT, "main"}, fake.Calls[0].Command)
	})

	t.Run("remove branch with force", func(t *testing.T) {
		t.Parallel()
		fake := &testutil.FakeRunner{}
		var stderr bytes.Buffer
		err := runner.RunWorktreeRemove(
			context.Background(),
			cfg,
			parse.WorktreeRemoveArgs{Branch: "feat-1", Force: true},
			repoDir,
			fake,
			&stderr,
		)
		require.NoError(t, err)
		wantWT := filepath.Join(wtRoot, "org", "repo", "feat-1")
		require.Len(t, fake.Calls, 1)
		assert.Equal(t, []string{"git", "worktree", "remove", "--force", "--", wantWT}, fake.Calls[0].Command)
	})
}

func TestRunWorktreePrune(t *testing.T) {
	t.Parallel()

	reposRoot := t.TempDir()
	wtRoot := filepath.Join(reposRoot, "git-worktree")
	repo1 := filepath.Join(reposRoot, "org", "repo1")
	repo2 := filepath.Join(reposRoot, "org", "repo2")
	require.NoError(t, os.MkdirAll(filepath.Join(repo1, ".git"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(repo2, ".git"), 0o755))

	wt1 := filepath.Join(wtRoot, "org", "repo1", "feat-a")
	wt2 := filepath.Join(wtRoot, "org", "repo2", "feat-b")

	cfg := &config.Config{
		ReposRoot:    reposRoot,
		WorktreeRoot: wtRoot,
		GitCommand:   "git",
	}

	t.Run("single repository prune", func(t *testing.T) {
		t.Parallel()
		fake := &testutil.FakeRunner{}
		var stdout, stderr bytes.Buffer
		err := runner.RunWorktreePrune(
			context.Background(),
			cfg,
			"text",
			parse.WorktreePruneArgs{DryRun: true},
			repo1,
			&testutil.FakeFinder{},
			fake,
			&stdout,
			&stderr,
		)
		require.NoError(t, err)
		require.Len(t, fake.Calls, 1)
		assert.Equal(t, repo1, fake.Calls[0].Dir)
		assert.Equal(t, []string{"git", "worktree", "prune", "--dry-run"}, fake.Calls[0].Command)
	})

	t.Run("all repositories prune", func(t *testing.T) {
		t.Parallel()
		fake := &testutil.FakeRunner{}
		finder := &testutil.FakeFinder{Paths: []string{wt1, wt2}}
		var stdout, stderr bytes.Buffer
		err := runner.RunWorktreePrune(
			context.Background(),
			cfg,
			"text",
			parse.WorktreePruneArgs{All: true},
			"",
			finder,
			fake,
			&stdout,
			&stderr,
		)
		require.NoError(t, err)
		require.Len(t, fake.Calls, 2)
		assert.Equal(t, repo1, fake.Calls[0].Dir)
		assert.Equal(t, repo2, fake.Calls[1].Dir)
	})
}
