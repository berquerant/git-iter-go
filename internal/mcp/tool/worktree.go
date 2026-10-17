package tool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/berquerant/git-iter-go/internal/output"
	"github.com/berquerant/git-iter-go/internal/repo"
	"github.com/berquerant/git-iter-go/internal/runner"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// WorktreeListToolInput defines parameters for git_iter_worktree_list tool.
type WorktreeListToolInput struct {
	WorktreeRoot string `json:"worktree_root,omitempty" jsonschema:"root directory of git worktrees (defaults to <repos_root>/git-worktree)"`
	Repo         bool   `json:"repo,omitempty" jsonschema:"when true, return unique main repository paths corresponding to discovered worktrees"`
	RepoDiscoveryInput
}

// WorktreeListToolOutput defines output for git_iter_worktree_list tool.
type WorktreeListToolOutput struct {
	Worktrees    []output.WorktreeResult `json:"worktrees,omitempty" jsonschema:"discovered worktrees with main repository path and branch"`
	Repositories []output.RepoResult     `json:"repositories,omitempty" jsonschema:"discovered worktrees or corresponding main repositories"`
}

// WorktreeAddToolInput defines parameters for git_iter_worktree_add tool.
type WorktreeAddToolInput struct {
	Branch       string `json:"branch" jsonschema:"existing branch name to checkout (or base branch when new_branch is specified)"`
	NewBranch    string `json:"new_branch,omitempty" jsonschema:"optional new branch name (-b) to create starting from branch"`
	RepoDir      string `json:"repo_dir,omitempty" jsonschema:"target main repository directory (defaults to current working directory)"`
	ReposRoot    string `json:"repos_root,omitempty" jsonschema:"root directory of local repositories"`
	WorktreeRoot string `json:"worktree_root,omitempty" jsonschema:"root directory of git worktrees (defaults to <repos_root>/git-worktree)"`
	Timeout      string `json:"timeout,omitempty" jsonschema:"per-process execution timeout (e.g. 5s, 1m)"`
}

// WorktreeAddToolOutput defines output for git_iter_worktree_add tool.
type WorktreeAddToolOutput struct {
	Worktree output.RepoResult `json:"worktree" jsonschema:"created worktree path"`
}

// WorktreeRemoveToolInput defines parameters for git_iter_worktree_remove tool.
type WorktreeRemoveToolInput struct {
	Branch       string `json:"branch" jsonschema:"branch name whose worktree should be removed"`
	Force        bool   `json:"force,omitempty" jsonschema:"force removal even if worktree is dirty"`
	NoForce      bool   `json:"no_force,omitempty" jsonschema:"do not force removal"`
	RepoDir      string `json:"repo_dir,omitempty" jsonschema:"target main repository directory (defaults to current working directory)"`
	ReposRoot    string `json:"repos_root,omitempty" jsonschema:"root directory of local repositories"`
	WorktreeRoot string `json:"worktree_root,omitempty" jsonschema:"root directory of git worktrees (defaults to <repos_root>/git-worktree)"`
	Timeout      string `json:"timeout,omitempty" jsonschema:"per-process execution timeout (e.g. 5s, 1m)"`
}

// WorktreeRemoveToolOutput defines output for git_iter_worktree_remove tool.
type WorktreeRemoveToolOutput struct {
	Worktree output.RepoResult `json:"worktree" jsonschema:"removed worktree path"`
}

// WorktreePruneToolInput defines parameters for git_iter_worktree_prune tool.
type WorktreePruneToolInput struct {
	DryRun       bool   `json:"dry_run,omitempty" jsonschema:"do not remove anything; report what would be pruned"`
	All          bool   `json:"all,omitempty" jsonschema:"prune across all repositories that have worktrees under worktree_root"`
	RepoDir      string `json:"repo_dir,omitempty" jsonschema:"target main repository directory when all is false (defaults to current working directory)"`
	WorktreeRoot string `json:"worktree_root,omitempty" jsonschema:"root directory of git worktrees (defaults to <repos_root>/git-worktree)"`
	RepoDiscoveryInput
	ProcessExecInput
}

// WorktreePruneToolOutput defines output for git_iter_worktree_prune tool.
type WorktreePruneToolOutput struct {
	Results []output.Result `json:"results" jsonschema:"prune execution results per repository"`
}

