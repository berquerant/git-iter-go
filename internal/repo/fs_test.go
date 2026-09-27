package repo_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/berquerant/git-iter-go/internal/repo"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeGitRepo delegates to testutil.MakeGitRepo.
func makeGitRepo(t *testing.T, parent, name string) string {
	return testutil.MakeGitRepo(t, parent, name)
}

func TestFilesystemFinder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		setup     func(root string)
		cancelled bool
		wantLen   int
		wantErr   error
	}{
		{
			name:    "empty root",
			setup:   func(_ string) {},
			wantLen: 0,
		},
		{
			name: "single repo",
			setup: func(root string) {
				makeGitRepo(t, root, "repo1")
			},
			wantLen: 1,
		},
		{
			name: "multiple repos",
			setup: func(root string) {
				makeGitRepo(t, root, "a")
				makeGitRepo(t, root, "b")
				makeGitRepo(t, root, "c")
			},
			wantLen: 3,
		},
		{
			name: "nested repos are not descended into",
			setup: func(root string) {
				makeGitRepo(t, root, "outer")
				// inner would be inside outer but since outer is a git repo,
				// WalkDir skips it.
				require.NoError(t, os.MkdirAll(filepath.Join(root, "outer", "inner", ".git"), 0o755))
			},
			wantLen: 1,
		},
		{
			name: "non-repo directories are ignored",
			setup: func(root string) {
				makeGitRepo(t, root, "repo")
				require.NoError(t, os.MkdirAll(filepath.Join(root, "notarepo"), 0o755))
			},
			wantLen: 1,
		},
		{
			name: "context cancelled immediately",
			setup: func(root string) {
				makeGitRepo(t, root, "a")
			},
			cancelled: true,
			wantErr:   context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			tt.setup(root)
			f := &repo.FilesystemFinder{Root: root}
			ctx := context.Background()
			if tt.cancelled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			got, err := f.Find(ctx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
		})
	}
}

func TestCommandFinder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		output  string
		wantErr bool
		want    []string
	}{
		{
			name:   "single path",
			output: "/home/user/repo\n",
			want:   []string{"/home/user/repo"},
		},
		{
			name:   "multiple paths",
			output: "/a\n/b\n/c\n",
			want:   []string{"/a", "/b", "/c"},
		},
		{
			name:   "blank lines ignored",
			output: "/a\n\n/b\n",
			want:   []string{"/a", "/b"},
		},
		{
			name:   "empty output",
			output: "",
			want:   nil,
		},
		{
			name:    "command error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := &repo.CommandFinder{
				Command: "unused",
				CommandFunc: func(_ context.Context, _ string) ([]byte, error) {
					if tt.wantErr {
						return nil, assert.AnError
					}
					return []byte(tt.output), nil
				},
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
