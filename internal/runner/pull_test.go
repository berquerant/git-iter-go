package runner_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type customFakeRunner struct {
	*testutil.SubcommandRunner
}

func newCustomFakeRunner() *customFakeRunner {
	return &customFakeRunner{
		SubcommandRunner: testutil.NewSubcommandRunner(),
	}
}

func (c *customFakeRunner) on(subcmd string, h func(cmd []string, stdout, stderr io.Writer) error) {
	c.On(subcmd, h)
}

func (c *customFakeRunner) getCalls() [][]string {
	return c.Calls()
}

func TestPullRunner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		switchDefault bool
		switchBack    bool
		pullFlags     []string
		currentBranch string
		originHead    string // e.g. "origin/main", or "" if fails
		initBranch    string // config init.defaultBranch, or ""
		mainExists    bool
		masterExists  bool
		pullErr       error
		wantCalls     [][]string
		wantErr       bool
	}{
		{
			name:          "simple pull without flags",
			switchDefault: false,
			switchBack:    false,
			pullFlags:     nil,
			wantCalls: [][]string{
				{"git", "pull"},
			},
		},
		{
			name:          "pull with flags forwarded",
			switchDefault: false,
			switchBack:    false,
			pullFlags:     []string{"--rebase", "--autostash"},
			wantCalls: [][]string{
				{"git", "pull", "--rebase", "--autostash"},
			},
		},
		{
			name:          "switch default when already on default branch",
			switchDefault: true,
			switchBack:    false,
			currentBranch: "main",
			originHead:    "origin/main",
			wantCalls: [][]string{
				{"git", "branch", "--show-current"},
				{"git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD"},
				{"git", "pull"},
			},
		},
		{
			name:          "switch default from feature branch (via origin/HEAD)",
			switchDefault: true,
			switchBack:    false,
			currentBranch: "feature",
			originHead:    "origin/main",
			wantCalls: [][]string{
				{"git", "branch", "--show-current"},
				{"git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD"},
				{"git", "checkout", "main"},
				{"git", "pull"},
			},
		},
		{
			name:          "switch default and switch back from feature branch",
			switchDefault: true,
			switchBack:    true,
			currentBranch: "feat-x",
			originHead:    "origin/develop",
			wantCalls: [][]string{
				{"git", "branch", "--show-current"},
				{"git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD"},
				{"git", "checkout", "develop"},
				{"git", "pull"},
				{"git", "checkout", "feat-x"},
			},
		},
		{
			name:          "switch default and switch back when already on default branch does nothing extra",
			switchDefault: true,
			switchBack:    true,
			currentBranch: "main",
			originHead:    "origin/main",
			wantCalls: [][]string{
				{"git", "branch", "--show-current"},
				{"git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD"},
				{"git", "pull"},
			},
		},
		{
			name:          "default branch fallback to init.defaultBranch",
			switchDefault: true,
			switchBack:    false,
			currentBranch: "feat",
			originHead:    "", // fails
			initBranch:    "trunk",
			wantCalls: [][]string{
				{"git", "branch", "--show-current"},
				{"git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD"},
				{"git", "config", "--get", "init.defaultBranch"},
				{"git", "checkout", "trunk"},
				{"git", "pull"},
			},
		},
		{
			name:          "default branch fallback to main existence",
			switchDefault: true,
			switchBack:    false,
			currentBranch: "feat",
			originHead:    "",
			initBranch:    "",
			mainExists:    true,
			wantCalls: [][]string{
				{"git", "branch", "--show-current"},
				{"git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD"},
				{"git", "config", "--get", "init.defaultBranch"},
				{"git", "show-ref", "--verify", "--quiet", "refs/heads/main"},
				{"git", "checkout", "main"},
				{"git", "pull"},
			},
		},
		{
			name:          "default branch fallback to master existence",
			switchDefault: true,
			switchBack:    false,
			currentBranch: "feat",
			originHead:    "",
			initBranch:    "",
			mainExists:    false,
			masterExists:  true,
			wantCalls: [][]string{
				{"git", "branch", "--show-current"},
				{"git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD"},
				{"git", "config", "--get", "init.defaultBranch"},
				{"git", "show-ref", "--verify", "--quiet", "refs/heads/main"},
				{"git", "show-ref", "--verify", "--quiet", "refs/heads/master"},
				{"git", "checkout", "master"},
				{"git", "pull"},
			},
		},
		{
			name:          "unable to determine default branch returns error",
			switchDefault: true,
			switchBack:    false,
			currentBranch: "feat",
			originHead:    "",
			initBranch:    "",
			mainExists:    false,
			masterExists:  false,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fake := newCustomFakeRunner()
			fake.on("branch", func(_ []string, stdout, _ io.Writer) error {
				_, _ = fmt.Fprintln(stdout, tt.currentBranch)
				return nil
			})
			fake.on("symbolic-ref", func(_ []string, stdout, _ io.Writer) error {
				if tt.originHead != "" {
					_, _ = fmt.Fprintln(stdout, tt.originHead)
					return nil
				}
				return fmt.Errorf("no symbolic-ref")
			})
			fake.on("config", func(_ []string, stdout, _ io.Writer) error {
				if tt.initBranch != "" {
					_, _ = fmt.Fprintln(stdout, tt.initBranch)
					return nil
				}
				return fmt.Errorf("no config")
			})
			fake.on("show-ref", func(cmd []string, _, _ io.Writer) error {
				if slices.Contains(cmd, "refs/heads/main") && tt.mainExists {
					return nil
				}
				if slices.Contains(cmd, "refs/heads/master") && tt.masterExists {
					return nil
				}
				return fmt.Errorf("ref not found")
			})
			fake.on("pull", func(_ []string, stdout, _ io.Writer) error {
				if tt.pullErr != nil {
					return tt.pullErr
				}
				_, _ = fmt.Fprintln(stdout, "Already up to date.")
				return nil
			})

			p := &runner.PullRunner{
				SwitchDefault: tt.switchDefault,
				SwitchBack:    tt.switchBack,
				PullFlags:     tt.pullFlags,
				Inner:         fake,
			}

			var stdout, stderr bytes.Buffer
			err := p.Run(context.Background(), "/repo", nil, &stdout, &stderr)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantCalls, fake.getCalls())
		})
	}
}

func TestPullRunner_CustomGitBin(t *testing.T) {
	t.Parallel()

	fake := newCustomFakeRunner()
	fake.on("pull", func(_ []string, stdout, _ io.Writer) error {
		_, _ = fmt.Fprintln(stdout, "ok")
		return nil
	})

	p := &runner.PullRunner{
		GitCommand: "/custom/bin/git",
		Inner:      fake,
	}

	var stdout, stderr bytes.Buffer
	err := p.Run(context.Background(), "/repo", nil, &stdout, &stderr)
	require.NoError(t, err)

	calls := fake.getCalls()
	require.Len(t, calls, 1)
	assert.Equal(t, []string{"/custom/bin/git", "pull"}, calls[0])
	assert.True(t, strings.Contains(stdout.String(), "ok"))
}
