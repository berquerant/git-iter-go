package tool

import (
	"context"
	"fmt"

	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/output"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListToolsInput defines parameters for git_iter_list tool.
type ListToolsInput struct {
	RepoDiscoveryInput
}

// ListToolsOutput defines output for git_iter_list tool.
type ListToolsOutput struct {
	Repositories []output.RepoResult `json:"repositories" jsonschema:"discovered repositories"`
}

func listRepositories(ctx context.Context, cfg *config.Config, discovery RepoDiscoveryInput) ([]output.RepoResult, error) {
	finder, root, err := buildFinder(cfg, finderParams{
		Discovery: discovery,
	})
	if err != nil {
		return nil, err
	}

	paths, err := finder.Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("find repositories: %w", err)
	}

	return toRepoResults(root, paths), nil
}

// RegisterList registers the git_iter_list tool on the given MCP server.
func RegisterList(server *mcpsdk.Server, cfg *config.Config) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "git_iter_list",
		Description: "Discover and list git repositories under the configured or specified root directory",
	}, func(
		ctx context.Context,
		_ *mcpsdk.CallToolRequest,
		in ListToolsInput,
	) (*mcpsdk.CallToolResult, ListToolsOutput, error) {
		repos, err := listRepositories(ctx, cfg, in.RepoDiscoveryInput)
		if err != nil {
			return nil, ListToolsOutput{}, err
		}
		return nil, ListToolsOutput{Repositories: repos}, nil
	})
}

