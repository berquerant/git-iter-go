package repo_test

import (
	"context"
	"errors"
	"testing"

	"github.com/berquerant/git-iter-go/internal/repo"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortedFinder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		paths   []string
		reverse bool
		want    []string
		wantErr bool
	}{
		{
			name:    "empty paths",
			paths:   nil,
			reverse: false,
			want:    nil,
		},
		{
			name:    "single path",
			paths:   []string{"/a"},
			reverse: false,
			want:    []string{"/a"},
		},
		{
			name:    "sort ascending",
			paths:   []string{"/c", "/a", "/b"},
			reverse: false,
			want:    []string{"/a", "/b", "/c"},
		},
		{
			name:    "sort descending (reverse)",
			paths:   []string{"/c", "/a", "/b"},
			reverse: true,
			want:    []string{"/c", "/b", "/a"},
		},
		{
			name:    "inner error propagated",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var fake testutil.FakeFinder
			if tt.wantErr {
				fake.Err = assert.AnError
			} else {
				fake.Paths = tt.paths
			}
			f := &repo.SortedFinder{
				Inner:   &fake,
				Reverse: tt.reverse,
			}
			got, err := f.Find(context.Background())
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewSortedFinder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		paths    []string
		sortAsc  bool
		sortDesc bool
		want     []string
	}{
		{
			name:     "neither sort specified returns original order",
			paths:    []string{"/c", "/a", "/b"},
			sortAsc:  false,
			sortDesc: false,
			want:     []string{"/c", "/a", "/b"},
		},
		{
			name:     "sortAsc true sorts ascending",
			paths:    []string{"/c", "/a", "/b"},
			sortAsc:  true,
			sortDesc: false,
			want:     []string{"/a", "/b", "/c"},
		},
		{
			name:     "sortDesc true sorts descending",
			paths:    []string{"/c", "/a", "/b"},
			sortAsc:  false,
			sortDesc: true,
			want:     []string{"/c", "/b", "/a"},
		},
		{
			name:     "both sortAsc and sortDesc true prioritizes sortDesc",
			paths:    []string{"/c", "/a", "/b"},
			sortAsc:  true,
			sortDesc: true,
			want:     []string{"/c", "/b", "/a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base := &testutil.FakeFinder{Paths: tt.paths}
			f := repo.NewSortedFinder(base, tt.sortAsc, tt.sortDesc)
			got, err := f.Find(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLimitedFinder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		paths   []string
		limit   int
		err     error
		want    []string
		wantErr bool
	}{
		{
			name:  "limit less than total",
			paths: []string{"/a", "/b", "/c", "/d"},
			limit: 2,
			want:  []string{"/a", "/b"},
		},
		{
			name:  "limit equal to total",
			paths: []string{"/a", "/b"},
			limit: 2,
			want:  []string{"/a", "/b"},
		},
		{
			name:  "limit greater than total",
			paths: []string{"/a", "/b"},
			limit: 5,
			want:  []string{"/a", "/b"},
		},
		{
			name:  "limit 0 returns all",
			paths: []string{"/a", "/b"},
			limit: 0,
			want:  []string{"/a", "/b"},
		},
		{
			name:  "negative limit returns all",
			paths: []string{"/a", "/b"},
			limit: -1,
			want:  []string{"/a", "/b"},
		},
		{
			name:    "inner error propagated",
			err:     errors.New("inner error"),
			limit:   2,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fake := testutil.FakeFinder{Paths: tt.paths, Err: tt.err}
			f := &repo.LimitedFinder{
				Inner: &fake,
				Limit: tt.limit,
			}
			got, err := f.Find(context.Background())
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewLimitedFinder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		paths []string
		limit int
		want  []string
	}{
		{
			name:  "limit <= 0 returns original finder without truncation",
			paths: []string{"/a", "/b", "/c"},
			limit: 0,
			want:  []string{"/a", "/b", "/c"},
		},
		{
			name:  "limit > 0 truncates paths",
			paths: []string{"/a", "/b", "/c"},
			limit: 2,
			want:  []string{"/a", "/b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base := &testutil.FakeFinder{Paths: tt.paths}
			f := repo.NewLimitedFinder(base, tt.limit)
			got, err := f.Find(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
