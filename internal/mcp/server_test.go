package mcp_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/berquerant/git-iter-go/internal/config"
	iterMcp "github.com/berquerant/git-iter-go/internal/mcp"
	"github.com/berquerant/git-iter-go/testutil"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMcpClientSession(t *testing.T, cfg *config.Config, opts ...iterMcp.ServerOption) *mcpsdk.ClientSession {
	t.Helper()
	server := iterMcp.NewServer(cfg, opts...)
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

func decodeToolResult[T any](t *testing.T, res *mcpsdk.CallToolResult) T {
	t.Helper()
	require.False(t, res.IsError)
	data, err := json.Marshal(res.StructuredContent)
	require.NoError(t, err)
	var out T
	require.NoError(t, json.Unmarshal(data, &out))
	return out
}

func TestMcpServer_ListTools(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		ReposRoot: t.TempDir(),
		MaxProcs:  1,
	}
	ctx := context.Background()

	t.Run("list tools with do, get, and worktree enabled", func(t *testing.T) {
		clientSession := setupMcpClientSession(
			t,
			cfg,
			iterMcp.WithEnableDo(true),
			iterMcp.WithEnableGet(true),
			iterMcp.WithEnableWorktree(true),
		)
		res, err := clientSession.ListTools(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, res)

		var toolNames []string
		for _, tool := range res.Tools {
			toolNames = append(toolNames, tool.Name)
		}

		assert.Contains(t, toolNames, "git_iter_list")
		assert.Contains(t, toolNames, "git_iter_worktree_list")
		assert.Contains(t, toolNames, "git_iter_worktree_add")
		assert.Contains(t, toolNames, "git_iter_worktree_remove")
		assert.Contains(t, toolNames, "git_iter_worktree_prune")
		assert.Contains(t, toolNames, "git_iter_do")
		assert.Contains(t, toolNames, "git_iter_get")
		assert.Contains(t, toolNames, "git_iter_grep")
		assert.Contains(t, toolNames, "git_iter_ls_files")
		assert.Contains(t, toolNames, "git_iter_status")
		assert.Contains(t, toolNames, "git_iter_diff")
	})

	t.Run("list tools with mutating tools disabled by default", func(t *testing.T) {
		cSession := setupMcpClientSession(t, cfg)
		res, err := cSession.ListTools(ctx, nil)
		require.NoError(t, err)

		var toolNames []string
		for _, tool := range res.Tools {
			toolNames = append(toolNames, tool.Name)
		}

		assert.Contains(t, toolNames, "git_iter_list")
		assert.Contains(t, toolNames, "git_iter_worktree_list")
		assert.NotContains(t, toolNames, "git_iter_do")
		assert.NotContains(t, toolNames, "git_iter_get")
		assert.NotContains(t, toolNames, "git_iter_worktree_add")
		assert.NotContains(t, toolNames, "git_iter_worktree_remove")
		assert.NotContains(t, toolNames, "git_iter_worktree_prune")
	})
}

func TestMcpServer_ReadOnlyTools(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repo1 := testutil.MakeGitRepo(t, tmpDir, "repo1")
	repo2 := testutil.MakeGitRepo(t, tmpDir, "repo2")
	require.NoError(t, os.WriteFile(filepath.Join(repo1, "hello.txt"), []byte("hello world\n"), 0o644))

	cfg := &config.Config{
		ReposRoot: tmpDir,
		MaxProcs:  1,
	}
	clientSession := setupMcpClientSession(t, cfg)
	ctx := context.Background()

	t.Run("call git_iter_list", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_list",
			Arguments: map[string]any{"patterns": []string{"repo1"}},
		})
		require.NoError(t, err)
		out := decodeToolResult[iterMcp.ListToolsOutput](t, res)
		require.Len(t, out.Repositories, 1)
		assert.Equal(t, repo1, out.Repositories[0].RepoAbsPath)
	})

	t.Run("call git_iter_list all repos", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_list",
			Arguments: map[string]any{"sort": true},
		})
		require.NoError(t, err)
		out := decodeToolResult[iterMcp.ListToolsOutput](t, res)
		require.Len(t, out.Repositories, 2)
		assert.Equal(t, repo1, out.Repositories[0].RepoAbsPath)
		assert.Equal(t, repo2, out.Repositories[1].RepoAbsPath)
	})

	t.Run("call git_iter_list with limit", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_list",
			Arguments: map[string]any{"sort": true, "limit": 1},
		})
		require.NoError(t, err)
		out := decodeToolResult[iterMcp.ListToolsOutput](t, res)
		require.Len(t, out.Repositories, 1)
		assert.Equal(t, repo1, out.Repositories[0].RepoAbsPath)
	})

	t.Run("call git_iter_grep", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_grep",
			Arguments: map[string]any{"grep_args": []string{"hello"}, "patterns": []string{"repo1"}},
		})
		require.NoError(t, err)
		out := decodeToolResult[iterMcp.GrepToolOutput](t, res)
		require.Len(t, out.Results, 1)
		assert.Equal(t, repo1, out.Results[0].RepoAbsPath)
	})

	t.Run("call git_iter_grep missing grep_args", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_grep",
			Arguments: map[string]any{"patterns": []string{"repo1"}},
		})
		require.NoError(t, err)
		assert.True(t, res.IsError)
	})

	t.Run("call git_iter_ls_files, status, and diff", func(t *testing.T) {
		lsRes, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_ls_files",
			Arguments: map[string]any{"patterns": []string{"repo1"}},
		})
		require.NoError(t, err)
		lsOut := decodeToolResult[iterMcp.LsFilesToolOutput](t, lsRes)
		require.Len(t, lsOut.Results, 1)
		assert.Equal(t, repo1, lsOut.Results[0].RepoAbsPath)

		stRes, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_status",
			Arguments: map[string]any{"patterns": []string{"repo1"}},
		})
		require.NoError(t, err)
		stOut := decodeToolResult[iterMcp.StatusToolOutput](t, stRes)
		require.Len(t, stOut.Results, 1)
		assert.Equal(t, repo1, stOut.Results[0].RepoAbsPath)

		diffRes, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_diff",
			Arguments: map[string]any{"patterns": []string{"repo1"}},
		})
		require.NoError(t, err)
		diffOut := decodeToolResult[iterMcp.DiffToolOutput](t, diffRes)
		require.Len(t, diffOut.Results, 1)
		assert.Equal(t, repo1, diffOut.Results[0].RepoAbsPath)
	})
}

