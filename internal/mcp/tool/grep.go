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

// GrepToolInput defines parameters for git_iter_grep tool.
type GrepToolInput struct {
	GrepArgs []string `json:"grep_args" jsonschema:"arguments passed to git grep (patterns, flags)"`
	AbsPath  bool     `json:"abs_path,omitempty" jsonschema:"show absolute paths instead of relative in grep results"`
	RepoDiscoveryInput
	ProcessExecInput
}

// GrepToolOutput defines output for git_iter_grep tool.
type GrepToolOutput struct {
	Results []output.Result `json:"results" jsonschema:"grep results per repository"`
}

// RegisterGrep registers the git_iter_grep tool on the given MCP server.
func RegisterGrep(server *mcpsdk.Server, cfg *config.Config) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "git_iter_grep",
		Description: "Run git grep across matching repositories with prefixed output paths",
	}, func(
		ctx context.Context,
		_ *mcpsdk.CallToolRequest,
		in GrepToolInput,
	) (*mcpsdk.CallToolResult, GrepToolOutput, error) {
		if len(in.GrepArgs) == 0 {
			return nil, GrepToolOutput{}, fmt.Errorf("grep_args is required")
		}

		finder, root, err := buildFinder(cfg, finderParams{
			Discovery: in.RepoDiscoveryInput,
		})
		if err != nil {
			return nil, GrepToolOutput{}, err
		}

		paths, err := finder.Find(ctx)
		if err != nil {
			return nil, GrepToolOutput{}, fmt.Errorf("find repositories: %w", err)
		}

		timeout, err := resolveTimeout(in.Timeout, cfg)
		if err != nil {
			return nil, GrepToolOutput{}, err
		}

		g := git.New(cfg.GitCommandOrFallback(), nil)
		gitGrepCmd := g.GrepCmd(in.GrepArgs...)

		tasks := make([]output.Task, len(paths))
		for i, p := range paths {
			tasks[i] = output.Task{
				RepoAbsPath: p,
				ReposRoot:   root,
				Dir:         p,
				Command:     gitGrepCmd,
				Runner: &runner.GrepRunner{
					Inner:    runner.ProcessRunner{},
					RepoPath: p,
					AbsPath:  in.AbsPath,
					Root:     root,
				},
			}
		}

		results, err := executeTasksCollect(ctx, tasks, resolveProcs(in.MaxProcs, cfg), timeout)
		if err != nil {
			return nil, GrepToolOutput{}, err
		}

		return nil, GrepToolOutput{Results: results}, nil
	})
}
