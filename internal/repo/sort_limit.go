package repo

import (
	"context"
	"sort"
)

// SortedFinder wraps another Finder and sorts repository paths.
// If Reverse is true, paths are sorted in descending order; otherwise ascending.
type SortedFinder struct {
	Inner   Finder
	Reverse bool
}

func (f *SortedFinder) Find(ctx context.Context) ([]string, error) {
	paths, err := f.Inner.Find(ctx)
	if err != nil {
		return nil, err
	}
	if len(paths) <= 1 {
		return paths, nil
	}
	res := make([]string, len(paths))
	copy(res, paths)
	if f.Reverse {
		sort.Slice(res, func(i, j int) bool {
			return res[i] > res[j]
		})
	} else {
		sort.Strings(res)
	}
	return res, nil
}

// NewSortedFinder wraps finder with a SortedFinder if sortAsc or sortDesc is true.
// If sortDesc is true, it takes precedence (or sorts in reverse order).
// If neither is true, finder is returned as-is.
func NewSortedFinder(finder Finder, sortAsc, sortDesc bool) Finder {
	if sortDesc {
		return &SortedFinder{Inner: finder, Reverse: true}
	}
	if sortAsc {
		return &SortedFinder{Inner: finder, Reverse: false}
	}
	return finder
}

// LimitedFinder wraps another Finder and limits the number of returned repository paths
// to at most Limit paths from the beginning.
type LimitedFinder struct {
	Inner Finder
	Limit int
}

func (f *LimitedFinder) Find(ctx context.Context) ([]string, error) {
	paths, err := f.Inner.Find(ctx)
	if err != nil {
		return nil, err
	}
	if f.Limit > 0 && len(paths) > f.Limit {
		return paths[:f.Limit], nil
	}
	return paths, nil
}

// NewLimitedFinder wraps finder with a LimitedFinder if limit > 0.
// If limit <= 0, finder is returned as-is.
func NewLimitedFinder(finder Finder, limit int) Finder {
	if limit > 0 {
		return &LimitedFinder{Inner: finder, Limit: limit}
	}
	return finder
}
