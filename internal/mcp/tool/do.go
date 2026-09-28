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
	Command []string `json:"command" jsonschema:"command and arguments to execute in each repository"`
	RepoDiscoveryInput
	ProcessExecInput
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
			Discovery: in.RepoDiscoveryInput,
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
