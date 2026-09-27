package parse_test

import (
	"testing"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/stretchr/testify/assert"
)

func TestSplitWithDash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		dashIndex    int
		wantPatterns []string
		wantCommand  []string
	}{
		{
			name:         "empty args",
			args:         []string{},
			dashIndex:    -1,
			wantPatterns: []string{},
			wantCommand:  nil,
		},
		{
			name:         "literal dash present",
			args:         []string{"repo1", "--", "git", "status"},
			dashIndex:    -1,
			wantPatterns: []string{"repo1"},
			wantCommand:  []string{"git", "status"},
		},
		{
			name:         "no dash at all",
			args:         []string{"repo1", "repo2"},
			dashIndex:    -1,
			wantPatterns: []string{"repo1", "repo2"},
			wantCommand:  nil,
		},
		{
			name:         "dashIndex provided by cobra",
			args:         []string{"repo1", "git", "status"},
			dashIndex:    1,
			wantPatterns: []string{"repo1"},
			wantCommand:  []string{"git", "status"},
		},
		{
			name:         "dashIndex at 0",
			args:         []string{"git", "status"},
			dashIndex:    0,
			wantPatterns: []string{},
			wantCommand:  []string{"git", "status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotPats, gotCmd := parse.SplitWithDash(tt.args, tt.dashIndex)
			assert.Equal(t, tt.wantPatterns, gotPats)
			assert.Equal(t, tt.wantCommand, gotCmd)
		})
	}
}