func resolveRoots(inReposRoot, inWorktreeRoot string, cfg *config.Config) (string, string, error) {
	effective := &config.Config{}
	if cfg != nil {
		*effective = *cfg
	}
	if inReposRoot != "" {
		effective.ReposRoot = inReposRoot
	}
	if inWorktreeRoot != "" {
		effective.WorktreeRoot = inWorktreeRoot
	}
	wtRoot := effective.WorktreeRootOrFallback()
	if wtRoot == "" {
		return "", "", fmt.Errorf("worktree_root or repos_root must be specified")
	}
	return effective.ReposRoot, wtRoot, nil
}

func resolveRepoDir(inRepoDir string) (string, error) {
	if inRepoDir != "" {
		return git.FindRepoRoot(inRepoDir)
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return git.FindRepoRoot(wd)
}

func findWorktreePaths(ctx context.Context, wtRoot string, cfg *config.Config, discovery RepoDiscoveryInput) ([]string, error) {
	base := &repo.FilesystemFinder{Root: wtRoot}
	finder, err := buildFinderFromBase(base, cfg, finderParams{Discovery: discovery})
	if err != nil {
		return nil, err
	}
	paths, err := finder.Find(ctx)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("find worktrees: %w", err)
	}
	return paths, nil
}

// RegisterWorktreeList registers the read-only git_iter_worktree_list tool on the given MCP server.
func RegisterWorktreeList(server *mcpsdk.Server, cfg *config.Config) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "git_iter_worktree_list",
		Description: "Discover and list git worktrees under worktree_root, or list their unique corresponding main repositories when repo is true",
	}, func(
		ctx context.Context,
		_ *mcpsdk.CallToolRequest,
		in WorktreeListToolInput,
	) (*mcpsdk.CallToolResult, WorktreeListToolOutput, error) {
		reposRoot, wtRoot, err := resolveRoots(in.ReposRoot, in.WorktreeRoot, cfg)
		if err != nil {
			return nil, WorktreeListToolOutput{}, err
		}

		wtPaths, err := findWorktreePaths(ctx, wtRoot, cfg, in.RepoDiscoveryInput)
		if err != nil {
			return nil, WorktreeListToolOutput{}, err
		}

		wtResults := make([]output.WorktreeResult, len(wtPaths))
		for i, wtPath := range wtPaths {
			mainRepo, branch, _ := git.ResolveMainRepoFromWorktree(reposRoot, wtRoot, wtPath)
			var repoRel string
			if mainRepo != "" && reposRoot != "" {
				repoRel = output.RelPath(reposRoot, mainRepo)
			}
			wtResults[i] = output.WorktreeResult{
				WorktreeAbsPath: wtPath,
				WorktreeRelPath: output.RelPath(wtRoot, wtPath),
				RepoAbsPath:     mainRepo,
				RepoRelPath:     repoRel,
				Branch:          branch,
			}
		}

		if in.Repo {
			mainRepos, err := git.ResolveUniqueMainRepos(reposRoot, wtRoot, wtPaths)
			if err != nil {
				return nil, WorktreeListToolOutput{}, err
			}
			return nil, WorktreeListToolOutput{
				Repositories: toRepoResults(reposRoot, mainRepos),
			}, nil
		}

		return nil, WorktreeListToolOutput{
			Worktrees:    wtResults,
			Repositories: toRepoResults(wtRoot, wtPaths),
		}, nil
	})
}

// RegisterWorktreeMutations registers git_iter_worktree_add, git_iter_worktree_remove,
// and git_iter_worktree_prune on the given MCP server.
func RegisterWorktreeMutations(server *mcpsdk.Server, cfg *config.Config) {
	registerWorktreeAdd(server, cfg)
	registerWorktreeRemove(server, cfg)
	registerWorktreePrune(server, cfg)
}

func runWorktreeTargetOp(
	ctx context.Context,
	cfg *config.Config,
	inReposRoot, inWorktreeRoot, inRepoDir, targetBranch, inTimeout string,
	op func(runCtx context.Context, g *git.Command, repoRoot, wtPath string) error,
) (output.RepoResult, error) {
	reposRoot, wtRoot, err := resolveRoots(inReposRoot, inWorktreeRoot, cfg)
	if err != nil {
		return output.RepoResult{}, err
	}
	repoRoot, err := resolveRepoDir(inRepoDir)
	if err != nil {
		return output.RepoResult{}, err
	}
	wtPath, err := git.ResolveWorktreePath(reposRoot, wtRoot, repoRoot, targetBranch)
	if err != nil {
		return output.RepoResult{}, err
	}
	runCtx, cancel, err := withResolvedTimeout(ctx, inTimeout, cfg)
	if err != nil {
		return output.RepoResult{}, err
	}
	defer cancel()

	g := git.New(cfg.GitCommandOrFallback(), nil)
	if err := op(runCtx, g, repoRoot, wtPath); err != nil {
		return output.RepoResult{}, err
	}
	return toRepoResult(wtRoot, wtPath), nil
}

