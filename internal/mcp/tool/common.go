package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/berquerant/git-iter-go/internal/output"
	"github.com/berquerant/git-iter-go/internal/repo"
	"github.com/berquerant/git-iter-go/internal/runner"
)

// ErrConflictingFilterFlags is returned when both dirty_only and clean_only are set.
var ErrConflictingFilterFlags = errors.New("cannot specify both dirty_only and clean_only")

// ErrConflictingBranchFlags is returned when both default_branch_only and not_default_branch_only are set.
var ErrConflictingBranchFlags = errors.New("cannot specify both default_branch_only and not_default_branch_only")

// RepoDiscoveryInput defines common parameters for discovering and filtering git repositories.
type RepoDiscoveryInput struct {
	Patterns             []string `json:"patterns,omitempty" jsonschema:"repository path filter regular expressions"`
	ReposRoot            string   `json:"repos_root,omitempty" jsonschema:"root directory to scan for repositories"`
	Sort                 bool     `json:"sort,omitempty" jsonschema:"sort repository paths in ascending order"`
	SortReverse          bool     `json:"sort_reverse,omitempty" jsonschema:"sort repository paths in descending order"`
	Limit                int      `json:"limit,omitempty" jsonschema:"limit number of processed repositories"`
	DefaultBranchOnly    bool     `json:"default_branch_only,omitempty" jsonschema:"only include repositories whose current branch is the default branch"`
	NotDefaultBranchOnly bool     `json:"not_default_branch_only,omitempty" jsonschema:"only include repositories whose current branch is not the default branch"`
	Branch               string   `json:"branch,omitempty" jsonschema:"filter repositories by current branch regex"`
	RemoteURL            string   `json:"remote_url,omitempty" jsonschema:"filter repositories by origin remote URL regex"`
}

// RepoStatusFilterInput defines dirty/clean status filters for git status and git diff.
type RepoStatusFilterInput struct {
	DirtyOnly bool `json:"dirty_only,omitempty" jsonschema:"only run in repositories with uncommitted changes"`
	CleanOnly bool `json:"clean_only,omitempty" jsonschema:"only run in repositories without uncommitted changes"`
}

// ProcessExecInput defines process execution control parameters.
type ProcessExecInput struct {
	MaxProcs int    `json:"max_procs,omitempty" jsonschema:"maximum parallel processes (default 1)"`
	Timeout  string `json:"timeout,omitempty" jsonschema:"per-process execution timeout (e.g. 5s, 1m)"`
}

type finderParams struct {
	Discovery RepoDiscoveryInput
	Status    RepoStatusFilterInput
}

func applyStatusFilter(finder repo.Finder, cfg *config.Config, dirtyOnly, cleanOnly bool) repo.Finder {
	if !dirtyOnly && !cleanOnly {
		return finder
	}
	gitCmd := ""
	if cfg != nil {
		gitCmd = cfg.GitCommandOrFallback()
	}
	g := git.New(gitCmd, nil)
	return repo.NewStatusFilterFinder(finder, g, dirtyOnly, cleanOnly)
}

func applyDefaultBranchFilter(finder repo.Finder, cfg *config.Config, defaultOnly, notDefaultOnly bool) repo.Finder {
	if !defaultOnly && !notDefaultOnly {
		return finder
	}
	gitCmd := ""
	if cfg != nil {
		gitCmd = cfg.GitCommandOrFallback()
	}
	g := git.New(gitCmd, nil)
	return repo.NewDefaultBranchFilterFinder(finder, g, defaultOnly, notDefaultOnly)
}

func applyBranchFilter(finder repo.Finder, cfg *config.Config, branch string) (repo.Finder, error) {
	if branch == "" && cfg != nil {
		branch = cfg.Branch
	}
	if branch == "" {
		return finder, nil
	}
	re, err := regexp.Compile(branch)
	if err != nil {
		return nil, fmt.Errorf("invalid branch regex %q: %w", branch, err)
	}
	gitCmd := ""
	if cfg != nil {
		gitCmd = cfg.GitCommandOrFallback()
	}
	g := git.New(gitCmd, nil)
	return repo.NewBranchFilterFinder(finder, g, re), nil
}

func applyRemoteURLFilter(finder repo.Finder, cfg *config.Config, remoteURL string) (repo.Finder, error) {
	if remoteURL == "" && cfg != nil {
		remoteURL = cfg.RemoteURL
	}
	if remoteURL == "" {
		return finder, nil
	}
	re, err := regexp.Compile(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("invalid remote_url regex %q: %w", remoteURL, err)
	}
	gitCmd := ""
	if cfg != nil {
		gitCmd = cfg.GitCommandOrFallback()
	}
	g := git.New(gitCmd, nil)
	return repo.NewRemoteURLFilterFinder(finder, g, re), nil
}

