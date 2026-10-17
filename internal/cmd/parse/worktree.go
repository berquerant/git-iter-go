package parse

import "fmt"

// WorktreeListArgs is the parsed argument set for "worktree list".
type WorktreeListArgs struct {
	Patterns []string // worktree path filter regexps; empty means all
	Repo     bool     // --repo: output unique main repository paths instead of worktree paths
	All      bool     // -a, --all: list across all repositories instead of current repository
	RepoDir  string   // target repo directory when All is false (defaults to cwd)
}

// ParseWorktreeListArgs parses raw cobra args and flags into WorktreeListArgs.
func ParseWorktreeListArgs(args []string, repoOnly bool, all ...bool) WorktreeListArgs {
	isAll := false
	if len(all) > 0 {
		isAll = all[0]
	}
	return WorktreeListArgs{
		Patterns: args,
		Repo:     repoOnly,
		All:      isAll,
	}
}

// WorktreeAddArgs is the parsed argument set for "worktree add".
type WorktreeAddArgs struct {
	Branch    string // existing branch (when NewBranch is empty) or base branch (when NewBranch is set)
	NewBranch string // -b <new-branch>
}

// TargetBranch returns the branch name that determines the worktree directory path.
func (a WorktreeAddArgs) TargetBranch() string {
	if a.NewBranch != "" {
		return a.NewBranch
	}
	return a.Branch
}

// ParseWorktreeAddArgs parses raw cobra args and the -b flag into WorktreeAddArgs.
func ParseWorktreeAddArgs(args []string, newBranch string) (WorktreeAddArgs, error) {
	if len(args) != 1 {
		return WorktreeAddArgs{}, fmt.Errorf("worktree add requires exactly 1 branch argument, got %d", len(args))
	}
	return WorktreeAddArgs{
		Branch:    args[0],
		NewBranch: newBranch,
	}, nil
}

// WorktreeRemoveArgs is the parsed argument set for "worktree remove".
type WorktreeRemoveArgs struct {
	Branch  string
	Force   bool
	NoForce bool
}

// ParseWorktreeRemoveArgs parses raw cobra args and flags into WorktreeRemoveArgs.
func ParseWorktreeRemoveArgs(args []string, force, noForce bool) (WorktreeRemoveArgs, error) {
	if len(args) != 1 {
		return WorktreeRemoveArgs{}, fmt.Errorf("worktree remove requires exactly 1 branch argument, got %d", len(args))
	}
	return WorktreeRemoveArgs{
		Branch:  args[0],
		Force:   force,
		NoForce: noForce,
	}, nil
}

// WorktreePruneArgs is the parsed argument set for "worktree prune".
type WorktreePruneArgs struct {
	Patterns []string
	DryRun   bool
	All      bool
}

// ParseWorktreePruneArgs parses raw cobra args and flags into WorktreePruneArgs.
func ParseWorktreePruneArgs(args []string, dryRun, all bool) WorktreePruneArgs {
	return WorktreePruneArgs{
		Patterns: args,
		DryRun:   dryRun,
		All:      all,
	}
}
