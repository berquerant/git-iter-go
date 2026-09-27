package repo_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/berquerant/git-iter-go/internal/repo"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilteredFinder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		paths   []string
		pattern *regexp.Regexp
		want    []string
	}{
		{
			name:    "nil pattern passes all",
			paths:   []string{"/a", "/b"},
			pattern: nil,
			want:    []string{"/a", "/b"},
		},
		{
			name:    "pattern filters",
			paths:   []string{"/home/foo", "/home/bar", "/opt/baz"},
			pattern: regexp.MustCompile("foo|baz"),
			want:    []string{"/home/foo", "/opt/baz"},
		},
		{
			name:    "no match returns empty",
			paths:   []string{"/a", "/b"},
			pattern: regexp.MustCompile("zzz"),
			want:    nil,
		},
		{
			name:    "empty input",
			paths:   nil,
			pattern: regexp.MustCompile("foo"),
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := &repo.FilteredFinder{
				Inner:   &testutil.FakeFinder{Paths: tt.paths},
				Pattern: tt.pattern,
			}
			got, err := f.Find(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewFilteredFinder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		paths    []string
		patterns []string
		want     []string
		wantErr  bool
	}{
		{
			name:     "no patterns returns all",
			paths:    []string{"/a", "/b"},
			patterns: nil,
			want:     []string{"/a", "/b"},
		},
		{
			name:     "single pattern",
			paths:    []string{"/home/foo", "/home/bar"},
			patterns: []string{"foo"},
			want:     []string{"/home/foo"},
		},
		{
			name:     "multiple patterns joined with OR",
			paths:    []string{"/home/foo", "/home/bar", "/home/baz"},
			patterns: []string{"foo", "baz"},
			want:     []string{"/home/foo", "/home/baz"},
		},
		{
			name:     "invalid regexp",
			patterns: []string{"[invalid"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base := &testutil.FakeFinder{Paths: tt.paths}
			f, err := repo.NewFilteredFinder(base, tt.patterns)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			got, err := f.Find(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

type fakeDirtyChecker struct {
	dirtyMap map[string]bool
	err      error
}

func (f *fakeDirtyChecker) IsDirty(_ context.Context, dir string) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return f.dirtyMap[dir], nil
}

func TestStatusFilterFinder(t *testing.T) {
	t.Parallel()

	paths := []string{"/repos/dirty1", "/repos/clean1", "/repos/dirty2", "/repos/clean2"}
	checker := &fakeDirtyChecker{
		dirtyMap: map[string]bool{
			"/repos/dirty1": true,
			"/repos/clean1": false,
			"/repos/dirty2": true,
			"/repos/clean2": false,
		},
	}

	tests := []struct {
		name      string
		dirtyOnly bool
		cleanOnly bool
		want      []string
	}{
		{
			name:      "dirty only filters to dirty repos",
			dirtyOnly: true,
			want:      []string{"/repos/dirty1", "/repos/dirty2"},
		},
		{
			name:      "clean only filters to clean repos",
			cleanOnly: true,
			want:      []string{"/repos/clean1", "/repos/clean2"},
		},
		{
			name: "neither passes all through",
			want: paths,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base := &testutil.FakeFinder{Paths: paths}
			f := repo.NewStatusFilterFinder(base, checker, tt.dirtyOnly, tt.cleanOnly)
			got, err := f.Find(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStatusFilterFinder_Errors(t *testing.T) {
	t.Parallel()

	t.Run("inner finder error", func(t *testing.T) {
		t.Parallel()
		base := &testutil.FakeFinder{Err: errors.New("find err")}
		f := &repo.StatusFilterFinder{Inner: base, DirtyOnly: true}
		_, err := f.Find(context.Background())
		require.Error(t, err)
	})

	t.Run("checker error", func(t *testing.T) {
		t.Parallel()
		base := &testutil.FakeFinder{Paths: []string{"/repo"}}
		checker := &fakeDirtyChecker{err: errors.New("check err")}
		f := &repo.StatusFilterFinder{Inner: base, Checker: checker, DirtyOnly: true}
		_, err := f.Find(context.Background())
		require.Error(t, err)
	})

	t.Run("context cancelled", func(t *testing.T) {
		t.Parallel()
		base := &testutil.FakeFinder{Paths: []string{"/repo"}}
		f := &repo.StatusFilterFinder{Inner: base, DirtyOnly: true}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := f.Find(ctx)
		require.Error(t, err)
	})
}

type fakeRemoteURLChecker struct {
	urls map[string]string
	err  error
}

func (f *fakeRemoteURLChecker) RemoteURL(_ context.Context, dir string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.urls[dir], nil
}

func TestRemoteURLFilterFinder(t *testing.T) {
	t.Parallel()

	paths := []string{"/repos/github", "/repos/gitlab", "/repos/norepmote"}
	checker := &fakeRemoteURLChecker{
		urls: map[string]string{
			"/repos/github":    "https://github.com/berquerant/git-iter-go.git",
			"/repos/gitlab":    "git@gitlab.com:foo/bar.git",
			"/repos/norepmote": "",
		},
	}

	tests := []struct {
		name    string
		pattern *regexp.Regexp
		want    []string
	}{
		{
			name:    "match github",
			pattern: regexp.MustCompile(`github\.com`),
			want:    []string{"/repos/github"},
		},
		{
			name:    "match gitlab or github",
			pattern: regexp.MustCompile(`github\.com|gitlab\.com`),
			want:    []string{"/repos/github", "/repos/gitlab"},
		},
		{
			name:    "no match",
			pattern: regexp.MustCompile(`bitbucket`),
			want:    nil,
		},
		{
			name:    "nil pattern passes all through",
			pattern: nil,
			want:    paths,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base := &testutil.FakeFinder{Paths: paths}
			f := repo.NewRemoteURLFilterFinder(base, checker, tt.pattern)
			got, err := f.Find(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRemoteURLFilterFinder_Errors(t *testing.T) {
	t.Parallel()

	t.Run("inner finder error", func(t *testing.T) {
		t.Parallel()
		base := &testutil.FakeFinder{Err: errors.New("find err")}
		f := &repo.RemoteURLFilterFinder{
			Inner:   base,
			Pattern: regexp.MustCompile("foo"),
		}
		_, err := f.Find(context.Background())
		require.Error(t, err)
	})

	t.Run("checker error", func(t *testing.T) {
		t.Parallel()
		base := &testutil.FakeFinder{Paths: []string{"/repo"}}
		checker := &fakeRemoteURLChecker{err: errors.New("check err")}
		f := &repo.RemoteURLFilterFinder{
			Inner:   base,
			Checker: checker,
			Pattern: regexp.MustCompile("foo"),
		}
		_, err := f.Find(context.Background())
		require.Error(t, err)
	})

	t.Run("context cancelled", func(t *testing.T) {
		t.Parallel()
		base := &testutil.FakeFinder{Paths: []string{"/repo"}}
		checker := &fakeRemoteURLChecker{urls: map[string]string{"/repo": "https://example.com"}}
		f := &repo.RemoteURLFilterFinder{
			Inner:   base,
			Checker: checker,
			Pattern: regexp.MustCompile("foo"),
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := f.Find(ctx)
		require.Error(t, err)
	})
}
