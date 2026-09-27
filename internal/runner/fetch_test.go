package runner_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchRunner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		defaultBranch bool
		fetchFlags    []string
		originHead    string
		wantCalls     [][]string
		wantErr       bool
	}{
		{
			name:          "simple fetch without flags",
			defaultBranch: false,
			fetchFlags:    nil,
			wantCalls: [][]string{
				{"git", "fetch"},
			},
		},
		{
			name:          "fetch with flags forwarded",
			defaultBranch: false,
			fetchFlags:    []string{"--prune", "--tags"},
			wantCalls: [][]string{
				{"git", "fetch", "--prune", "--tags"},
			},
		},
		{
			name:          "fetch default branch",
			defaultBranch: true,
			fetchFlags:    nil,
			originHead:    "origin/main",
			wantCalls: [][]string{
				{"git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD"},
				{"git", "fetch", "origin", "main"},
			},
		},
		{
			name:          "fetch default branch with flags",
			defaultBranch: true,
			fetchFlags:    []string{"--prune"},
			originHead:    "origin/master",
			wantCalls: [][]string{
				{"git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD"},
				{"git", "fetch", "--prune", "origin", "master"},
			},
		},
		{
			name:          "fetch default branch fails when branch cannot be resolved",
			defaultBranch: true,
			originHead:    "", // no default branch
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fake := newCustomFakeRunner()
			fake.on("symbolic-ref", func(_ []string, stdout, _ io.Writer) error {
				if tt.originHead != "" {
					fmt.Fprintln(stdout, tt.originHead)
					return nil
				}
				return fmt.Errorf("no ref")
			})
			fake.on("config", func(_ []string, _, _ io.Writer) error {
				return fmt.Errorf("no config")
			})
			fake.on("show-ref", func(_ []string, _, _ io.Writer) error {
				return fmt.Errorf("no ref")
			})

			f := &runner.FetchRunner{
				Default:    tt.defaultBranch,
				FetchFlags: tt.fetchFlags,
				Inner:      fake,
			}

			var stdout, stderr bytes.Buffer
			err := f.Run(context.Background(), "/repo", nil, &stdout, &stderr)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantCalls, fake.getCalls())
		})
	}
}
