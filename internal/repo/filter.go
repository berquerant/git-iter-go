package repo

import (
	"context"
	"regexp"
	"strings"
)

// FilteredFinder wraps another Finder and keeps only paths that match Pattern.
// A nil Pattern passes all paths through.
type FilteredFinder struct {
	Inner   Finder
	Pattern *regexp.Regexp
}

func (f *FilteredFinder) Find(ctx context.Context) ([]string, error) {
	all, err := f.Inner.Find(ctx)
	if err != nil {
		return nil, err
	}
	if f.Pattern == nil {
		return all, nil
	}
	var out []string
	for _, p := range all {
		if f.Pattern.MatchString(p) {
			out = append(out, p)
		}
	}
	return out, nil
}

// NewFilteredFinder wraps finder with a FilteredFinder using patterns.
// Multiple patterns are joined with "|" to form a single regexp.
// If patterns is empty, finder is returned as-is.
func NewFilteredFinder(finder Finder, patterns []string) (Finder, error) {
	if len(patterns) == 0 {
		return finder, nil
	}
	combined := strings.Join(patterns, "|")
	re, err := regexp.Compile(combined)
	if err != nil {
		return nil, err
	}
	return &FilteredFinder{Inner: finder, Pattern: re}, nil
}

// DirtyChecker determines if a repository at dir has uncommitted changes.
type DirtyChecker interface {
	IsDirty(ctx context.Context, dir string) (bool, error)
}

// StatusFilterFinder filters repositories based on whether they are dirty or clean.
type StatusFilterFinder struct {
	Inner     Finder
	Checker   DirtyChecker
	DirtyOnly bool // if true, only keep dirty repos
	CleanOnly bool // if true, only keep clean repos
}

func (f *StatusFilterFinder) Find(ctx context.Context) ([]string, error) {
	paths, err := f.Inner.Find(ctx)
	if err != nil {
		return nil, err
	}
	if !f.DirtyOnly && !f.CleanOnly {
		return paths, nil
	}

	var res []string
	for _, p := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		isDirty, err := f.Checker.IsDirty(ctx, p)
		if err != nil {
			return nil, err
		}
		if f.DirtyOnly && isDirty {
			res = append(res, p)
		} else if f.CleanOnly && !isDirty {
			res = append(res, p)
		}
	}
	return res, nil
}

// NewStatusFilterFinder wraps finder with a StatusFilterFinder if dirtyOnly or cleanOnly is true.
func NewStatusFilterFinder(finder Finder, checker DirtyChecker, dirtyOnly, cleanOnly bool) Finder {
	if dirtyOnly || cleanOnly {
		return &StatusFilterFinder{
			Inner:     finder,
			Checker:   checker,
			DirtyOnly: dirtyOnly,
			CleanOnly: cleanOnly,
		}
	}
	return finder
}

// RemoteURLChecker determines the remote URL of a repository at dir.
type RemoteURLChecker interface {
	RemoteURL(ctx context.Context, dir string) (string, error)
}

// RemoteURLFilterFinder filters repositories based on whether their remote URL matches Pattern.
type RemoteURLFilterFinder struct {
	Inner   Finder
	Checker RemoteURLChecker
	Pattern *regexp.Regexp
}

func (f *RemoteURLFilterFinder) Find(ctx context.Context) ([]string, error) {
	paths, err := f.Inner.Find(ctx)
	if err != nil {
		return nil, err
	}
	if f.Pattern == nil {
		return paths, nil
	}

	var res []string
	for _, p := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		url, err := f.Checker.RemoteURL(ctx, p)
		if err != nil {
			return nil, err
		}
		if url != "" && f.Pattern.MatchString(url) {
			res = append(res, p)
		}
	}
	return res, nil
}

// NewRemoteURLFilterFinder wraps finder with a RemoteURLFilterFinder if pattern is not nil.
func NewRemoteURLFilterFinder(finder Finder, checker RemoteURLChecker, pattern *regexp.Regexp) Finder {
	if pattern != nil {
		return &RemoteURLFilterFinder{
			Inner:   finder,
			Checker: checker,
			Pattern: pattern,
		}
	}
	return finder
}

// BranchChecker determines the current branch and default branch of a repository at dir.
type BranchChecker interface {
	CurrentBranch(ctx context.Context, dir string) (string, error)
	ResolveDefaultBranch(ctx context.Context, dir string) (string, error)
}

// DefaultBranchFilterFinder filters repositories based on whether their current branch
// is the default branch (DefaultOnly) or not the default branch (NotDefaultOnly).
type DefaultBranchFilterFinder struct {
	Inner          Finder
	Checker        BranchChecker
	DefaultOnly    bool
	NotDefaultOnly bool
}

func (f *DefaultBranchFilterFinder) Find(ctx context.Context) ([]string, error) {
	paths, err := f.Inner.Find(ctx)
	if err != nil {
		return nil, err
	}
	if !f.DefaultOnly && !f.NotDefaultOnly {
		return paths, nil
	}

	var res []string
	for _, p := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		current, err := f.Checker.CurrentBranch(ctx, p)
		if err != nil {
			return nil, err
		}
		def, err := f.Checker.ResolveDefaultBranch(ctx, p)
		if err != nil {
			return nil, err
		}
		isDefault := current != "" && def != "" && current == def
		if f.DefaultOnly && isDefault {
			res = append(res, p)
		} else if f.NotDefaultOnly && !isDefault {
			res = append(res, p)
		}
	}
	return res, nil
}

// NewDefaultBranchFilterFinder wraps finder with a DefaultBranchFilterFinder if defaultOnly or notDefaultOnly is true.
func NewDefaultBranchFilterFinder(finder Finder, checker BranchChecker, defaultOnly, notDefaultOnly bool) Finder {
	if defaultOnly || notDefaultOnly {
		return &DefaultBranchFilterFinder{
			Inner:          finder,
			Checker:        checker,
			DefaultOnly:    defaultOnly,
			NotDefaultOnly: notDefaultOnly,
		}
	}
	return finder
}

// BranchFilterFinder filters repositories based on whether their current branch matches Pattern.
type BranchFilterFinder struct {
	Inner   Finder
	Checker BranchChecker
	Pattern *regexp.Regexp
}

func (f *BranchFilterFinder) Find(ctx context.Context) ([]string, error) {
	paths, err := f.Inner.Find(ctx)
	if err != nil {
		return nil, err
	}
	if f.Pattern == nil {
		return paths, nil
	}

	var res []string
	for _, p := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		current, err := f.Checker.CurrentBranch(ctx, p)
		if err != nil {
			return nil, err
		}
		if current != "" && f.Pattern.MatchString(current) {
			res = append(res, p)
		}
	}
	return res, nil
}

// NewBranchFilterFinder wraps finder with a BranchFilterFinder if pattern is not nil.
func NewBranchFilterFinder(finder Finder, checker BranchChecker, pattern *regexp.Regexp) Finder {
	if pattern != nil {
		return &BranchFilterFinder{
			Inner:   finder,
			Checker: checker,
			Pattern: pattern,
		}
	}
	return finder
}
