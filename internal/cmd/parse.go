package cmd

import (
	"fmt"
	"regexp"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/berquerant/git-iter-go/internal/repo"
)

// ErrConflictingFilterFlags is returned when both --dirty-only and --clean-only are set.
var ErrConflictingFilterFlags = fmt.Errorf("cannot specify both --dirty-only (-d) and --clean-only (-c)")

// ErrConflictingBranchFlags is returned when both --default-branch-only and --not-default-branch-only are set.
var ErrConflictingBranchFlags = fmt.Errorf("cannot specify both --default-branch-only and --not-default-branch-only")

// buildFinder constructs a filtered, sorted, and limited Finder from config and patterns.
func buildFinder(cfg *config.Config, patterns []string) (repo.Finder, error) {
	if cfg.DirtyOnly && cfg.CleanOnly {
		return nil, ErrConflictingFilterFlags
	}
	if cfg.DefaultBranchOnly && cfg.NotDefaultBranchOnly {
		return nil, ErrConflictingBranchFlags
	}

	base := repo.NewFinder(cfg.ReposRoot, cfg.ListRepos)
	finder, err := repo.NewFilteredFinder(base, patterns)
	if err != nil {
		return nil, err
	}
	finder = repo.NewSortedFinder(finder, cfg.Sort, cfg.SortReverse)
	finder = repo.NewLimitedFinder(finder, cfg.Limit)

	if cfg.DirtyOnly || cfg.CleanOnly {
		g := git.New(cfg.GitCommandOrFallback(), nil)
		finder = repo.NewStatusFilterFinder(finder, g, cfg.DirtyOnly, cfg.CleanOnly)
	}

	if cfg.DefaultBranchOnly || cfg.NotDefaultBranchOnly {
		g := git.New(cfg.GitCommandOrFallback(), nil)
		finder = repo.NewDefaultBranchFilterFinder(finder, g, cfg.DefaultBranchOnly, cfg.NotDefaultBranchOnly)
	}

	if cfg.Branch != "" {
		re, err := regexp.Compile(cfg.Branch)
		if err != nil {
			return nil, fmt.Errorf("invalid branch regex %q: %w", cfg.Branch, err)
		}
		g := git.New(cfg.GitCommandOrFallback(), nil)
		finder = repo.NewBranchFilterFinder(finder, g, re)
	}

	if cfg.RemoteURL != "" {
		re, err := regexp.Compile(cfg.RemoteURL)
		if err != nil {
			return nil, fmt.Errorf("invalid remote-url regex %q: %w", cfg.RemoteURL, err)
		}
		g := git.New(cfg.GitCommandOrFallback(), nil)
		finder = repo.NewRemoteURLFilterFinder(finder, g, re)
	}

	return finder, nil
}

// Aliases for types and functions moved to internal/cmd/parse.
type (
	DoArgs      = parse.DoArgs
	GrepArgs    = parse.GrepArgs
	ListArgs    = parse.ListArgs
	ReadArgs    = parse.ReadArgs
	PullArgs    = parse.PullArgs
	FetchArgs   = parse.FetchArgs
	LsFilesArgs = parse.LsFilesArgs
	StatusArgs  = parse.StatusArgs
	DiffArgs    = parse.DiffArgs
)

var (
	ParseDoArgs              = parse.ParseDoArgs
	ParseDoArgsWithDash      = parse.ParseDoArgsWithDash
	ParseGrepArgs            = parse.ParseGrepArgs
	ParseGrepArgsWithDash    = parse.ParseGrepArgsWithDash
	ParseListArgs            = parse.ParseListArgs
	ParseReadArgs            = parse.ParseReadArgs
	ParsePullArgs            = parse.ParsePullArgs
	ParsePullArgsWithDash    = parse.ParsePullArgsWithDash
	ParseFetchArgs           = parse.ParseFetchArgs
	ParseFetchArgsWithDash   = parse.ParseFetchArgsWithDash
	ParseLsFilesArgs         = parse.ParseLsFilesArgs
	ParseLsFilesArgsWithDash = parse.ParseLsFilesArgsWithDash
	ParseStatusArgs          = parse.ParseStatusArgs
	ParseStatusArgsWithDash  = parse.ParseStatusArgsWithDash
	ParseDiffArgs            = parse.ParseDiffArgs
	ParseDiffArgsWithDash    = parse.ParseDiffArgsWithDash
)
