package git

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// FindRepoRoot walks up from dir until it finds a directory containing a ".git" entry.
// Returns the absolute path to the repository root.
func FindRepoRoot(dir string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("empty repository directory")
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve repository directory: %w", err)
	}
	curr := absDir
	for {
		if IsRepo(curr) {
			return curr, nil
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			return "", fmt.Errorf("not a git repository: %q", dir)
		}
		curr = parent
	}
}

// FindMainRepo resolves the root directory of the main git repository for dir.
// If dir is within a linked git worktree, it resolves the main repository.
// If dir is within a normal git repository, it returns that repository's root.
func FindMainRepo(reposRoot, worktreeRoot, dir string) (string, error) {
	repoRoot, err := FindRepoRoot(dir)
	if err != nil {
		return "", err
	}
	if worktreeRoot != "" {
		if mainRepo, _, err := ResolveMainRepoFromWorktree(reposRoot, worktreeRoot, repoRoot); err == nil && mainRepo != "" {
			return mainRepo, nil
		}
	}
	if mainRepo, _, err := FindMainRepoAndBranchFromGit(repoRoot); err == nil && mainRepo != "" {
		if reposRoot == "" {
			return mainRepo, nil
		}
		if absReposRoot, err := filepath.Abs(reposRoot); err == nil && isSubDir(absReposRoot, mainRepo) {
			return mainRepo, nil
		}
	}
	return repoRoot, nil
}

func strictRelPath(base, target string) (string, bool) {
	rel, err := filepath.Rel(base, target)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return rel, true
}

// ResolveRepoRelPath resolves the relative path of the git repository containing repoDir
// with respect to reposRoot (i.e. DIR in repos_root/DIR).
func ResolveRepoRelPath(reposRoot, repoDir string) (string, error) {
	if reposRoot == "" {
		return "", fmt.Errorf("--repos-root (-r) is required")
	}
	absRoot, err := filepath.Abs(reposRoot)
	if err != nil {
		return "", fmt.Errorf("resolve repos-root: %w", err)
	}
	repoRoot, err := FindRepoRoot(repoDir)
	if err != nil {
		return "", err
	}
	rel, ok := strictRelPath(absRoot, repoRoot)
	if !ok {
		return "", fmt.Errorf("repository %q is not under repos-root %q", repoRoot, absRoot)
	}
	return rel, nil
}

func validateBranchName(branch string) error {
	if strings.TrimSpace(branch) == "" {
		return fmt.Errorf("empty branch name")
	}
	if strings.TrimSpace(branch) != branch || strings.HasPrefix(branch, "-") {
		return fmt.Errorf("invalid branch name: %q", branch)
	}
	parts := strings.Split(filepath.ToSlash(branch), "/")
	if slices.ContainsFunc(parts, func(p string) bool {
		t := strings.TrimSpace(p)
		return t == "" || t != p || p == "." || p == ".."
	}) {
		return fmt.Errorf("invalid branch name: %q", branch)
	}
	return nil
}

// ResolveWorktreePath computes worktree_root/DIR/branch for a repository at repoDir under reposRoot.
func ResolveWorktreePath(reposRoot, worktreeRoot, repoDir, branch string) (string, error) {
	if err := validateBranchName(branch); err != nil {
		return "", err
	}
	if worktreeRoot == "" {
		return "", fmt.Errorf("worktree root is required")
	}
	absWorktreeRoot, err := filepath.Abs(worktreeRoot)
	if err != nil {
		return "", fmt.Errorf("resolve worktree-root: %w", err)
	}
	relRepo, err := ResolveRepoRelPath(reposRoot, repoDir)
	if err != nil {
		return "", err
	}
	return filepath.Join(absWorktreeRoot, relRepo, filepath.FromSlash(branch)), nil
}

func isSubDir(parent, sub string) bool {
	rel, err := filepath.Rel(parent, sub)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// FindMainRepoAndBranchFromGit discovers the main repository path and branch name
// by reading the git worktree metadata (.git file and commondir) directly.
func FindMainRepoAndBranchFromGit(wtPath string) (string, string, error) {
	if wtPath == "" {
		return "", "", fmt.Errorf("empty worktree path")
	}
	absWT, err := filepath.Abs(wtPath)
	if err != nil {
		return "", "", fmt.Errorf("resolve worktree path: %w", err)
	}
	gitPath := filepath.Join(absWT, ".git")
	fi, err := os.Stat(gitPath)
	if err != nil {
		return "", "", fmt.Errorf("not a git repository or worktree: %q: %w", absWT, err)
	}

	if fi.IsDir() {
		return "", "", fmt.Errorf("%q is a standalone git repository, not a linked worktree", absWT)
	}

	data, err := os.ReadFile(gitPath)
	if err != nil {
		return "", "", fmt.Errorf("read worktree .git: %w", err)
	}
	line := strings.TrimSpace(string(data))
	if !strings.HasPrefix(line, "gitdir:") {
		return "", "", fmt.Errorf("invalid .git worktree pointer in %q", absWT)
	}
	gitDir := strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(absWT, gitDir)
	}
	gitDir = filepath.Clean(gitDir)

	commonDirBytes, err := os.ReadFile(filepath.Join(gitDir, "commondir"))
	var mainGitDir string
	if err == nil {
		comm := strings.TrimSpace(string(commonDirBytes))
		if !filepath.IsAbs(comm) {
			comm = filepath.Join(gitDir, comm)
		}
		mainGitDir = filepath.Clean(comm)
	} else {
		mainGitDir = gitDir
	}
	mainRepo := filepath.Dir(mainGitDir)
	branch := readBranchFromHeadFile(filepath.Join(gitDir, "HEAD"))

	return mainRepo, branch, nil
}

