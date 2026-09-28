package tool

import (
	"context"
	"fmt"

	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/berquerant/git-iter-go/internal/output"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type gitCommandToolParams struct {
	Discovery   RepoDiscoveryInput
	Status      RepoStatusFilterInput
	Exec        ProcessExecInput
	commandFunc func(g *git.Command) []string
}

func executeGitCommandTool(ctx context.Context, cfg *config.Config, p gitCommandToolParams) ([]output.Result, error) {
	finder, root, err := buildFinder(cfg, finderParams{
		Discovery: p.Discovery,
		Status:    p.Status,
	})
	if err != nil {
		return nil, err
	}

	paths, err := finder.Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("find repositories: %w", err)
	}

	timeout, err := resolveTimeout(p.Exec.Timeout, cfg)
	if err != nil {
		return nil, err
	}

	gitCmd := ""
	if cfg != nil {
		gitCmd = cfg.GitCommandOrFallback()
	}
	g := git.New(gitCmd, nil)
	cmd := p.commandFunc(g)

	tasks := buildProcessTasks(paths, root, cmd)
	return executeTasksCollect(ctx, tasks, resolveProcs(p.Exec.MaxProcs, cfg), timeout)
}

func registerGitCmdTool[I any, O any](
	server *mcpsdk.Server,
	cfg *config.Config,
	name string,
	description string,
	toParams func(in I) gitCommandToolParams,
	toOutput func(results []output.Result) O,
) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        name,
		Description: description,
	}, func(
		ctx context.Context,
		_ *mcpsdk.CallToolRequest,
		in I,
	) (*mcpsdk.CallToolResult, O, error) {
		results, err := executeGitCommandTool(ctx, cfg, toParams(in))
		if err != nil {
			var zero O
			return nil, zero, err
		}
		return nil, toOutput(results), nil
	})
}