func buildFinder(cfg *config.Config, p finderParams) (repo.Finder, string, error) {
	dirtyOnly := p.Status.DirtyOnly
	cleanOnly := p.Status.CleanOnly
	if !dirtyOnly && !cleanOnly && cfg != nil {
		dirtyOnly = cfg.DirtyOnly
		cleanOnly = cfg.CleanOnly
	}
	if dirtyOnly && cleanOnly {
		return nil, "", ErrConflictingFilterFlags
	}

	defaultBranchOnly := p.Discovery.DefaultBranchOnly
	notDefaultBranchOnly := p.Discovery.NotDefaultBranchOnly
	if !defaultBranchOnly && !notDefaultBranchOnly && cfg != nil {
		defaultBranchOnly = cfg.DefaultBranchOnly
		notDefaultBranchOnly = cfg.NotDefaultBranchOnly
	}
	if defaultBranchOnly && notDefaultBranchOnly {
		return nil, "", ErrConflictingBranchFlags
	}

	root := p.Discovery.ReposRoot
	listRepos := ""
	if root == "" && cfg != nil {
		root = cfg.ReposRoot
		listRepos = cfg.ListRepos
	}
	if root == "" && listRepos == "" {
		return nil, "", fmt.Errorf("repos_root must be specified in parameters or via server configuration")
	}

	base := repo.NewFinder(root, listRepos)
	finder, err := repo.NewFilteredFinder(base, p.Discovery.Patterns)
	if err != nil {
		return nil, "", fmt.Errorf("invalid pattern: %w", err)
	}

	sortAsc := p.Discovery.Sort
	sortDesc := p.Discovery.SortReverse
	if !sortAsc && !sortDesc && cfg != nil {
		sortAsc = cfg.Sort
		sortDesc = cfg.SortReverse
	}
	finder = repo.NewSortedFinder(finder, sortAsc, sortDesc)

	limit := p.Discovery.Limit
	if limit <= 0 && cfg != nil {
		limit = cfg.Limit
	}
	finder = repo.NewLimitedFinder(finder, limit)
	finder = applyStatusFilter(finder, cfg, dirtyOnly, cleanOnly)
	finder = applyDefaultBranchFilter(finder, cfg, defaultBranchOnly, notDefaultBranchOnly)

	finder, err = applyBranchFilter(finder, cfg, p.Discovery.Branch)
	if err != nil {
		return nil, "", err
	}

	finder, err = applyRemoteURLFilter(finder, cfg, p.Discovery.RemoteURL)
	if err != nil {
		return nil, "", err
	}

	return finder, root, nil
}

func resolveProcs(inMaxProcs int, cfg *config.Config) int {
	if inMaxProcs >= 1 {
		return inMaxProcs
	}
	if cfg != nil && cfg.MaxProcs > 0 {
		return cfg.MaxProcs
	}
	return 1
}

func resolveTimeout(inTimeout string, cfg *config.Config) (time.Duration, error) {
	if inTimeout != "" {
		timeout, err := time.ParseDuration(inTimeout)
		if err != nil {
			return 0, fmt.Errorf("invalid timeout %q: %w", inTimeout, err)
		}
		return timeout, nil
	}
	if cfg != nil {
		return cfg.Timeout, nil
	}
	return 0, nil
}

func buildProcessTasks(paths []string, root string, command []string) []output.Task {
	tasks := make([]output.Task, len(paths))
	for i, p := range paths {
		tasks[i] = output.Task{
			RepoAbsPath: p,
			ReposRoot:   root,
			Dir:         p,
			Command:     command,
			Runner:      runner.ProcessRunner{},
		}
	}
	return tasks
}

// executeTasksCollect runs tasks concurrently and collects Results in-memory using JSONLExecutor.
func executeTasksCollect(
	ctx context.Context,
	tasks []output.Task,
	maxProcs int,
	timeout time.Duration,
) ([]output.Result, error) {
	if len(tasks) == 0 {
		return []output.Result{}, nil
	}

	var buf bytes.Buffer
	exec := &output.JSONLExecutor{
		MaxProcs: maxProcs,
		Timeout:  timeout,
		Out:      &buf,
	}

	if err := exec.Run(ctx, tasks); err != nil {
		return nil, err
	}

	var results []output.Result
	dec := json.NewDecoder(&buf)
	for dec.More() {
		var res output.Result
		if err := dec.Decode(&res); err != nil {
			return nil, fmt.Errorf("decode task result: %w", err)
		}
		results = append(results, res)
	}

	return results, nil
}
