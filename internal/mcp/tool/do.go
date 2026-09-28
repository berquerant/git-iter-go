package tool

import (
	"context"
	"fmt"

	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/output"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// DoToolInput defines parameters for git_iter_do tool.
type DoToolInput struct {
	Command              []string `json:"command" jsonschema:"command and arguments to execute in each repository"`
	Patterns             []string `json:"patterns,omitempty" jsonschema:"repository path filter regular expressions"`
	ReposRoot            string   `json:"repos_root,omitempty" jsonschema:"root directory to scan for repositories"`
	MaxProcs             int      `json:"max_procs,omitempty" jsonschema:"maximum parallel processes (default 1)"`
	Timeout              string   `json:"timeout,omitempty" jsonschema:"per-process execution timeout (e.g. 5s, 1m)"`
	Sort                 bool     `json:"sort,omitempty" jsonschema:"sort repository paths in ascending order"`
	SortReverse          bool     `json:"sort_reverse,omitempty" jsonschema:"sort repository paths in descending order"`
	Limit                int      `json:"limit,omitempty" jsonschema:"limit number of processed repositories"`
	DefaultBranchOnly    bool     `json:"default_branch_only,omitempty" jsonschema:"only include repositories whose current branch is the default branch"`
	NotDefaultBranchOnly bool     `json:"not_default_branch_only,omitempty" jsonschema:"only include repositories whose current branch is not the default branch"`
	Branch               string   `json:"branch,omitempty" jsonschema:"filter repositories by current branch regex"`
	RemoteURL            string   `json:"remote_url,omitempty" jsonschema:"filter repositories by origin remote URL regex"`
}

// DoToolOutput defines output for git_iter_do tool.
type DoToolOutput struct {
	Results []output.Result `json:"results" jsonschema:"execution results per repository"`
}

// RegisterDo registers the git_iter_do tool on the given MCP server.
func RegisterDo(server *mcpsdk.Server, cfg *config.Config) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "git_iter_do",
		Description: "Run an arbitrary command in each matching repository and return exit codes, stdout, and stderr",
	}, func(
		ctx context.Context,
		_ *mcpsdk.CallToolRequest,
		in DoToolInput,
	) (*mcpsdk.CallToolResult, DoToolOutput, error) {
		if len(in.Command) == 0 {
			return nil, DoToolOutput{}, fmt.Errorf("command is required")
		}

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
			return nil, DoToolOutput{}, err
		}

		paths, err := finder.Find(ctx)
		if err != nil {
			return nil, DoToolOutput{}, fmt.Errorf("find repositories: %w", err)
		}

		timeout, err := resolveTimeout(in.Timeout, cfg)
		if err != nil {
			return nil, DoToolOutput{}, err
		}

		tasks := buildProcessTasks(paths, root, in.Command)
		results, err := executeTasksCollect(ctx, tasks, resolveProcs(in.MaxProcs, cfg), timeout)
		if err != nil {
			return nil, DoToolOutput{}, err
		}

		return nil, DoToolOutput{Results: results}, nil
	})
}
