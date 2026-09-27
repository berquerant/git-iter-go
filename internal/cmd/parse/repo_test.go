package parse_test

import (
	"testing"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/stretchr/testify/assert"
)

func TestParseListArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		args         []string
		wantPatterns []string
	}{
		{
			name:         "empty args",
			args:         nil,
			wantPatterns: nil,
		},
		{
			name:         "single pattern",
			args:         []string{"myorg"},
			wantPatterns: []string{"myorg"},
		},
		{
			name:         "multiple patterns",
			args:         []string{"p1", "p2"},
			wantPatterns: []string{"p1", "p2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parse.ParseListArgs(tt.args)
			assert.Equal(t, tt.wantPatterns, got.Patterns)
		})
	}
}

func TestParseReadArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		args        []string
		wantCommand []string
	}{
		{
			name:        "simple command",
			args:        []string{"git", "status"},
			wantCommand: []string{"git", "status"},
		},
		{
			name:        "command with flags",
			args:        []string{"git", "log", "-1", "--oneline"},
			wantCommand: []string{"git", "log", "-1", "--oneline"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parse.ParseReadArgs(tt.args)
			assert.Equal(t, tt.wantCommand, got.Command)
		})
	}
}
