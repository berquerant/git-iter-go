package runner_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestRunGet(t *testing.T) {
	t.Parallel()

	t.Run("no URI lists repositories", func(t *testing.T) {
		t.Parallel()
		cfg := &config.Config{ReposRoot: "/repos"}
		finder := &testutil.FakeFinder{Paths: []string{"/repos/a", "/repos/b"}}
		r := &testutil.FakeRunner{}
		var stdout, stderr bytes.Buffer

		err := runner.RunGet(context.Background(), cfg, "text", parse.GetArgs{}, finder, r, &stdout, &stderr)
		require.NoError(t, err)
		assert.Equal(t, "/repos/a\n/repos/b\n", stdout.String())
		assert.Empty(t, r.Calls)
	})

	t.Run("existing repository does not clone", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		existing := testutil.MakeGitRepo(t, root, filepath.Join("github.com", "berquerant", "git-iter-go"))

		cfg := &config.Config{ReposRoot: root}
		finder := &testutil.FakeFinder{}
		r := &testutil.FakeRunner{}
		var stdout, stderr bytes.Buffer

		err := runner.RunGet(
			context.Background(),
			cfg,
			"text",
			parse.GetArgs{URI: "https://github.com/berquerant/git-iter-go.git"},
			finder,
			r,
			&stdout,
			&stderr,
		)
		require.NoError(t, err)
		assert.Equal(t, existing+"\n", stdout.String())
		assert.Empty(t, r.Calls)
	})

	t.Run("missing repository clones and outputs path to stdout while clone logs go to stderr", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		wantDir := filepath.Join(root, "github.com", "berquerant", "git-iter-go")

		cfg := &config.Config{
			ReposRoot: root,
			Timeout:   5 * time.Second,
		}
		finder := &testutil.FakeFinder{}
		r := &testutil.FakeRunner{Output: "Cloning into...\n"}
		var stdout, stderr bytes.Buffer

		err := runner.RunGet(
			context.Background(),
			cfg,
			"text",
			parse.GetArgs{URI: "git@github.com:berquerant/git-iter-go.git"},
			finder,
			r,
			&stdout,
			&stderr,
		)
		require.NoError(t, err)
		assert.Equal(t, wantDir+"\n", stdout.String())
		assert.Equal(t, "Cloning into...\n", stderr.String())
		require.Len(t, r.Calls, 1)
		assert.Equal(t, []string{"git", "clone", "--", "git@github.com:berquerant/git-iter-go.git", wantDir}, r.Calls[0].Command)
	})

	t.Run("clone failure returns error", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		cfg := &config.Config{ReposRoot: root}
		finder := &testutil.FakeFinder{}
		r := &testutil.FakeRunner{Err: errors.New("clone failed")}
		var stdout, stderr bytes.Buffer

		err := runner.RunGet(
			context.Background(),
			cfg,
			"text",
			parse.GetArgs{URI: "https://github.com/berquerant/git-iter-go.git"},
			finder,
			r,
			&stdout,
			&stderr,
		)
		require.Error(t, err)
		assert.Empty(t, stdout.String())
	})

	t.Run("missing repos-root with URI returns error", func(t *testing.T) {
		t.Parallel()
		cfg := &config.Config{}
		finder := &testutil.FakeFinder{}
		r := &testutil.FakeRunner{}
		var stdout, stderr bytes.Buffer

		err := runner.RunGet(
			context.Background(),
			cfg,
			"text",
			parse.GetArgs{URI: "https://github.com/berquerant/git-iter-go.git"},
			finder,
			r,
			&stdout,
			&stderr,
		)
		require.Error(t, err)
	})
}

