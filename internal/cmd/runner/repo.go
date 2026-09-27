package runner

import (
	"context"
	"fmt"
	"io"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/output"
	"github.com/berquerant/git-iter-go/internal/repo"
)

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
