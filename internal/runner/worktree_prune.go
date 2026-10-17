package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"

	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/berquerant/git-iter-go/internal/repo"
)

// WorktreePruneRunner prunes merged worktrees under WorktreeRoot for a repository
// and runs "git worktree prune".
type WorktreePruneRunner struct {
	// GitCommand is the git binary name/path. Defaults to "git" if empty.
	GitCommand string
	// ReposRoot is the root directory of main repositories.
	ReposRoot string
	// WorktreeRoot is the root directory where worktrees are located.
	WorktreeRoot string
	// DryRun controls whether worktree removal and prune run in dry-run mode.
	DryRun bool
	// Inner executes git commands. Defaults to ProcessRunner{} if nil.
	Inner Runner
}

func (p *WorktreePruneRunner) git() *git.Command {
	return git.New(p.GitCommand, p.Inner)
}

func (p *WorktreePruneRunner) findRepoWorktrees(ctx context.Context, dir string) (string, []string, error) {
	relRepo, err := git.ResolveRepoRelPath(p.ReposRoot, dir)
	if err != nil {
		return "", nil, err
	}
	absWorktreeRoot, err := filepath.Abs(p.WorktreeRoot)
	if err != nil {
		return "", nil, fmt.Errorf("resolve worktree-root: %w", err)
	}
	wtRepoDir := filepath.Join(absWorktreeRoot, relRepo)
	finder := &repo.FilesystemFinder{Root: wtRepoDir}
	wtPaths, err := finder.Find(ctx)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return wtRepoDir, nil, nil
		}
		return "", nil, err
	}
	return wtRepoDir, wtPaths, nil
}

func (p *WorktreePruneRunner) resolveWorktreeBranch(ctx context.Context, g *git.Command, wtRepoDir, wtPath string) string {
	if branch, err := g.CurrentBranch(ctx, wtPath); err == nil && branch != "" {
		return branch
	}
	if rel, err := filepath.Rel(wtRepoDir, wtPath); err == nil && rel != "." {
		return filepath.ToSlash(rel)
	}
	return ""
}

// Run discovers worktrees for dir under WorktreeRoot, removes those merged into the
// repository's default branch, and runs "git worktree prune".
func (p *WorktreePruneRunner) Run(ctx context.Context, dir string, _ []string, stdout, stderr io.Writer) error {
	g := p.git()

	defaultBranch, err := g.ResolveDefaultBranch(ctx, dir)
	if err != nil {
		return err
	}

	wtRepoDir, wtPaths, err := p.findRepoWorktrees(ctx, dir)
	if err != nil {
		return err
	}

	for _, wtPath := range wtPaths {
		if err := ctx.Err(); err != nil {
			return err
		}
		branch := p.resolveWorktreeBranch(ctx, g, wtRepoDir, wtPath)
		if branch == "" || branch == defaultBranch {
			continue
		}

		slog.InfoContext(ctx, "process worktree", "repo", dir, "branch", branch, "path", wtPath, "default_branch", defaultBranch)

		merged, err := g.IsMerged(ctx, dir, branch, defaultBranch)
		if err != nil || !merged {
			continue
		}

		slog.InfoContext(ctx, "removing merged worktree", "repo", dir, "branch", branch, "path", wtPath, "dry_run", p.DryRun)
		if p.DryRun {
			continue
		}
		if err := g.WorktreeRemove(ctx, dir, wtPath, true, false, stderr); err != nil {
			return fmt.Errorf("remove merged worktree %s (repo=%s branch=%s): %w", wtPath, dir, branch, err)
		}
	}

	return g.WorktreePrune(ctx, dir, p.DryRun, stdout, stderr)
}
