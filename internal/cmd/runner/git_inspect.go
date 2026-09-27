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

// RunLsFiles executes the "ls-files" subcommand logic.
// finder is expected to already be filtered by LsFilesArgs.Patterns.
// showAbsPath overrides cfg.AbsPath if true.
// innerRunner executes git commands.
func RunLsFiles(
	ctx context.Context,
	cfg *config.Config,
	outMode string,
	a parse.LsFilesArgs,
	finder repo.Finder,
	innerRunner iterRunner.Runner,
	showAbsPath bool,
	stdout, stderr io.Writer,
) error {
	paths, err := finder.Find(ctx)
	if err != nil {
		return fmt.Errorf("ls-files: %w", err)
	}

	useAbsPath := showAbsPath || (cfg != nil && cfg.AbsPath)

	g := git.New(cfg.GitCommandOrFallback(), innerRunner)
	gitLsFilesCmd := g.LsFilesCmd(a.LsFilesFlags...)

	switch outMode {
	case outModeMarkdown, outModeMD:
		tasks := buildOutputTasks(paths, cfg.ReposRoot, gitLsFilesCmd, innerRunner)
		var title string
		if cfg.ReposRoot != "" {
			title = fmt.Sprintf("ls-files: root = %s", cfg.ReposRoot)
		} else {
			title = "ls-files"
		}
		return (&output.LsFilesMarkdownExecutor{
			Title:       title,
			MaxProcs:    cfg.MaxProcs,
			FailFast:    cfg.FailFast,
			Timeout:     cfg.Timeout,
			HeaderLevel: cfg.MarkdownHeaderLevel,
			Out:         stdout,
		}).Run(ctx, tasks)

	case outModeJSON:
		tasks := buildOutputTasks(paths, cfg.ReposRoot, gitLsFilesCmd, innerRunner)
		return (&output.FileJSONLExecutor{
			MaxProcs: cfg.MaxProcs,
			FailFast: cfg.FailFast,
			Timeout:  cfg.Timeout,
			Out:      stdout,
		}).Run(ctx, tasks)

	default: // "text"
		return runPathPrefixText(
			ctx,
			cfg.MaxProcs,
			cfg.FailFast,
			useAbsPath,
			cfg.ReposRoot,
			paths,
			gitLsFilesCmd,
			innerRunner,
			stdout,
			stderr,
		)
	}
}

// RunStatus executes the "status" subcommand logic.
// finder is expected to already be filtered by StatusArgs.Patterns and dirty/clean filter.
// r is the runner used to execute git status in each repository.
func RunStatus(
	ctx context.Context,
	cfg *config.Config,
	outMode string,
	a parse.StatusArgs,
	finder repo.Finder,
	r iterRunner.Runner,
	stdout, stderr io.Writer,
) error {
	paths, err := finder.Find(ctx)
	if err != nil {
		return fmt.Errorf("status: %w", err)
	}

	g := git.New(cfg.GitCommandOrFallback(), r)
	cmd := g.StatusCmd(a.StatusFlags...)

	return runCommandOnPaths(ctx, cfg, outMode, "status", paths, cmd, r, stdout, stderr)
}

// RunDiff executes the "diff" subcommand logic.
// finder is expected to already be filtered by DiffArgs.Patterns and dirty/clean filter.
// r is the runner used to execute git diff in each repository.
func RunDiff(
	ctx context.Context,
	cfg *config.Config,
	outMode string,
	a parse.DiffArgs,
	finder repo.Finder,
	r iterRunner.Runner,
	stdout, stderr io.Writer,
) error {
	paths, err := finder.Find(ctx)
	if err != nil {
		return fmt.Errorf("diff: %w", err)
	}

	g := git.New(cfg.GitCommandOrFallback(), r)
	cmd := g.DiffCmd(a.Interactive, a.DiffFlags...)

	return runCommandOnPaths(ctx, cfg, outMode, "diff", paths, cmd, r, stdout, stderr)
}
