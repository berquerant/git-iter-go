package runner_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorktreePruneRunner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dryRun       bool
		worktrees    []string // branch names to create under worktree_root/org/repo/
		mergedBranch string
		setup        func(fake *customFakeRunner)
		wantRemoved  []string
		wantErr      bool
	}{
		{
			name:         "removes merged worktree and runs git worktree prune",
			dryRun:       false,
			worktrees:    []string{"feat-merged", "feat-unmerged", "main"},
			mergedBranch: "feat-merged",
			setup:        func(_ *customFakeRunner) {},
			wantRemoved:  []string{"feat-merged"},
		},
		{
			name:         "dry-run does not execute worktree remove and passes --dry-run to prune",
			dryRun:       true,
			worktrees:    []string{"feat-merged"},
			mergedBranch: "feat-merged",
			setup:        func(_ *customFakeRunner) {},
			wantRemoved:  nil,
		},
		{
			name:      "no worktree directory for repo still runs git worktree prune",
			dryRun:    false,
			worktrees: nil,
			setup:     func(_ *customFakeRunner) {},
		},
		{
			name:         "propagates error when worktree remove fails",
			dryRun:       false,
			worktrees:    []string{"feat-merged"},
			mergedBranch: "feat-merged",
			setup: func(fake *customFakeRunner) {
				fake.on("worktree", func(cmd []string, _, _ io.Writer) error {
					if slices.Contains(cmd, "remove") {
						return fmt.Errorf("remove failed")
					}
					return nil
				})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reposRoot := t.TempDir()
			wtRoot := filepath.Join(reposRoot, "git-worktree")
			repoDir := filepath.Join(reposRoot, "org", "repo")
			require.NoError(t, os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755))

			for _, br := range tt.worktrees {
				wtDir := filepath.Join(wtRoot, "org", "repo", br)
				require.NoError(t, os.MkdirAll(filepath.Join(wtDir, ".git"), 0o755))
			}

			fake := newCustomFakeRunner()
			fake.on("symbolic-ref", func(_ []string, stdout, _ io.Writer) error {
				_, _ = fmt.Fprintln(stdout, "origin/main")
				return nil
			})
			fake.on("branch", func(_ []string, _, _ io.Writer) error {
				// Let resolveWorktreeBranch fall back to relative path from worktree_root/org/repo.
				return fmt.Errorf("no git in tempdir")
			})
			fake.on("merge-base", func(cmd []string, _, _ io.Writer) error {
				if tt.mergedBranch != "" && slices.Contains(cmd, tt.mergedBranch) {
					return nil
				}
				return fmt.Errorf("not ancestor")
			})
			fake.on("rev-parse", func(_ []string, _, _ io.Writer) error {
				return fmt.Errorf("not squash merged")
			})
			tt.setup(fake)

			r := &runner.WorktreePruneRunner{
				GitCommand:   "git",
				ReposRoot:    reposRoot,
				WorktreeRoot: wtRoot,
				DryRun:       tt.dryRun,
				Inner:        fake,
			}

			var stdout, stderr bytes.Buffer
			err := r.Run(context.Background(), repoDir, nil, &stdout, &stderr)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			calls := fake.getCalls()
			for _, removed := range tt.wantRemoved {
				expectedWT := filepath.Join(wtRoot, "org", "repo", removed)
				assert.Contains(t, calls, []string{"git", "worktree", "remove", "--force", "--", expectedWT})
			}
			if tt.dryRun {
				assert.Contains(t, calls, []string{"git", "worktree", "prune", "--dry-run"})
			} else {
				assert.Contains(t, calls, []string{"git", "worktree", "prune"})
			}
		})
	}
}
