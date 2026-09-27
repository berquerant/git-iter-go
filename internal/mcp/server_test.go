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

func TestMcpServer_Tools(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repo1 := testutil.MakeGitRepo(t, tmpDir, "repo1")
	repo2 := testutil.MakeGitRepo(t, tmpDir, "repo2")

	// Create a dummy file in repo1 for git grep test
	require.NoError(t, os.WriteFile(filepath.Join(repo1, "hello.txt"), []byte("hello world\n"), 0o644))

	cfg := &config.Config{
		ReposRoot: tmpDir,
		MaxProcs:  1,
	}

	server := iterMcp.NewServer(cfg, iterMcp.WithEnableDo(true))
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)

	t1, t2 := mcpsdk.NewInMemoryTransports()
	ctx := context.Background()

	serverSession, err := server.Connect(ctx, t1, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()

	clientSession, err := client.Connect(ctx, t2, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	t.Run("list tools with do enabled", func(t *testing.T) {
		res, err := clientSession.ListTools(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, res)

		var toolNames []string
		for _, tool := range res.Tools {
			toolNames = append(toolNames, tool.Name)
		}

		assert.Contains(t, toolNames, "git_iter_list")
		assert.Contains(t, toolNames, "git_iter_do")
		assert.Contains(t, toolNames, "git_iter_grep")
		assert.Contains(t, toolNames, "git_iter_ls_files")
		assert.Contains(t, toolNames, "git_iter_status")
		assert.Contains(t, toolNames, "git_iter_diff")
	})

	t.Run("list tools with do disabled", func(t *testing.T) {
		defaultServer := iterMcp.NewServer(cfg)
		st1, ct1 := mcpsdk.NewInMemoryTransports()

		sSession, err := defaultServer.Connect(ctx, st1, nil)
		require.NoError(t, err)
		defer func() { _ = sSession.Close() }()

		defaultClient := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client-2", Version: "v1.0.0"}, nil)
		cSession, err := defaultClient.Connect(ctx, ct1, nil)
		require.NoError(t, err)
		defer func() { _ = cSession.Close() }()

		res, err := cSession.ListTools(ctx, nil)
		require.NoError(t, err)

		var toolNames []string
		for _, tool := range res.Tools {
			toolNames = append(toolNames, tool.Name)
		}

		assert.Contains(t, toolNames, "git_iter_list")
		assert.NotContains(t, toolNames, "git_iter_do")
	})

	tests := []struct {
		name      string
		toolName  string
		arguments map[string]any
		verify    func(t *testing.T, res *mcpsdk.CallToolResult)
	}{
		{
			name:     "call git_iter_list",
			toolName: "git_iter_list",
			arguments: map[string]any{
				"patterns": []string{"repo1"},
			},
			verify: func(t *testing.T, res *mcpsdk.CallToolResult) {
				require.False(t, res.IsError)

				data, err := json.Marshal(res.StructuredContent)
				require.NoError(t, err)

				var out iterMcp.ListToolsOutput
				require.NoError(t, json.Unmarshal(data, &out))
				require.Len(t, out.Repositories, 1)
				assert.Equal(t, repo1, out.Repositories[0].RepoAbsPath)
			},
		},
		{
			name:     "call git_iter_list all repos",
			toolName: "git_iter_list",
			arguments: map[string]any{
				"sort": true,
			},
			verify: func(t *testing.T, res *mcpsdk.CallToolResult) {
				require.False(t, res.IsError)

				data, err := json.Marshal(res.StructuredContent)
				require.NoError(t, err)

				var out iterMcp.ListToolsOutput
				require.NoError(t, json.Unmarshal(data, &out))
				require.Len(t, out.Repositories, 2)
				assert.Equal(t, repo1, out.Repositories[0].RepoAbsPath)
				assert.Equal(t, repo2, out.Repositories[1].RepoAbsPath)
			},
		},
		{
			name:     "call git_iter_list with limit",
			toolName: "git_iter_list",
			arguments: map[string]any{
				"sort":  true,
				"limit": 1,
			},
			verify: func(t *testing.T, res *mcpsdk.CallToolResult) {
				require.False(t, res.IsError)

				data, err := json.Marshal(res.StructuredContent)
				require.NoError(t, err)

				var out iterMcp.ListToolsOutput
				require.NoError(t, json.Unmarshal(data, &out))
				require.Len(t, out.Repositories, 1)
				assert.Equal(t, repo1, out.Repositories[0].RepoAbsPath)
			},
		},
		{
			name:     "call git_iter_do",
			toolName: "git_iter_do",
			arguments: map[string]any{
				"command":  []string{"echo", "hi"},
				"patterns": []string{"repo1"},
			},
			verify: func(t *testing.T, res *mcpsdk.CallToolResult) {
				require.False(t, res.IsError)

				data, err := json.Marshal(res.StructuredContent)
				require.NoError(t, err)

				var out iterMcp.DoToolOutput
				require.NoError(t, json.Unmarshal(data, &out))
				require.Len(t, out.Results, 1)
				assert.Equal(t, repo1, out.Results[0].RepoAbsPath)
				assert.Equal(t, 0, out.Results[0].ExitCode)
				assert.Contains(t, out.Results[0].Stdout, "hi")
			},
		},
		{
			name:     "call git_iter_do missing command",
			toolName: "git_iter_do",
			arguments: map[string]any{
				"patterns": []string{"repo1"},
			},
			verify: func(t *testing.T, res *mcpsdk.CallToolResult) {
				assert.True(t, res.IsError)
			},
		},
		{
			name:     "call git_iter_grep",
			toolName: "git_iter_grep",
			arguments: map[string]any{
				"grep_args": []string{"hello"},
				"patterns":  []string{"repo1"},
			},
			verify: func(t *testing.T, res *mcpsdk.CallToolResult) {
				require.False(t, res.IsError)

				data, err := json.Marshal(res.StructuredContent)
				require.NoError(t, err)

				var out iterMcp.GrepToolOutput
				require.NoError(t, json.Unmarshal(data, &out))
				require.Len(t, out.Results, 1)
				assert.Equal(t, repo1, out.Results[0].RepoAbsPath)
			},
		},
		{
			name:     "call git_iter_grep missing grep_args",
			toolName: "git_iter_grep",
			arguments: map[string]any{
				"patterns": []string{"repo1"},
			},
			verify: func(t *testing.T, res *mcpsdk.CallToolResult) {
				assert.True(t, res.IsError)
			},
		},
		{
			name:     "call git_iter_ls_files",
			toolName: "git_iter_ls_files",
			arguments: map[string]any{
				"patterns": []string{"repo1"},
			},
			verify: func(t *testing.T, res *mcpsdk.CallToolResult) {
				require.False(t, res.IsError)

				data, err := json.Marshal(res.StructuredContent)
				require.NoError(t, err)

				var out iterMcp.LsFilesToolOutput
				require.NoError(t, json.Unmarshal(data, &out))
				require.Len(t, out.Results, 1)
				assert.Equal(t, repo1, out.Results[0].RepoAbsPath)
			},
		},
		{
			name:     "call git_iter_status",
			toolName: "git_iter_status",
			arguments: map[string]any{
				"patterns": []string{"repo1"},
			},
			verify: func(t *testing.T, res *mcpsdk.CallToolResult) {
				require.False(t, res.IsError)

				data, err := json.Marshal(res.StructuredContent)
				require.NoError(t, err)

				var out iterMcp.StatusToolOutput
				require.NoError(t, json.Unmarshal(data, &out))
				require.Len(t, out.Results, 1)
				assert.Equal(t, repo1, out.Results[0].RepoAbsPath)
			},
		},
		{
			name:     "call git_iter_diff",
			toolName: "git_iter_diff",
			arguments: map[string]any{
				"patterns": []string{"repo1"},
			},
			verify: func(t *testing.T, res *mcpsdk.CallToolResult) {
				require.False(t, res.IsError)

				data, err := json.Marshal(res.StructuredContent)
				require.NoError(t, err)

				var out iterMcp.DiffToolOutput
				require.NoError(t, json.Unmarshal(data, &out))
				require.Len(t, out.Results, 1)
				assert.Equal(t, repo1, out.Results[0].RepoAbsPath)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
				Name:      tt.toolName,
				Arguments: tt.arguments,
			})
			require.NoError(t, err)
			tt.verify(t, res)
		})
	}
}
