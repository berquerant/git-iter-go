package runner

import (
	"context"
	"fmt"
	"io"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/berquerant/git-iter-go/internal/repo"
	iterRunner "github.com/berquerant/git-iter-go/internal/runner"
)

// RunPull executes the "pull" subcommand logic.
// finder is expected to already be filtered by PullArgs.Patterns.
// r is the runner used to execute git pull in each repository.
func RunPull(
	ctx context.Context,
	cfg *config.Config,
	outMode string,
	a parse.PullArgs,
	finder repo.Finder,
	r iterRunner.Runner,
	stdout, stderr io.Writer,
) error {
	paths, err := finder.Find(ctx)
	if err != nil {
		return fmt.Errorf("pull: %w", err)
	}

	g := git.New(cfg.GitCommandOrFallback(), r)
	cmd := g.PullCmd(a.PullFlags...)

	return runCommandOnPaths(ctx, cfg, outMode, "pull", paths, cmd, r, stdout, stderr)
}

// RunFetch executes the "fetch" subcommand logic.
// finder is expected to already be filtered by FetchArgs.Patterns.
// r is the runner used to execute git fetch in each repository.
func RunFetch(
	ctx context.Context,
	cfg *config.Config,
	outMode string,
	a parse.FetchArgs,
	finder repo.Finder,
	r iterRunner.Runner,
	stdout, stderr io.Writer,
) error {
	paths, err := finder.Find(ctx)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	g := git.New(cfg.GitCommandOrFallback(), r)
	cmd := g.FetchCmd(a.FetchFlags...)

	return runCommandOnPaths(ctx, cfg, outMode, "fetch", paths, cmd, r, stdout, stderr)
}
