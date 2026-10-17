package cmd

import (
	"fmt"
	"os"

	"github.com/berquerant/git-iter-go/internal/repo"
	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/spf13/cobra"
)

// NewWorktreeCmd returns the "worktree" (alias: "wt") parent sub-command.
func NewWorktreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "worktree",
		Aliases: []string{"wt"},
		Short:   "Manage git worktrees under worktree-root",
		Long: `Manage git worktrees placed under --worktree-root (or <repos-root>/git-worktree).

For a repository located at <repos-root>/<DIR>, worktrees for <BRANCH> are placed at:
  <worktree-root>/<DIR>/<BRANCH>`,
	}

	cmd.AddCommand(
		NewWorktreeListCmd(),
		NewWorktreeAddCmd(),
		NewWorktreeRemoveCmd(),
		NewWorktreePruneCmd(),
	)

	return cmd
}

// NewWorktreeListCmd returns the "worktree list" (aliases: "ls", "l") sub-command.
func NewWorktreeListCmd() *cobra.Command {
	var repoOnly bool

	cmd := &cobra.Command{
		Use:     "list [REPO_REGEX...]",
		Aliases: []string{"ls", "l"},
		Short:   "List git worktrees or their corresponding main repositories",
		Long: LongDoc{
			Header:     "Print the absolute path of every discovered git worktree under --worktree-root matching REPO_REGEX.",
			RepoFilter: true,
			Sections: []Section{
				{
					Title: "Flags:",
					Items: []Item{
						{Term: "--repo", Desc: "print unique main repository paths corresponding to discovered worktrees"},
					},
				},
			},
			OutputModes: RepoListOutputModeDoc(),
			Examples: []Example{
				{
					Desc:    "List all worktrees",
					Command: "git-iter -r /repos worktree list",
				},
				{
					Desc:    "List unique main repositories that have worktrees",
					Command: "git-iter -r /repos worktree list --repo",
				},
			},
		}.Build(),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())
			listArgs := ParseWorktreeListArgs(args, repoOnly)

			finder, err := buildWorktreeFinder(cfg, listArgs.Patterns)
			if err != nil {
				return fmt.Errorf("worktree list: %w", err)
			}

			return RunWorktreeList(
				cmd.Context(),
				cfg,
				OutModeFromContext(cmd.Context()),
				listArgs,
				finder,
				cmd.OutOrStdout(),
			)
		},
	}

	cmd.Flags().BoolVar(&repoOnly, "repo", false,
		"list unique main repository paths corresponding to discovered worktrees")

	return cmd
}

// NewWorktreeAddCmd returns the "worktree add" (alias: "a") sub-command.
func NewWorktreeAddCmd() *cobra.Command {
	var newBranch string

	cmd := &cobra.Command{
		Use:     "add [-b NEW_BRANCH] BRANCH",
		Aliases: []string{"a"},
		Short:   "Add a worktree for the current repository under worktree-root",
		Args:    cobra.ExactArgs(1),
		Long: LongDoc{
			Header: `Create a git worktree for the current repository at <worktree-root>/<DIR>/<BRANCH>
and print the worktree's absolute path to stdout.`,
			Sections: []Section{
				{
					Title: "Flags:",
					Items: []Item{
						{Term: "-b, --branch NEW_BRANCH", Desc: "create a new branch named NEW_BRANCH starting at BRANCH"},
					},
				},
			},
			OutputModes: RepoListOutputModeDoc(),
			Examples: []Example{
				{
					Desc:    "Checkout an existing branch into a new worktree",
					Command: "git-iter -r /repos worktree add feat-1",
				},
				{
					Desc:    "Create a new branch feat-2 from main in a new worktree",
					Command: "git-iter -r /repos worktree add -b feat-2 main",
				},
			},
		}.Build(),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())
			addArgs, err := ParseWorktreeAddArgs(args, newBranch)
			if err != nil {
				return err
			}
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("worktree add: %w", err)
			}

			return RunWorktreeAdd(
				cmd.Context(),
				cfg,
				OutModeFromContext(cmd.Context()),
				addArgs,
				wd,
				runner.ProcessRunner{},
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			)
		},
	}

	cmd.Flags().StringVarP(&newBranch, "branch", "b", "",
		"create a new branch starting at the specified base branch")

	return cmd
}

