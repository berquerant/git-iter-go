package tool

import (
	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/berquerant/git-iter-go/internal/output"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// DiffToolInput defines parameters for git_iter_diff tool.
type DiffToolInput struct {
	DiffFlags   []string `json:"diff_flags,omitempty" jsonschema:"arguments and flags passed to git diff"`
	Interactive bool     `json:"interactive,omitempty" jsonschema:"enable pager and interactive diff (default false)"`
	RepoDiscoveryInput
	RepoStatusFilterInput
	ProcessExecInput
}

// DiffToolOutput defines output for git_iter_diff tool.
type DiffToolOutput struct {
	Results []output.Result `json:"results" jsonschema:"git diff results per repository"`
}

// RegisterDiff registers the git_iter_diff tool on the given MCP server.
func RegisterDiff(server *mcpsdk.Server, cfg *config.Config) {
	registerGitCmdTool(
		server,
		cfg,
		"git_iter_diff",
		"Run git diff in each matching repository and return exit codes, stdout, and stderr",
		func(in DiffToolInput) gitCommandToolParams {
			return gitCommandToolParams{
				Discovery: in.RepoDiscoveryInput,
				Status:    in.RepoStatusFilterInput,
				Exec:      in.ProcessExecInput,
				commandFunc: func(g *git.Command) []string {
					return g.DiffCmd(in.Interactive, in.DiffFlags...)
				},
			}
		},
		func(results []output.Result) DiffToolOutput {
			return DiffToolOutput{Results: results}
		},
	)
}
