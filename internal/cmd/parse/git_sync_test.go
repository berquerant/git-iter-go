package parse_test

import (
	"testing"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/stretchr/testify/assert"
)

func TestParsePullArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		args         []string
		wantPatterns []string
		wantFlags    []string
	}{
		{
			name:         "empty args",
			args:         []string{},
			wantPatterns: []string{},
			wantFlags:    nil,
		},
		{
			name:         "only patterns",
			args:         []string{"repo1", "repo2"},
			wantPatterns: []string{"repo1", "repo2"},
			wantFlags:    nil,
		},
		{
			name:         "patterns and flags with dash",
			args:         []string{"repo1", "--", "--rebase", "--autostash"},
			wantPatterns: []string{"repo1"},
			wantFlags:    []string{"--rebase", "--autostash"},
		},
		{
			name:         "only flags with dash",
			args:         []string{"--", "--ff-only"},
			wantPatterns: []string{},
			wantFlags:    []string{"--ff-only"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parse.ParsePullArgs(tt.args)
			assert.Equal(t, tt.wantPatterns, got.Patterns)
			assert.Equal(t, tt.wantFlags, got.PullFlags)
		})
	}
}

func TestParsePullArgsWithDash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		args         []string
		dashIndex    int
		wantPatterns []string
		wantFlags    []string
	}{
		{
			name:         "dashIndex splits cleanly",
			args:         []string{"repo1", "--rebase"},
			dashIndex:    1,
			wantPatterns: []string{"repo1"},
			wantFlags:    []string{"--rebase"},
		},
		{
			name:         "dashIndex at start",
			args:         []string{"--ff-only"},
			dashIndex:    0,
			wantPatterns: []string{},
			wantFlags:    []string{"--ff-only"},
		},
		{
			name:         "dashIndex negative falls back to literal dash",
			args:         []string{"repo1", "--", "--no-rebase"},
			dashIndex:    -1,
			wantPatterns: []string{"repo1"},
			wantFlags:    []string{"--no-rebase"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parse.ParsePullArgsWithDash(tt.args, tt.dashIndex)
			assert.Equal(t, tt.wantPatterns, got.Patterns)
			assert.Equal(t, tt.wantFlags, got.PullFlags)
		})
	}
}

func TestParseFetchArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		args         []string
		wantPatterns []string
		wantFlags    []string
	}{
		{
			name:         "empty args",
			args:         []string{},
			wantPatterns: []string{},
			wantFlags:    nil,
		},
		{
			name:         "only patterns",
			args:         []string{"repo1", "repo2"},
			wantPatterns: []string{"repo1", "repo2"},
			wantFlags:    nil,
		},
		{
			name:         "patterns and flags with dash",
			args:         []string{"repo1", "--", "--prune", "--tags"},
			wantPatterns: []string{"repo1"},
			wantFlags:    []string{"--prune", "--tags"},
		},
		{
			name:         "only flags with dash",
			args:         []string{"--", "--all"},
			wantPatterns: []string{},
			wantFlags:    []string{"--all"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parse.ParseFetchArgs(tt.args)
			assert.Equal(t, tt.wantPatterns, got.Patterns)
			assert.Equal(t, tt.wantFlags, got.FetchFlags)
		})
	}
}

func TestParseFetchArgsWithDash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		args         []string
		dashIndex    int
		wantPatterns []string
		wantFlags    []string
	}{
		{
			name:         "dashIndex splits cleanly",
			args:         []string{"repo1", "--prune"},
			dashIndex:    1,
			wantPatterns: []string{"repo1"},
			wantFlags:    []string{"--prune"},
		},
		{
			name:         "dashIndex at start",
			args:         []string{"--all"},
			dashIndex:    0,
			wantPatterns: []string{},
			wantFlags:    []string{"--all"},
		},
		{
			name:         "dashIndex negative falls back to literal dash",
			args:         []string{"repo1", "--", "--depth=1"},
			dashIndex:    -1,
			wantPatterns: []string{"repo1"},
			wantFlags:    []string{"--depth=1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parse.ParseFetchArgsWithDash(tt.args, tt.dashIndex)
			assert.Equal(t, tt.wantPatterns, got.Patterns)
			assert.Equal(t, tt.wantFlags, got.FetchFlags)
		})
	}
}
