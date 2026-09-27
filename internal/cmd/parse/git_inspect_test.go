package parse_test

import (
	"testing"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/stretchr/testify/assert"
)

func TestParseLsFilesArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		args         []string
		wantPatterns []string
		wantFlags    []string
	}{
		{
			name:         "empty args",
			args:         nil,
			wantPatterns: nil,
			wantFlags:    nil,
		},
		{
			name:         "patterns only without dash",
			args:         []string{"repo1", "repo2"},
			wantPatterns: []string{"repo1", "repo2"},
			wantFlags:    nil,
		},
		{
			name:         "patterns and flags with literal dash",
			args:         []string{"repo1", "--", "-c", "--modified"},
			wantPatterns: []string{"repo1"},
			wantFlags:    []string{"-c", "--modified"},
		},
		{
			name:         "flags only with leading dash",
			args:         []string{"--", "-c"},
			wantPatterns: []string{},
			wantFlags:    []string{"-c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parse.ParseLsFilesArgs(tt.args)
			assert.Equal(t, tt.wantPatterns, got.Patterns)
			assert.Equal(t, tt.wantFlags, got.LsFilesFlags)
		})
	}
}

func TestParseLsFilesArgsWithDash(t *testing.T) {
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
			args:         []string{"repo1", "-c"},
			dashIndex:    1,
			wantPatterns: []string{"repo1"},
			wantFlags:    []string{"-c"},
		},
		{
			name:         "dashIndex at start",
			args:         []string{"-s"},
			dashIndex:    0,
			wantPatterns: []string{},
			wantFlags:    []string{"-s"},
		},
		{
			name:         "dashIndex negative falls back to literal dash",
			args:         []string{"repo1", "--", "--others"},
			dashIndex:    -1,
			wantPatterns: []string{"repo1"},
			wantFlags:    []string{"--others"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parse.ParseLsFilesArgsWithDash(tt.args, tt.dashIndex)
			assert.Equal(t, tt.wantPatterns, got.Patterns)
			assert.Equal(t, tt.wantFlags, got.LsFilesFlags)
		})
	}
}

func TestParseStatusArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		wantPatterns []string
		wantFlags    []string
	}{
		{
			name:         "no args",
			args:         []string{},
			wantPatterns: []string{},
			wantFlags:    nil,
		},
		{
			name:         "patterns only without dash",
			args:         []string{"repo1", "repo2"},
			wantPatterns: []string{"repo1", "repo2"},
			wantFlags:    nil,
		},
		{
			name:         "flags only after dash",
			args:         []string{"--", "--short", "--branch"},
			wantPatterns: []string{},
			wantFlags:    []string{"--short", "--branch"},
		},
		{
			name:         "patterns and flags",
			args:         []string{"repo1", "--", "--porcelain"},
			wantPatterns: []string{"repo1"},
			wantFlags:    []string{"--porcelain"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parse.ParseStatusArgs(tt.args)
			assert.Equal(t, tt.wantPatterns, got.Patterns)
			assert.Equal(t, tt.wantFlags, got.StatusFlags)
		})
	}
}

func TestParseStatusArgsWithDash(t *testing.T) {
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
			args:         []string{"repo1", "--short"},
			dashIndex:    1,
			wantPatterns: []string{"repo1"},
			wantFlags:    []string{"--short"},
		},
		{
			name:         "dashIndex at start",
			args:         []string{"--porcelain"},
			dashIndex:    0,
			wantPatterns: []string{},
			wantFlags:    []string{"--porcelain"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parse.ParseStatusArgsWithDash(tt.args, tt.dashIndex)
			assert.Equal(t, tt.wantPatterns, got.Patterns)
			assert.Equal(t, tt.wantFlags, got.StatusFlags)
		})
	}
}

func TestParseDiffArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		wantPatterns []string
		wantFlags    []string
	}{
		{
			name:         "no args",
			args:         []string{},
			wantPatterns: []string{},
			wantFlags:    nil,
		},
		{
			name:         "patterns only without dash",
			args:         []string{"repo1", "repo2"},
			wantPatterns: []string{"repo1", "repo2"},
			wantFlags:    nil,
		},
		{
			name:         "flags only after dash",
			args:         []string{"--", "--stat", "--cached"},
			wantPatterns: []string{},
			wantFlags:    []string{"--stat", "--cached"},
		},
		{
			name:         "patterns and flags",
			args:         []string{"repo1", "--", "main..HEAD"},
			wantPatterns: []string{"repo1"},
			wantFlags:    []string{"main..HEAD"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parse.ParseDiffArgs(tt.args)
			assert.Equal(t, tt.wantPatterns, got.Patterns)
			assert.Equal(t, tt.wantFlags, got.DiffFlags)
		})
	}
}

func TestParseDiffArgsWithDash(t *testing.T) {
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
			args:         []string{"repo1", "--stat"},
			dashIndex:    1,
			wantPatterns: []string{"repo1"},
			wantFlags:    []string{"--stat"},
		},
		{
			name:         "dashIndex at start",
			args:         []string{"--cached"},
			dashIndex:    0,
			wantPatterns: []string{},
			wantFlags:    []string{"--cached"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parse.ParseDiffArgsWithDash(tt.args, tt.dashIndex)
			assert.Equal(t, tt.wantPatterns, got.Patterns)
			assert.Equal(t, tt.wantFlags, got.DiffFlags)
		})
	}
}