func readBranchFromHeadFile(headPath string) string {
	data, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}
	head := strings.TrimSpace(string(data))
	if after, ok := strings.CutPrefix(head, "ref: refs/heads/"); ok {
		return after
	}
	return head
}

// ResolveMainRepoFromWorktree resolves the main repository path (repos_root/DIR) and branch name
// from a worktree path (worktree_root/DIR/branch).
func ResolveMainRepoFromWorktree(reposRoot, worktreeRoot, wtPath string) (string, string, error) {
	if worktreeRoot == "" {
		return "", "", fmt.Errorf("worktree root is required")
	}
	absWorktreeRoot, err := filepath.Abs(worktreeRoot)
	if err != nil {
		return "", "", fmt.Errorf("resolve worktree-root: %w", err)
	}
	absWT, err := filepath.Abs(wtPath)
	if err != nil {
		return "", "", fmt.Errorf("resolve worktree path: %w", err)
	}
	if !isSubDir(absWorktreeRoot, absWT) || absWorktreeRoot == absWT {
		return "", "", fmt.Errorf("worktree %q is not under worktree-root %q", absWT, absWorktreeRoot)
	}

	// 1. Try resolving directly from git metadata (.git file and commondir).
	if mainRepo, branch, err := FindMainRepoAndBranchFromGit(wtPath); err == nil && mainRepo != "" {
		if reposRoot != "" {
			absReposRoot, err := filepath.Abs(reposRoot)
			if err != nil {
				return "", "", fmt.Errorf("resolve repos-root: %w", err)
			}
			if isSubDir(absReposRoot, mainRepo) {
				return mainRepo, branch, nil
			}
		} else {
			return mainRepo, branch, nil
		}
	}

	// 2. Fall back to path-based resolution if .git metadata is unavailable (e.g. mock directory tests).
	return resolveMainRepoFromWorktreeByPath(reposRoot, worktreeRoot, wtPath)
}

func resolveMainRepoFromWorktreeByPath(reposRoot, worktreeRoot, wtPath string) (string, string, error) {
	if reposRoot == "" {
		return "", "", fmt.Errorf("--repos-root (-r) is required")
	}
	absReposRoot, err := filepath.Abs(reposRoot)
	if err != nil {
		return "", "", fmt.Errorf("resolve repos-root: %w", err)
	}
	absWorktreeRoot, err := filepath.Abs(worktreeRoot)
	if err != nil {
		return "", "", fmt.Errorf("resolve worktree-root: %w", err)
	}
	absWT, err := filepath.Abs(wtPath)
	if err != nil {
		return "", "", fmt.Errorf("resolve worktree path: %w", err)
	}

	relWT, ok := strictRelPath(absWorktreeRoot, absWT)
	if !ok {
		return "", "", fmt.Errorf("worktree %q is not under worktree-root %q", absWT, absWorktreeRoot)
	}

	cand := filepath.Dir(filepath.Join(absReposRoot, relWT))
	for cand != absReposRoot && strings.HasPrefix(cand, absReposRoot+string(filepath.Separator)) {
		if IsRepo(cand) {
			relRepo, err := filepath.Rel(absReposRoot, cand)
			if err != nil {
				return "", "", err
			}
			wtRepoPrefix := filepath.Join(absWorktreeRoot, relRepo)
			relBranch, err := filepath.Rel(wtRepoPrefix, absWT)
			if err != nil {
				return "", "", err
			}
			return cand, filepath.ToSlash(relBranch), nil
		}
		parent := filepath.Dir(cand)
		if parent == cand {
			break
		}
		cand = parent
	}

	return "", "", fmt.Errorf("main repository not found under %q for worktree %q", absReposRoot, absWT)
}

