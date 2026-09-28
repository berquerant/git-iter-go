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

type finderParams struct {
	reposRoot            string
	patterns             []string
	sortAsc              bool
	sortDesc             bool
	limit                int
	dirtyOnly            bool
	cleanOnly            bool
	defaultBranchOnly    bool
	notDefaultBranchOnly bool
	branch               string
	remoteURL            string
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
	dirtyOnly := p.dirtyOnly
	cleanOnly := p.cleanOnly
	if !dirtyOnly && !cleanOnly && cfg != nil {
		dirtyOnly = cfg.DirtyOnly
		cleanOnly = cfg.CleanOnly
	}
	if dirtyOnly && cleanOnly {
		return nil, "", ErrConflictingFilterFlags
	}

	defaultBranchOnly := p.defaultBranchOnly
	notDefaultBranchOnly := p.notDefaultBranchOnly
	if !defaultBranchOnly && !notDefaultBranchOnly && cfg != nil {
		defaultBranchOnly = cfg.DefaultBranchOnly
		notDefaultBranchOnly = cfg.NotDefaultBranchOnly
	}
	if defaultBranchOnly && notDefaultBranchOnly {
		return nil, "", ErrConflictingBranchFlags
	}

	root := p.reposRoot
	listRepos := ""
	if root == "" && cfg != nil {
		root = cfg.ReposRoot
		listRepos = cfg.ListRepos
	}
	if root == "" && listRepos == "" {
		return nil, "", fmt.Errorf("repos_root must be specified in parameters or via server configuration")
	}

	base := repo.NewFinder(root, listRepos)
	finder, err := repo.NewFilteredFinder(base, p.patterns)
	if err != nil {
		return nil, "", fmt.Errorf("invalid pattern: %w", err)
	}

	sortAsc := p.sortAsc
	sortDesc := p.sortDesc
	if !sortAsc && !sortDesc && cfg != nil {
		sortAsc = cfg.Sort
		sortDesc = cfg.SortReverse
	}
	finder = repo.NewSortedFinder(finder, sortAsc, sortDesc)

	limit := p.limit
	if limit <= 0 && cfg != nil {
		limit = cfg.Limit
	}
	finder = repo.NewLimitedFinder(finder, limit)
	finder = applyStatusFilter(finder, cfg, dirtyOnly, cleanOnly)
	finder = applyDefaultBranchFilter(finder, cfg, defaultBranchOnly, notDefaultBranchOnly)

	finder, err = applyBranchFilter(finder, cfg, p.branch)
	if err != nil {
		return nil, "", err
	}

	finder, err = applyRemoteURLFilter(finder, cfg, p.remoteURL)
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
