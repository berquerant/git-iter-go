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
		finder, root, err := buildFinder(cfg, finderParams{
			Discovery: in.RepoDiscoveryInput,
		})
		if err != nil {
			return nil, ListToolsOutput{}, err
		}

		paths, err := finder.Find(ctx)
		if err != nil {
			return nil, ListToolsOutput{}, fmt.Errorf("find repositories: %w", err)
		}

		repos := make([]output.RepoResult, len(paths))
		for i, p := range paths {
			repos[i] = output.RepoResult{
				RepoAbsPath: p,
				RepoRelPath: output.RelPath(root, p),
			}
		}

		return nil, ListToolsOutput{Repositories: repos}, nil
	})
}
