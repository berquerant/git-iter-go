package tool

import (
	"context"
	"fmt"

	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/berquerant/git-iter-go/internal/output"
	"github.com/berquerant/git-iter-go/internal/runner"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// LsFilesToolInput defines parameters for git_iter_ls_files tool.
type LsFilesToolInput struct {
	LsFilesFlags []string `json:"ls_files_flags,omitempty" jsonschema:"arguments passed to git ls-files"`
	Patterns     []string `json:"patterns,omitempty" jsonschema:"repository path filter regular expressions"`
	ReposRoot    string   `json:"repos_root,omitempty" jsonschema:"root directory to scan for repositories"`
	AbsPath      bool     `json:"abs_path,omitempty" jsonschema:"show absolute paths instead of relative"`
	MaxProcs     int      `json:"max_procs,omitempty" jsonschema:"maximum parallel processes (default 1)"`
	Timeout      string   `json:"timeout,omitempty" jsonschema:"per-process execution timeout (e.g. 5s, 1m)"`
	Sort         bool     `json:"sort,omitempty" jsonschema:"sort repository paths in ascending order"`
	SortReverse  bool     `json:"sort_reverse,omitempty" jsonschema:"sort repository paths in descending order"`
	Limit        int      `json:"limit,omitempty" jsonschema:"limit number of processed repositories"`
	RemoteURL    string   `json:"remote_url,omitempty" jsonschema:"filter repositories by origin remote URL regex"`
}

// LsFilesToolOutput defines output for git_iter_ls_files tool.
type LsFilesToolOutput struct {
	Results []output.Result `json:"results" jsonschema:"matching files with prefixed paths"`
}

// RegisterLsFiles registers the git_iter_ls_files tool on the given MCP server.
func RegisterLsFiles(server *mcpsdk.Server, cfg *config.Config) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "git_iter_ls_files",
		Description: "Run git ls-files across matching repositories with prefixed output paths",
	}, func(
		ctx context.Context,
		_ *mcpsdk.CallToolRequest,
		in LsFilesToolInput,
	) (*mcpsdk.CallToolResult, LsFilesToolOutput, error) {
		finder, root, err := buildFinder(cfg, finderParams{
			reposRoot: in.ReposRoot,
			patterns:  in.Patterns,
			sortAsc:   in.Sort,
			sortDesc:  in.SortReverse,
			limit:     in.Limit,
			remoteURL: in.RemoteURL,
		})
		if err != nil {
			return nil, LsFilesToolOutput{}, err
		}

		paths, err := finder.Find(ctx)
		if err != nil {
			return nil, LsFilesToolOutput{}, fmt.Errorf("find repositories: %w", err)
		}

		timeout, err := resolveTimeout(in.Timeout, cfg)
		if err != nil {
			return nil, LsFilesToolOutput{}, err
		}

		g := git.New(cfg.GitCommandOrFallback(), nil)
		gitLsFilesCmd := g.LsFilesCmd(in.LsFilesFlags...)

		tasks := make([]output.Task, len(paths))
		for i, p := range paths {
			tasks[i] = output.Task{
				RepoAbsPath: p,
				ReposRoot:   root,
				Dir:         p,
				Command:     gitLsFilesCmd,
				Runner: &runner.PathPrefixRunner{
					Inner:    runner.ProcessRunner{},
					RepoPath: p,
					AbsPath:  in.AbsPath,
					Root:     root,
				},
			}
		}

		results, err := executeTasksCollect(ctx, tasks, resolveProcs(in.MaxProcs, cfg), timeout)
		if err != nil {
			return nil, LsFilesToolOutput{}, err
		}

		return nil, LsFilesToolOutput{Results: results}, nil
	})
}
