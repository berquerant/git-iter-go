package tool

import (
	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/berquerant/git-iter-go/internal/output"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// StatusToolInput defines parameters for git_iter_status tool.
type StatusToolInput struct {
	StatusFlags []string `json:"status_flags,omitempty" jsonschema:"arguments and flags passed to git status"`
	RepoDiscoveryInput
	RepoStatusFilterInput
	ProcessExecInput
}

// StatusToolOutput defines output for git_iter_status tool.
type StatusToolOutput struct {
	Results []output.Result `json:"results" jsonschema:"git status results per repository"`
}

// RegisterStatus registers the git_iter_status tool on the given MCP server.
func RegisterStatus(server *mcpsdk.Server, cfg *config.Config) {
	registerGitCmdTool(
		server,
		cfg,
		"git_iter_status",
		"Run git status in each matching repository and return exit codes, stdout, and stderr",
		func(in StatusToolInput) gitCommandToolParams {
			return gitCommandToolParams{
				Discovery: in.RepoDiscoveryInput,
				Status:    in.RepoStatusFilterInput,
				Exec:      in.ProcessExecInput,
				commandFunc: func(g *git.Command) []string {
					return g.StatusCmd(in.StatusFlags...)
				},
			}
		},
		func(results []output.Result) StatusToolOutput {
			return StatusToolOutput{Results: results}
		},
	)
}
