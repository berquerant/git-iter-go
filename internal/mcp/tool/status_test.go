package tool_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/mcp/tool"
	"github.com/berquerant/git-iter-go/testutil"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterStatus(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repo1 := testutil.MakeGitRepo(t, tmpDir, "repo1")

	cfg := &config.Config{
		ReposRoot: tmpDir,
	}

	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "test-server", Version: "v1.0.0"}, nil)
	tool.RegisterStatus(server, cfg)

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)
	t1, t2 := mcpsdk.NewInMemoryTransports()
	ctx := context.Background()

	serverSession, err := server.Connect(ctx, t1, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()

	clientSession, err := client.Connect(ctx, t2, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	res, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
		Name: "git_iter_status",
		Arguments: map[string]any{
			"patterns": []string{"repo1"},
		},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)

	data, err := json.Marshal(res.StructuredContent)
	require.NoError(t, err)

	var out tool.StatusToolOutput
	require.NoError(t, json.Unmarshal(data, &out))
	require.Len(t, out.Results, 1)
	assert.Equal(t, repo1, out.Results[0].RepoAbsPath)
}