// NewWorktreeRemoveCmd returns the "worktree remove" (aliases: "r", "rm") sub-command.
func NewWorktreeRemoveCmd() *cobra.Command {
	var (
		force   bool
		noForce bool
	)

	cmd := &cobra.Command{
		Use:     "remove [FLAGS] BRANCH",
		Aliases: []string{"r", "rm"},
		Short:   "Remove the worktree corresponding to BRANCH for the current repository",
		Args:    cobra.ExactArgs(1),
		Long: LongDoc{
			Header: "Remove the git worktree at <worktree-root>/<DIR>/<BRANCH> for the current repository.",
			Sections: []Section{
				{
					Title: "Flags:",
					Items: []Item{
						{Term: "-f, --force", Desc: "force removal even with untracked or modified files"},
						{Term: "--no-force", Desc: "do not force removal"},
					},
				},
			},
			Examples: []Example{
				{
					Desc:    "Remove the worktree for branch feat-1",
					Command: "git-iter -r /repos worktree remove feat-1",
				},
				{
					Desc:    "Force-remove the worktree for branch feat-1",
					Command: "git-iter -r /repos worktree remove -f feat-1",
				},
			},
		}.Build(),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())
			rmArgs, err := ParseWorktreeRemoveArgs(args, force, noForce)
			if err != nil {
				return err
			}
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("worktree remove: %w", err)
			}

			return RunWorktreeRemove(
				cmd.Context(),
				cfg,
				rmArgs,
				wd,
				runner.ProcessRunner{},
				cmd.ErrOrStderr(),
			)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false,
		"force removal even if worktree is dirty or has untracked files")
	cmd.Flags().BoolVar(&noForce, "no-force", false,
		"do not force removal")

	return cmd
}

// NewWorktreePruneCmd returns the "worktree prune" (alias: "p") sub-command.
func NewWorktreePruneCmd() *cobra.Command {
	var (
		dryRun bool
		all    bool
	)

	cmd := &cobra.Command{
		Use:     "prune [FLAGS] [REPO_REGEX...]",
		Aliases: []string{"p"},
		Short:   "Remove worktrees merged into the default branch and prune stale worktree metadata",
		Args:    cobra.ArbitraryArgs,
		Long: LongDoc{
			Header: `Discover worktrees under --worktree-root, remove those whose branch has been merged
into the repository's default branch, and run "git worktree prune".
Without --all, operates on the current repository. With --all, operates on all repositories
that have worktrees under --worktree-root (filtered by optional REPO_REGEX).`,
			RepoFilter: true,
			Sections: []Section{
				{
					Title: "Flags:",
					Items: []Item{
						{Term: "--dry-run", Desc: "do not remove worktrees; report what would be pruned"},
						{Term: "--all", Desc: "prune across all repositories that have worktrees under worktree-root"},
					},
				},
			},
			ExecutionNote: true,
			OutputModes:   DefaultOutputModeDoc(),
			Examples: []Example{
				{
					Desc:    "Prune merged worktrees for the current repository",
					Command: "git-iter -r /repos worktree prune",
				},
				{
					Desc:    "Dry-run prune across all repositories with worktrees",
					Command: "git-iter -r /repos worktree prune --all --dry-run",
				},
			},
		}.Build(),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())
			pruneArgs := ParseWorktreePruneArgs(args, dryRun, all)

			var (
				wd     string
				finder repo.Finder
				err    error
			)
			if pruneArgs.All {
				finder, err = buildWorktreeFinder(cfg, pruneArgs.Patterns)
				if err != nil {
					return fmt.Errorf("worktree prune: %w", err)
				}
			} else {
				wd, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("worktree prune: %w", err)
				}
			}

			pruneRunner := &runner.WorktreePruneRunner{
				GitCommand:   cfg.GitCommand,
				ReposRoot:    cfg.ReposRoot,
				WorktreeRoot: cfg.WorktreeRootOrFallback(),
				DryRun:       pruneArgs.DryRun,
				Inner:        runner.ProcessRunner{},
			}

			return RunWorktreePrune(
				cmd.Context(),
				cfg,
				OutModeFromContext(cmd.Context()),
				pruneArgs,
				wd,
				finder,
				pruneRunner,
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"do not remove anything; just report what would be pruned")
	cmd.Flags().BoolVar(&all, "all", false,
		"prune across all repositories that have worktrees under worktree-root")

	return cmd
}