// ResolveUniqueMainRepos resolves the unique main repository paths under reposRoot
// corresponding to the given worktree paths under worktreeRoot, preserving order.
func ResolveUniqueMainRepos(reposRoot, worktreeRoot string, wtPaths []string) ([]string, error) {
	seen := make(map[string]struct{}, len(wtPaths))
	var repos []string
	for _, wtPath := range wtPaths {
		mainRepo, _, err := ResolveMainRepoFromWorktree(reposRoot, worktreeRoot, wtPath)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[mainRepo]; !ok {
			seen[mainRepo] = struct{}{}
			repos = append(repos, mainRepo)
		}
	}
	return repos, nil
}

// WorktreeAddCmd constructs argv for "git worktree add".
// If newBranch is non-empty, it constructs "git worktree add -b <newBranch> -- <wtPath> <branch>".
// Otherwise, it constructs "git worktree add -- <wtPath> <branch>".
func (c *Command) WorktreeAddCmd(wtPath, branch, newBranch string) []string {
	if newBranch != "" {
		return []string{c.GitBin(), "worktree", "add", "-b", newBranch, "--", wtPath, branch}
	}
	return []string{c.GitBin(), "worktree", "add", "--", wtPath, branch}
}

// WorktreeAdd adds a worktree at wtPath for the repository in repoDir.
func (c *Command) WorktreeAdd(ctx context.Context, repoDir, wtPath, branch, newBranch string, stderr io.Writer) error {
	if err := validateBranchName(branch); err != nil {
		return err
	}
	if newBranch != "" {
		if err := validateBranchName(newBranch); err != nil {
			return err
		}
	}
	return c.Run(ctx, repoDir, c.WorktreeAddCmd(wtPath, branch, newBranch)[1:], stderr, stderr)
}

// WorktreeRemoveCmd constructs argv for "git worktree remove [--force] [--no-force] -- <wtPath>".
func (c *Command) WorktreeRemoveCmd(wtPath string, force, noForce bool) []string {
	cmd := []string{c.GitBin(), "worktree", "remove"}
	if force {
		cmd = append(cmd, "--force")
	}
	if noForce {
		cmd = append(cmd, "--no-force")
	}
	return append(cmd, "--", wtPath)
}

// WorktreeRemove removes the worktree at wtPath from the repository in repoDir.
func (c *Command) WorktreeRemove(ctx context.Context, repoDir, wtPath string, force, noForce bool, stderr io.Writer) error {
	return c.Run(ctx, repoDir, c.WorktreeRemoveCmd(wtPath, force, noForce)[1:], stderr, stderr)
}

// BranchDeleteCmd constructs argv for "git branch -d <branch>" (or "-D" when force is true).
func (c *Command) BranchDeleteCmd(branch string, force bool) []string {
	flag := "-d"
	if force {
		flag = "-D"
	}
	return []string{c.GitBin(), "branch", flag, "--", branch}
}

// BranchDelete deletes branch in repoDir. If force is true, it uses -D instead of -d.
func (c *Command) BranchDelete(ctx context.Context, repoDir, branch string, force bool, stderr io.Writer) error {
	if err := validateBranchName(branch); err != nil {
		return err
	}
	return c.Run(ctx, repoDir, c.BranchDeleteCmd(branch, force)[1:], stderr, stderr)
}


// WorktreePruneCmd constructs argv for "git worktree prune [--dry-run]".
func (c *Command) WorktreePruneCmd(dryRun bool) []string {
	if dryRun {
		return []string{c.GitBin(), "worktree", "prune", "--dry-run"}
	}
	return []string{c.GitBin(), "worktree", "prune"}
}

// WorktreePrune runs "git worktree prune [--dry-run]" in repoDir.
func (c *Command) WorktreePrune(ctx context.Context, repoDir string, dryRun bool, stdout, stderr io.Writer) error {
	return c.Run(ctx, repoDir, c.WorktreePruneCmd(dryRun)[1:], stdout, stderr)
}

// IsMerged reports whether branch has been merged into target in repoDir.
// It checks standard ancestor merges first (git merge-base --is-ancestor), and falls back
// to tree comparison via git merge-tree --write-tree to detect squash merges.
func (c *Command) IsMerged(ctx context.Context, repoDir, branch, target string) (bool, error) {
	if err := validateBranchName(branch); err != nil {
		return false, err
	}
	if err := validateBranchName(target); err != nil {
		return false, err
	}

	// 1. Standard ancestor merge check.
	ancestorArgs := []string{"merge-base", "--is-ancestor", "--", branch, target}
	if err := c.Run(ctx, repoDir, ancestorArgs, io.Discard, io.Discard); err == nil {
		return true, nil
	}

	// 2. Squash/rebase merge check: if merging branch into target produces the exact same tree as target.
	targetTree, err := c.Output(ctx, repoDir, "rev-parse", target+"^{tree}")
	if err != nil || targetTree == "" {
		return false, nil
	}
	mergedTree, err := c.Output(ctx, repoDir, "merge-tree", "--write-tree", "--", target, branch)
	if err != nil || mergedTree == "" {
		return false, nil
	}
	return targetTree == mergedTree, nil
}
