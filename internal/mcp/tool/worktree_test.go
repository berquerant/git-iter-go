package tool_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/mcp/tool"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initRealGitRepo(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test User"},
		{"config", "commit.gpgsign", "false"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v: %s", args, string(out))
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("init\n"), 0o644))
	for _, args := range [][]string{
		{"add", "README.md"},
		{"commit", "-m", "initial commit"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v: %s", args, string(out))
	}
}

func setupWorktreeToolSession(t *testing.T, cfg *config.Config) *mcpsdk.ClientSession {
	t.Helper()
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "test-server", Version: "v1.0.0"}, nil)
	tool.RegisterWorktreeList(server, cfg)
	tool.RegisterWorktreeMutations(server, cfg)

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)
	t1, t2 := mcpsdk.NewInMemoryTransports()
	ctx := context.Background()

	serverSession, err := server.Connect(ctx, t1, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverSession.Close() })

	clientSession, err := client.Connect(ctx, t2, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })

	return clientSession
}

func TestWorktreeTools(t *testing.T) {
	t.Parallel()

	reposRoot := t.TempDir()
	wtRoot := filepath.Join(reposRoot, "git-worktree")
	repo1 := filepath.Join(reposRoot, "org", "repo1")
	initRealGitRepo(t, repo1)

	clientSession := setupWorktreeToolSession(t, &config.Config{
		ReposRoot:    reposRoot,
		WorktreeRoot: wtRoot,
	})
	ctx := context.Background()
	expectedWT := filepath.Join(wtRoot, "org", "repo1", "feat-wt")

	t.Run("worktree add with new_branch", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name: "git_iter_worktree_add",
			Arguments: map[string]any{
				"repo_dir":   repo1,
				"branch":     "main",
				"new_branch": "feat-wt",
				"timeout":    "10s",
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)

		data, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out tool.WorktreeAddToolOutput
		require.NoError(t, json.Unmarshal(data, &out))
		assert.Equal(t, expectedWT, out.Worktree.RepoAbsPath)
		assert.Equal(t, filepath.Join("org", "repo1", "feat-wt"), out.Worktree.RepoRelPath)
	})

	t.Run("worktree list and list --repo", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_worktree_list",
			Arguments: map[string]any{},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)

		data, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out tool.WorktreeListToolOutput
		require.NoError(t, json.Unmarshal(data, &out))
		require.Len(t, out.Repositories, 1)
		assert.Equal(t, expectedWT, out.Repositories[0].RepoAbsPath)

		resRepo, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_worktree_list",
			Arguments: map[string]any{"repo": true},
		})
		require.NoError(t, err)
		require.False(t, resRepo.IsError)

		dataRepo, err := json.Marshal(resRepo.StructuredContent)
		require.NoError(t, err)

		var outRepo tool.WorktreeListToolOutput
		require.NoError(t, json.Unmarshal(dataRepo, &outRepo))
		require.Len(t, outRepo.Repositories, 1)
		assert.Equal(t, repo1, outRepo.Repositories[0].RepoAbsPath)
	})

	t.Run("worktree prune and remove", func(t *testing.T) {
		resSingle, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_worktree_prune",
			Arguments: map[string]any{"repo_dir": repo1, "dry_run": true},
		})
		require.NoError(t, err)
		require.False(t, resSingle.IsError)

		resAll, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_worktree_prune",
			Arguments: map[string]any{"all": true, "dry_run": true},
		})
		require.NoError(t, err)
		require.False(t, resAll.IsError)

		resRm, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_worktree_remove",
			Arguments: map[string]any{"repo_dir": repo1, "branch": "feat-wt", "force": true},
		})
		require.NoError(t, err)
		require.False(t, resRm.IsError)
	})
}
