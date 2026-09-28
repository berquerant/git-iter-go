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
	Patterns             []string `json:"patterns,omitempty" jsonschema:"repository path filter regular expressions"`
	ReposRoot            string   `json:"repos_root,omitempty" jsonschema:"root directory to scan for repositories"`
	Sort                 bool     `json:"sort,omitempty" jsonschema:"sort repository paths in ascending order"`
	SortReverse          bool     `json:"sort_reverse,omitempty" jsonschema:"sort repository paths in descending order"`
	Limit                int      `json:"limit,omitempty" jsonschema:"limit number of processed repositories"`
	DefaultBranchOnly    bool     `json:"default_branch_only,omitempty" jsonschema:"only include repositories whose current branch is the default branch"`
	NotDefaultBranchOnly bool     `json:"not_default_branch_only,omitempty" jsonschema:"only include repositories whose current branch is not the default branch"`
	Branch               string   `json:"branch,omitempty" jsonschema:"filter repositories by current branch regex"`
	RemoteURL            string   `json:"remote_url,omitempty" jsonschema:"filter repositories by origin remote URL regex"`
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
			reposRoot:            in.ReposRoot,
			patterns:             in.Patterns,
			sortAsc:              in.Sort,
			sortDesc:             in.SortReverse,
			limit:                in.Limit,
			defaultBranchOnly:    in.DefaultBranchOnly,
			notDefaultBranchOnly: in.NotDefaultBranchOnly,
			branch:               in.Branch,
			remoteURL:            in.RemoteURL,
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
