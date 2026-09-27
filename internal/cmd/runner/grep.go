package runner

import (
	"context"
	"fmt"
	"io"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/berquerant/git-iter-go/internal/output"
	"github.com/berquerant/git-iter-go/internal/repo"
	iterRunner "github.com/berquerant/git-iter-go/internal/runner"
)

// RunGrep executes the "grep" subcommand logic.
// finder is expected to already be filtered by GrepArgs.Patterns.
// innerRunner is the base runner wrapped by GrepRunner for each repository.
func RunGrep(
	ctx context.Context,
	cfg *config.Config,
	outMode string,
	a parse.GrepArgs,
	finder repo.Finder,
	innerRunner iterRunner.Runner,
	stdout, stderr io.Writer,
) error {
	paths, err := finder.Find(ctx)
	if err != nil {
		return fmt.Errorf("grep: %w", err)
	}

	g := git.New(cfg.GitCommandOrFallback(), innerRunner)
	gitGrepCmd := g.GrepCmd(a.GitGrepArgs...)

	buildGrepTasks := func() []output.Task {
		return buildOutputTasksWithRunner(paths, cfg.ReposRoot, gitGrepCmd, func(p string) iterRunner.Runner {
			return &iterRunner.GrepRunner{
				Inner:    innerRunner,
				RepoPath: p,
				AbsPath:  cfg.AbsPath,
				Root:     cfg.ReposRoot,
			}
		})
	}

	switch outMode {
	case outModeMarkdown, outModeMD:
		tasks := buildGrepTasks()
		title := formatCommandTitle("grep", cfg.ReposRoot, gitGrepCmd)
		return (&output.MarkdownExecutor{
			Title:       title,
			MaxProcs:    cfg.MaxProcs,
			FailFast:    cfg.FailFast,
			Timeout:     cfg.Timeout,
			HeaderLevel: cfg.MarkdownHeaderLevel,
			Out:         stdout,
		}).Run(ctx, tasks)

	case outModeJSON:
		tasks := buildGrepTasks()
		// JSONLExecutor captures exit codes (including git grep no-match exit 1)
		// as data; only write/context errors are propagated.
		return (&output.JSONLExecutor{MaxProcs: cfg.MaxProcs, Out: stdout}).Run(ctx, tasks)

	default: // "text"
		return runPathPrefixText(
			ctx,
			cfg.MaxProcs,
			cfg.FailFast,
			cfg.AbsPath,
			cfg.ReposRoot,
			paths,
			gitGrepCmd,
			innerRunner,
			stdout,
			stderr,
		)
	}
}