func TestMcpServer_DoAndGetTools(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repo1 := testutil.MakeGitRepo(t, tmpDir, "repo1")

	getTmpDir := t.TempDir()
	existingRepo := testutil.MakeGitRepo(t, getTmpDir, filepath.Join("github.com", "org", "existing"))

	cfg := &config.Config{
		ReposRoot: tmpDir,
		MaxProcs:  1,
	}
	clientSession := setupMcpClientSession(t, cfg, iterMcp.WithEnableDo(true), iterMcp.WithEnableGet(true))
	ctx := context.Background()

	t.Run("call git_iter_do", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name: "git_iter_do",
			Arguments: map[string]any{
				"command":  []string{"echo", "hi"},
				"patterns": []string{"repo1"},
			},
		})
		require.NoError(t, err)
		out := decodeToolResult[iterMcp.DoToolOutput](t, res)
		require.Len(t, out.Results, 1)
		assert.Equal(t, repo1, out.Results[0].RepoAbsPath)
		assert.Equal(t, 0, out.Results[0].ExitCode)
		assert.Contains(t, out.Results[0].Stdout, "hi")
	})

	t.Run("call git_iter_do missing command", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_do",
			Arguments: map[string]any{"patterns": []string{"repo1"}},
		})
		require.NoError(t, err)
		assert.True(t, res.IsError)
	})

	t.Run("call git_iter_get without uri lists repositories", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_get",
			Arguments: map[string]any{"patterns": []string{"repo1"}},
		})
		require.NoError(t, err)
		out := decodeToolResult[iterMcp.GetToolOutput](t, res)
		require.Len(t, out.Repositories, 1)
		assert.Equal(t, repo1, out.Repositories[0].RepoAbsPath)
	})

	t.Run("call git_iter_get with existing repo uri", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name: "git_iter_get",
			Arguments: map[string]any{
				"repos_root": getTmpDir,
				"uri":        "https://github.com/org/existing.git",
			},
		})
		require.NoError(t, err)
		out := decodeToolResult[iterMcp.GetToolOutput](t, res)
		require.Len(t, out.Repositories, 1)
		assert.Equal(t, existingRepo, out.Repositories[0].RepoAbsPath)
		assert.Equal(t, filepath.Join("github.com", "org", "existing"), out.Repositories[0].RepoRelPath)
	})

	t.Run("call git_iter_get with invalid uri", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_get",
			Arguments: map[string]any{"uri": "invalid"},
		})
		require.NoError(t, err)
		assert.True(t, res.IsError)
	})
}

func TestMcpServer_WorktreeTools(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repo1 := testutil.MakeGitRepo(t, tmpDir, "repo1")

	wtTmpDir := t.TempDir()
	wt1 := testutil.MakeGitRepo(t, wtTmpDir, filepath.Join("repo1", "wt-feat"))

	cfg := &config.Config{
		ReposRoot: tmpDir,
		MaxProcs:  1,
	}
	clientSession := setupMcpClientSession(t, cfg, iterMcp.WithEnableWorktree(true))
	ctx := context.Background()

	t.Run("call git_iter_worktree_list", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_worktree_list",
			Arguments: map[string]any{"worktree_root": wtTmpDir},
		})
		require.NoError(t, err)
		out := decodeToolResult[iterMcp.WorktreeListToolOutput](t, res)
		require.Len(t, out.Repositories, 1)
		assert.Equal(t, wt1, out.Repositories[0].RepoAbsPath)
	})

	t.Run("call git_iter_worktree_list with repo=true", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_worktree_list",
			Arguments: map[string]any{"worktree_root": wtTmpDir, "repo": true},
		})
		require.NoError(t, err)
		out := decodeToolResult[iterMcp.WorktreeListToolOutput](t, res)
		require.Len(t, out.Repositories, 1)
		assert.Equal(t, repo1, out.Repositories[0].RepoAbsPath)
	})

	t.Run("call git_iter_worktree_prune dry_run", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name: "git_iter_worktree_prune",
			Arguments: map[string]any{
				"worktree_root": wtTmpDir,
				"all":           true,
				"dry_run":       true,
			},
		})
		require.NoError(t, err)
		out := decodeToolResult[iterMcp.WorktreePruneToolOutput](t, res)
		require.Len(t, out.Results, 1)
		assert.Equal(t, repo1, out.Results[0].RepoAbsPath)
	})

	t.Run("call git_iter_worktree_add and remove with invalid branch", func(t *testing.T) {
		addRes, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_worktree_add",
			Arguments: map[string]any{"repo_dir": repo1, "branch": "-invalid"},
		})
		require.NoError(t, err)
		assert.True(t, addRes.IsError)

		rmRes, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name:      "git_iter_worktree_remove",
			Arguments: map[string]any{"repo_dir": repo1, "branch": "-invalid"},
		})
		require.NoError(t, err)
		assert.True(t, rmRes.IsError)
	})
}
