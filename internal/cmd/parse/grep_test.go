package parse_test

import (
	"testing"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGrepArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		args         []string
		wantPatterns []string
		wantGrepArgs []string
		wantErr      bool
	}{
		{
			name:    "no dash and no grep args",
			args:    []string{"pat1"},
			wantErr: true,
		},
		{
			name:    "dash but empty grep args",
			args:    []string{"pat1", "--"},
			wantErr: true,
		},
		{
			name:         "grep args only",
			args:         []string{"--", "TODO", "-n"},
			wantPatterns: []string{},
			wantGrepArgs: []string{"TODO", "-n"},
			wantErr:      false,
		},
		{
			name:         "patterns and grep args",
			args:         []string{"repo1", "repo2", "--", "FIXME"},
			wantPatterns: []string{"repo1", "repo2"},
			wantGrepArgs: []string{"FIXME"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parse.ParseGrepArgs(tt.args)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPatterns, got.Patterns)
			assert.Equal(t, tt.wantGrepArgs, got.GitGrepArgs)
		})
	}
}
