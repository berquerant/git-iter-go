package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/berquerant/git-iter-go/internal/repo"
	iterRunner "github.com/berquerant/git-iter-go/internal/runner"
)

func findWorktreePaths(ctx context.Context, wtFinder repo.Finder) ([]string, error) {
	paths, err := wtFinder.Find(ctx)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return paths, nil
}

// RunWorktreeList executes the "worktree list" subcommand logic.
func RunWorktreeList(
	ctx context.Context,
	cfg *config.Config,
	outMode string,
	a parse.WorktreeListArgs,
	wtFinder repo.Finder,
	stdout io.Writer,
) error {
	wtRoot := cfg.WorktreeRootOrFallback()
	if wtRoot == "" {
		return fmt.Errorf("worktree list: worktree root is not configured")
	}

	wtPaths, err := findWorktreePaths(ctx, wtFinder)
	if err != nil {
		return fmt.Errorf("worktree list: %w", err)
	}

	if a.Repo {
		mainRepos, err := git.ResolveUniqueMainRepos(cfg.ReposRoot, wtRoot, wtPaths)
		if err != nil {
			return fmt.Errorf("worktree list: %w", err)
		}
		return writeRepoPaths(cfg, outMode, mainRepos, stdout)
	}

	return writeRepoPathsWithRoot(wtRoot, cfg.MarkdownHeaderLevel, outMode, wtPaths, stdout)
}

func resolveRepoWorktree(cfg *config.Config, repoDir, branch string) (string, string, string, error) {
	wtRoot := cfg.WorktreeRootOrFallback()
	repoRoot, err := git.FindRepoRoot(repoDir)
	if err != nil {
		return "", "", "", err
	}
	wtPath, err := git.ResolveWorktreePath(cfg.ReposRoot, wtRoot, repoRoot, branch)
	if err != nil {
		return "", "", "", err
	}
	return wtRoot, repoRoot, wtPath, nil
}

func withConfigTimeout(ctx context.Context, cfg *config.Config) (context.Context, context.CancelFunc) {
	if cfg != nil && cfg.Timeout > 0 {
		return context.WithTimeout(ctx, cfg.Timeout)
	}
	return ctx, func() {}
}

// RunWorktreeAdd executes the "worktree add" subcommand logic for the repository at repoDir.
func RunWorktreeAdd(
	ctx context.Context,
	cfg *config.Config,
	outMode string,
	a parse.WorktreeAddArgs,
	repoDir string,
	r iterRunner.Runner,
	stdout, stderr io.Writer,
) error {
	wtRoot, repoRoot, wtPath, err := resolveRepoWorktree(cfg, repoDir, a.TargetBranch())
	if err != nil {
		return fmt.Errorf("worktree add: %w", err)
	}

	runCtx, cancel := withConfigTimeout(ctx, cfg)
	defer cancel()

	g := git.New(cfg.GitCommandOrFallback(), r)
	if err := g.WorktreeAdd(runCtx, repoRoot, wtPath, a.Branch, a.NewBranch, stderr); err != nil {
		return fmt.Errorf("worktree add: %w", err)
	}

	return writeRepoPathsWithRoot(wtRoot, cfg.MarkdownHeaderLevel, outMode, []string{wtPath}, stdout)
}

// RunWorktreeRemove executes the "worktree remove" subcommand logic for the repository at repoDir.
func RunWorktreeRemove(
	ctx context.Context,
	cfg *config.Config,
	a parse.WorktreeRemoveArgs,
	repoDir string,
	r iterRunner.Runner,
	stderr io.Writer,
) error {
	_, repoRoot, wtPath, err := resolveRepoWorktree(cfg, repoDir, a.Branch)
	if err != nil {
		return fmt.Errorf("worktree remove: %w", err)
	}

	runCtx, cancel := withConfigTimeout(ctx, cfg)
	defer cancel()

	g := git.New(cfg.GitCommandOrFallback(), r)
	if err := g.WorktreeRemove(runCtx, repoRoot, wtPath, a.Force, a.NoForce, stderr); err != nil {
		return fmt.Errorf("worktree remove: %w", err)
	}
	return nil
}

// RunWorktreePrune executes the "worktree prune" subcommand logic.
func RunWorktreePrune(
	ctx context.Context,
	cfg *config.Config,
	outMode string,
	a parse.WorktreePruneArgs,
	repoDir string,
	wtFinder repo.Finder,
	r iterRunner.Runner,
	stdout, stderr io.Writer,
) error {
	wtRoot := cfg.WorktreeRootOrFallback()
	if wtRoot == "" {
		return fmt.Errorf("worktree prune: worktree root is not configured")
	}

	var paths []string
	if a.All {
		wtPaths, err := findWorktreePaths(ctx, wtFinder)
		if err != nil {
			return fmt.Errorf("worktree prune: %w", err)
		}
		paths, err = git.ResolveUniqueMainRepos(cfg.ReposRoot, wtRoot, wtPaths)
		if err != nil {
			return fmt.Errorf("worktree prune: %w", err)
		}
	} else {
		repoRoot, err := git.FindRepoRoot(repoDir)
		if err != nil {
			return fmt.Errorf("worktree prune: %w", err)
		}
		paths = []string{repoRoot}
	}

	g := git.New(cfg.GitCommandOrFallback(), r)
	cmd := g.WorktreePruneCmd(a.DryRun)

	return runCommandOnPaths(ctx, cfg, outMode, "worktree prune", paths, cmd, r, stdout, stderr)
}
