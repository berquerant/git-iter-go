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
	"github.com/berquerant/git-iter-go/internal/output"
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

func buildWorktreeResults(reposRoot, wtRoot string, wtPaths []string) []output.WorktreeResult {
	results := make([]output.WorktreeResult, len(wtPaths))
	for i, wtPath := range wtPaths {
		mainRepo, branch, _ := git.ResolveMainRepoFromWorktree(reposRoot, wtRoot, wtPath)
		var repoRel string
		if mainRepo != "" && reposRoot != "" {
			repoRel = output.RelPath(reposRoot, mainRepo)
		}
		results[i] = output.WorktreeResult{
			WorktreeAbsPath: wtPath,
			WorktreeRelPath: output.RelPath(wtRoot, wtPath),
			RepoAbsPath:     mainRepo,
			RepoRelPath:     repoRel,
			Branch:          branch,
		}
	}
	return results
}

func writeWorktreeResults(cfg *config.Config, wtRoot, outMode string, results []output.WorktreeResult, stdout io.Writer) error {
	switch outMode {
	case outModeMarkdown, outModeMD:
		return output.WriteWorktreeListMarkdown(stdout, results, wtRoot, cfg.MarkdownHeaderLevel)
	case outModeJSON:
		for _, r := range results {
			if err := output.WriteWorktreeJSONL(stdout, r); err != nil {
				return err
			}
		}
		return nil
	default: // "text"
		for _, r := range results {
			if _, err := fmt.Fprintln(stdout, r.WorktreeAbsPath); err != nil {
				return err
			}
		}
		return nil
	}
}

func filterWorktreeResultsByRepo(results []output.WorktreeResult, targetRepo string) []output.WorktreeResult {
	if targetRepo == "" {
		return results
	}
	filtered := make([]output.WorktreeResult, 0, len(results))
	for _, r := range results {
		if r.RepoAbsPath == targetRepo {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func extractUniqueMainRepos(results []output.WorktreeResult) []string {
	seen := make(map[string]struct{})
	var repos []string
	for _, r := range results {
		if r.RepoAbsPath == "" {
			continue
		}
		if _, ok := seen[r.RepoAbsPath]; !ok {
			seen[r.RepoAbsPath] = struct{}{}
			repos = append(repos, r.RepoAbsPath)
		}
	}
	return repos
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

	var targetMainRepo string
	if !a.All && a.RepoDir != "" {
		mainRepo, err := git.FindMainRepo(cfg.ReposRoot, wtRoot, a.RepoDir)
		if err != nil {
			return fmt.Errorf("worktree list: %w", err)
		}
		targetMainRepo = mainRepo
	}

	wtPaths, err := findWorktreePaths(ctx, wtFinder)
	if err != nil {
		return fmt.Errorf("worktree list: %w", err)
	}

	wtResults := buildWorktreeResults(cfg.ReposRoot, wtRoot, wtPaths)
	wtResults = filterWorktreeResultsByRepo(wtResults, targetMainRepo)

	if a.Repo {
		mainRepos := extractUniqueMainRepos(wtResults)
		return writeRepoPaths(cfg, outMode, mainRepos, stdout)
	}

	return writeWorktreeResults(cfg, wtRoot, outMode, wtResults, stdout)
}

func resolveRepoWorktree(cfg *config.Config, repoDir, branch string) (string, string, string, error) {
	wtRoot := cfg.WorktreeRootOrFallback()
	repoRoot, err := git.FindMainRepo(cfg.ReposRoot, wtRoot, repoDir)
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

	res := output.WorktreeResult{
		WorktreeAbsPath: wtPath,
		WorktreeRelPath: output.RelPath(wtRoot, wtPath),
		RepoAbsPath:     repoRoot,
		RepoRelPath:     output.RelPath(cfg.ReposRoot, repoRoot),
		Branch:          a.TargetBranch(),
	}
	return writeWorktreeResults(cfg, wtRoot, outMode, []output.WorktreeResult{res}, stdout)
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
	if a.DeleteBranch || a.ForceDeleteBranch {
		force := a.ForceDeleteBranch || a.Force
		if err := g.BranchDelete(runCtx, repoRoot, a.Branch, force, stderr); err != nil {
			return fmt.Errorf("worktree remove: branch delete: %w", err)
		}
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
		repoRoot, err := git.FindMainRepo(cfg.ReposRoot, wtRoot, repoDir)
		if err != nil {
			return fmt.Errorf("worktree prune: %w", err)
		}
		paths = []string{repoRoot}
	}

	g := git.New(cfg.GitCommandOrFallback(), r)
	cmd := g.WorktreePruneCmd(a.DryRun)

	return runCommandOnPaths(ctx, cfg, outMode, "worktree prune", paths, cmd, r, stdout, stderr)
}