func registerWorktreeAdd(server *mcpsdk.Server, cfg *config.Config) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "git_iter_worktree_add",
		Description: "Add a git worktree under worktree_root for the target repository and return the created worktree path",
	}, func(
		ctx context.Context,
		_ *mcpsdk.CallToolRequest,
		in WorktreeAddToolInput,
	) (*mcpsdk.CallToolResult, WorktreeAddToolOutput, error) {
		targetBranch := in.Branch
		if in.NewBranch != "" {
			targetBranch = in.NewBranch
		}
		wt, err := runWorktreeTargetOp(ctx, cfg, in.ReposRoot, in.WorktreeRoot, in.RepoDir, targetBranch, in.Timeout,
			func(runCtx context.Context, g *git.Command, repoRoot, wtPath string) error {
				return g.WorktreeAdd(runCtx, repoRoot, wtPath, in.Branch, in.NewBranch, io.Discard)
			},
		)
		if err != nil {
			return nil, WorktreeAddToolOutput{}, err
		}
		return nil, WorktreeAddToolOutput{Worktree: wt}, nil
	})
}

func registerWorktreeRemove(server *mcpsdk.Server, cfg *config.Config) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "git_iter_worktree_remove",
		Description: "Remove the git worktree corresponding to branch for the target repository",
	}, func(
		ctx context.Context,
		_ *mcpsdk.CallToolRequest,
		in WorktreeRemoveToolInput,
	) (*mcpsdk.CallToolResult, WorktreeRemoveToolOutput, error) {
		wt, err := runWorktreeTargetOp(ctx, cfg, in.ReposRoot, in.WorktreeRoot, in.RepoDir, in.Branch, in.Timeout,
			func(runCtx context.Context, g *git.Command, repoRoot, wtPath string) error {
				return g.WorktreeRemove(runCtx, repoRoot, wtPath, in.Force, in.NoForce, io.Discard)
			},
		)
		if err != nil {
			return nil, WorktreeRemoveToolOutput{}, err
		}
		return nil, WorktreeRemoveToolOutput{Worktree: wt}, nil
	})
}

func registerWorktreePrune(server *mcpsdk.Server, cfg *config.Config) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "git_iter_worktree_prune",
		Description: "Remove worktrees merged into the default branch and run git worktree prune",
	}, func(
		ctx context.Context,
		_ *mcpsdk.CallToolRequest,
		in WorktreePruneToolInput,
	) (*mcpsdk.CallToolResult, WorktreePruneToolOutput, error) {
		reposRoot, wtRoot, err := resolveRoots(in.ReposRoot, in.WorktreeRoot, cfg)
		if err != nil {
			return nil, WorktreePruneToolOutput{}, err
		}

		var paths []string
		if in.All {
			wtPaths, err := findWorktreePaths(ctx, wtRoot, cfg, in.RepoDiscoveryInput)
			if err != nil {
				return nil, WorktreePruneToolOutput{}, err
			}
			paths, err = git.ResolveUniqueMainRepos(reposRoot, wtRoot, wtPaths)
			if err != nil {
				return nil, WorktreePruneToolOutput{}, err
			}
		} else {
			repoRoot, err := resolveRepoDir(in.RepoDir)
			if err != nil {
				return nil, WorktreePruneToolOutput{}, err
			}
			paths = []string{repoRoot}
		}

		timeout, err := resolveTimeout(in.Timeout, cfg)
		if err != nil {
			return nil, WorktreePruneToolOutput{}, err
		}

		gitCmd := cfg.GitCommandOrFallback()
		pruneRunner := &runner.WorktreePruneRunner{
			GitCommand:   gitCmd,
			ReposRoot:    reposRoot,
			WorktreeRoot: wtRoot,
			DryRun:       in.DryRun,
			Inner:        runner.ProcessRunner{},
		}
		cmd := git.New(gitCmd, nil).WorktreePruneCmd(in.DryRun)

		tasks := make([]output.Task, len(paths))
		for i, p := range paths {
			tasks[i] = output.Task{
				RepoAbsPath: p,
				ReposRoot:   reposRoot,
				Dir:         p,
				Command:     cmd,
				Runner:      pruneRunner,
			}
		}

		results, err := executeTasksCollect(ctx, tasks, resolveProcs(in.MaxProcs, cfg), timeout)
		if err != nil {
			return nil, WorktreePruneToolOutput{}, err
		}
		return nil, WorktreePruneToolOutput{Results: results}, nil
	})
}
