package tool_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/mcp/tool"
	"github.com/berquerant/git-iter-go/testutil"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterGrep(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repo1 := testutil.MakeGitRepo(t, tmpDir, "repo1")
	require.NoError(t, os.WriteFile(filepath.Join(repo1, "hello.txt"), []byte("hello world\n"), 0o644))

	cfg := &config.Config{
		ReposRoot: tmpDir,
	}

	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "test-server", Version: "v1.0.0"}, nil)
	tool.RegisterGrep(server, cfg)

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)
	t1, t2 := mcpsdk.NewInMemoryTransports()
	ctx := context.Background()

	serverSession, err := server.Connect(ctx, t1, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()

	clientSession, err := client.Connect(ctx, t2, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	t.Run("call grep success", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name: "git_iter_grep",
			Arguments: map[string]any{
				"grep_args": []string{"hello"},
				"patterns":  []string{"repo1"},
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)

		data, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out tool.GrepToolOutput
		require.NoError(t, json.Unmarshal(data, &out))
		require.Len(t, out.Results, 1)
		assert.Equal(t, repo1, out.Results[0].RepoAbsPath)
	})

	t.Run("call grep missing args", func(t *testing.T) {
		res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
			Name: "git_iter_grep",
			Arguments: map[string]any{
				"patterns": []string{"repo1"},
			},
		})
		require.NoError(t, err)
		assert.True(t, res.IsError)
	})
}
