package runner_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/berquerant/git-iter-go/internal/cmd/parse"
	"github.com/berquerant/git-iter-go/internal/cmd/runner"
	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/output"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		outMode      string
		paths        []string
		wantStdout   string
		validateJSON bool
	}{
		{
			name:       "text mode",
			outMode:    "text",
			paths:      []string{"/repos/a", "/repos/b"},
			wantStdout: "/repos/a\n/repos/b\n",
		},
		{
			name:         "json mode",
			outMode:      "json",
			paths:        []string{"/repos/org/repo1"},
			validateJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &config.Config{
				ReposRoot: "/repos",
			}
			finder := &testutil.FakeFinder{Paths: tt.paths}
			var stdout bytes.Buffer

			err := runner.RunList(context.Background(), cfg, tt.outMode, parse.ListArgs{}, finder, &stdout)
			require.NoError(t, err)

			if tt.validateJSON {
				line := strings.TrimSpace(stdout.String())
				var res output.RepoResult
				require.NoError(t, json.Unmarshal([]byte(line), &res))
				assert.Equal(t, "/repos/org/repo1", res.RepoAbsPath)
				assert.Equal(t, "org/repo1", res.RepoRelPath)
			} else {
				assert.Equal(t, tt.wantStdout, stdout.String())
			}
		})
	}
}

func TestRunList_Markdown(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		ReposRoot:           "/repos",
		MarkdownHeaderLevel: 2,
	}
	finder := &testutil.FakeFinder{Paths: []string{"/repos/a", "/repos/b"}}
	var stdout bytes.Buffer

	err := runner.RunList(context.Background(), cfg, "markdown", parse.ListArgs{}, finder, &stdout)
	require.NoError(t, err)

	want := "# List: root = /repos\n\n- a\n- b\n\n---\n\n## Summary\n\n- **Total Repositories**: 2\n\n"
	assert.Equal(t, want, stdout.String())
}
