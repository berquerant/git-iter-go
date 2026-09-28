package tool

import (
	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/berquerant/git-iter-go/internal/output"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// StatusToolInput defines parameters for git_iter_status tool.
type StatusToolInput struct {
	StatusFlags          []string `json:"status_flags,omitempty" jsonschema:"arguments and flags passed to git status"`
	Patterns             []string `json:"patterns,omitempty" jsonschema:"repository path filter regular expressions"`
	ReposRoot            string   `json:"repos_root,omitempty" jsonschema:"root directory to scan for repositories"`
	MaxProcs             int      `json:"max_procs,omitempty" jsonschema:"maximum parallel processes (default 1)"`
	Timeout              string   `json:"timeout,omitempty" jsonschema:"per-process execution timeout (e.g. 5s, 1m)"`
	Sort                 bool     `json:"sort,omitempty" jsonschema:"sort repository paths in ascending order"`
	SortReverse          bool     `json:"sort_reverse,omitempty" jsonschema:"sort repository paths in descending order"`
	Limit                int      `json:"limit,omitempty" jsonschema:"limit number of processed repositories"`
	DirtyOnly            bool     `json:"dirty_only,omitempty" jsonschema:"only run in repositories with uncommitted changes"`
	CleanOnly            bool     `json:"clean_only,omitempty" jsonschema:"only run in repositories without uncommitted changes"`
	DefaultBranchOnly    bool     `json:"default_branch_only,omitempty" jsonschema:"only include repositories whose current branch is the default branch"`
	NotDefaultBranchOnly bool     `json:"not_default_branch_only,omitempty" jsonschema:"only include repositories whose current branch is not the default branch"`
	Branch               string   `json:"branch,omitempty" jsonschema:"filter repositories by current branch regex"`
	RemoteURL            string   `json:"remote_url,omitempty" jsonschema:"filter repositories by origin remote URL regex"`
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
				reposRoot:            in.ReposRoot,
				patterns:             in.Patterns,
				sortAsc:              in.Sort,
				sortDesc:             in.SortReverse,
				limit:                in.Limit,
				dirtyOnly:            in.DirtyOnly,
				cleanOnly:            in.CleanOnly,
				defaultBranchOnly:    in.DefaultBranchOnly,
				notDefaultBranchOnly: in.NotDefaultBranchOnly,
				branch:               in.Branch,
				remoteURL:            in.RemoteURL,
				timeout:              in.Timeout,
				maxProcs:             in.MaxProcs,
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
