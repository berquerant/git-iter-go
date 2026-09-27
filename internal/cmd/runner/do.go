package runner

import (
	"context"
	"fmt"
	"io"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/repo"
	iterRunner "github.com/berquerant/git-iter-go/internal/runner"
)

// RunDo executes the "do" subcommand logic.
// finder is expected to already be filtered by DoArgs.Patterns.
// r is the runner used to execute the command in each repository.
func RunDo(
	ctx context.Context,
	cfg *config.Config,
	outMode string,
	a parse.DoArgs,
	finder repo.Finder,
	r iterRunner.Runner,
	stdout, stderr io.Writer,
) error {
	paths, err := finder.Find(ctx)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	return runCommandOnPaths(ctx, cfg, outMode, "do", paths, a.Command, r, stdout, stderr)
}

// RunRead executes the "read" subcommand logic.
// finder supplies the repository paths (typically a StdinFinder in production).
// r is the runner used to execute the command in each repository.
func RunRead(
	ctx context.Context,
	cfg *config.Config,
	outMode string,
	a parse.ReadArgs,
	finder repo.Finder,
	r iterRunner.Runner,
	stdout, stderr io.Writer,
) error {
	paths, err := finder.Find(ctx)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	return runCommandOnPaths(ctx, cfg, outMode, "read", paths, a.Command, r, stdout, stderr)
}
