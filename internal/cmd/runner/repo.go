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

func writeRepoPaths(cfg *config.Config, outMode string, paths []string, stdout io.Writer) error {
	switch outMode {
	case outModeMarkdown, outModeMD:
		results := make([]output.RepoResult, len(paths))
		for i, p := range paths {
			results[i] = output.RepoResult{
				RepoAbsPath: p,
				RepoRelPath: output.RelPath(cfg.ReposRoot, p),
			}
		}
		return output.WriteListMarkdown(stdout, results, cfg.ReposRoot, cfg.MarkdownHeaderLevel)

	case outModeJSON:
		for _, p := range paths {
			if err := output.WriteRepoJSONL(stdout, output.RepoResult{
				RepoAbsPath: p,
				RepoRelPath: output.RelPath(cfg.ReposRoot, p),
			}); err != nil {
				return err
			}
		}
		return nil

	default: // "text"
		for _, p := range paths {
			if _, err := fmt.Fprintln(stdout, p); err != nil {
				return err
			}
		}
		return nil
	}
}

// RunList executes the "list" subcommand logic.
// finder is expected to already be filtered by ListArgs.Patterns.
func RunList(
	ctx context.Context,
	cfg *config.Config,
	outMode string,
	_ parse.ListArgs,
	finder repo.Finder,
	stdout io.Writer,
) error {
	paths, err := finder.Find(ctx)
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}
	return writeRepoPaths(cfg, outMode, paths, stdout)
}

// RunGet executes the "get" subcommand logic.
// If a.URI is empty, it lists discovered repositories via finder.
// If a.URI is non-empty, it checks whether the corresponding local repository
// exists under cfg.ReposRoot; if not, it clones a.URI into that directory using r
// (routing clone output to stderr) and writes the repository path to stdout.
func RunGet(
	ctx context.Context,
	cfg *config.Config,
	outMode string,
	a parse.GetArgs,
	finder repo.Finder,
	r iterRunner.Runner,
	stdout, stderr io.Writer,
) error {
	if a.URI == "" {
		paths, err := finder.Find(ctx)
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}
		return writeRepoPaths(cfg, outMode, paths, stdout)
	}

	runCtx := ctx
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	g := git.New(cfg.GitCommandOrFallback(), r)
	targetDir, err := g.EnsureCloned(runCtx, cfg.ReposRoot, a.URI, stderr)
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}

	return writeRepoPaths(cfg, outMode, []string{targetDir}, stdout)
}


