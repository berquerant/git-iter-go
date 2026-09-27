package parse_test

import (
	"testing"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDoArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		args         []string
		wantPatterns []string
		wantCommand  []string
		wantErr      bool
	}{
		{
			name:    "no dash and no command",
			args:    []string{"pat1", "pat2"},
			wantErr: true,
		},
		{
			name:    "dash but empty command",
			args:    []string{"pat1", "--"},
			wantErr: true,
		},
		{
			name:         "command only",
			args:         []string{"--", "git", "status"},
			wantPatterns: []string{},
			wantCommand:  []string{"git", "status"},
			wantErr:      false,
		},
		{
			name:         "patterns and command",
			args:         []string{"org/.*", "--", "git", "diff", "--stat"},
			wantPatterns: []string{"org/.*"},
			wantCommand:  []string{"git", "diff", "--stat"},
			wantErr:      false,
		},
		{
			name:         "multiple patterns",
			args:         []string{"p1", "p2", "--", "echo", "1"},
			wantPatterns: []string{"p1", "p2"},
			wantCommand:  []string{"echo", "1"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parse.ParseDoArgs(tt.args)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPatterns, got.Patterns)
			assert.Equal(t, tt.wantCommand, got.Command)
		})
	}
}
