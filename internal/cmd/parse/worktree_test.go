package parse_test

import (
	"testing"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseWorktreeListArgs(t *testing.T) {
	t.Parallel()

	t.Run("defaults all to false", func(t *testing.T) {
		t.Parallel()
		got := parse.ParseWorktreeListArgs([]string{"org", "repo"}, true)
		assert.Equal(t, parse.WorktreeListArgs{
			Patterns: []string{"org", "repo"},
			Repo:     true,
			All:      false,
		}, got)
	})

	t.Run("with all true", func(t *testing.T) {
		t.Parallel()
		got := parse.ParseWorktreeListArgs([]string{"org", "repo"}, false, true)
		assert.Equal(t, parse.WorktreeListArgs{
			Patterns: []string{"org", "repo"},
			Repo:     false,
			All:      true,
		}, got)
	})
}

func TestParseWorktreeAddArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		newBranch string
		want      parse.WorktreeAddArgs
		wantTarget string
		wantErr   bool
	}{
		{
			name:       "existing branch",
			args:       []string{"feat-1"},
			want:       parse.WorktreeAddArgs{Branch: "feat-1"},
			wantTarget: "feat-1",
		},
		{
			name:       "new branch from base",
			args:       []string{"main"},
			newBranch:  "feat-2",
			want:       parse.WorktreeAddArgs{Branch: "main", NewBranch: "feat-2"},
			wantTarget: "feat-2",
		},
		{
			name:    "missing branch arg",
			args:    nil,
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"a", "b"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parse.ParseWorktreeAddArgs(tt.args, tt.newBranch)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantTarget, got.TargetBranch())
		})
	}
}

func TestParseWorktreeRemoveArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		args              []string
		force             bool
		noForce           bool
		deleteBranch      bool
		forceDeleteBranch bool
		want              parse.WorktreeRemoveArgs
		wantErr           bool
	}{
		{
			name:  "valid remove with force",
			args:  []string{"feat-1"},
			force: true,
			want:  parse.WorktreeRemoveArgs{Branch: "feat-1", Force: true},
		},
		{
			name:         "valid remove with delete branch",
			args:         []string{"feat-1"},
			deleteBranch: true,
			want:         parse.WorktreeRemoveArgs{Branch: "feat-1", DeleteBranch: true},
		},
		{
			name:              "valid remove with force delete branch (-D)",
			args:              []string{"feat-1"},
			forceDeleteBranch: true,
			want: parse.WorktreeRemoveArgs{
				Branch:            "feat-1",
				Force:             true,
				DeleteBranch:      true,
				ForceDeleteBranch: true,
			},
		},
		{
			name:    "missing branch",
			args:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parse.ParseWorktreeRemoveArgs(tt.args, tt.force, tt.noForce, tt.deleteBranch, tt.forceDeleteBranch)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}


func TestParseWorktreePruneArgs(t *testing.T) {
	t.Parallel()

	got := parse.ParseWorktreePruneArgs([]string{"org"}, true, true)
	assert.Equal(t, parse.WorktreePruneArgs{
		Patterns: []string{"org"},
		DryRun:   true,
		All:      true,
	}, got)
}
