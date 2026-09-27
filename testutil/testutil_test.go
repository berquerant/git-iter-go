package testutil_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFakeFinder(t *testing.T) {
	t.Parallel()

	f := &testutil.FakeFinder{
		Paths: []string{"/path/1", "/path/2"},
		Err:   nil,
	}

	paths, err := f.Find(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"/path/1", "/path/2"}, paths)

	errBoom := errors.New("boom")
	fErr := &testutil.FakeFinder{Err: errBoom}
	_, err = fErr.Find(context.Background())
	require.ErrorIs(t, err, errBoom)
}

func TestFakeRunner(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	r := &testutil.FakeRunner{
		Output: "some output",
	}

	err := r.Run(context.Background(), "/dir", []string{"echo", "hi"}, &buf, io.Discard)
	require.NoError(t, err)
	assert.Equal(t, "some output", buf.String())
	require.Len(t, r.Calls, 1)
	assert.Equal(t, "/dir", r.Calls[0].Dir)
	assert.Equal(t, []string{"echo", "hi"}, r.Calls[0].Command)
}

func TestMakeGitRepo(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	repoDir := testutil.MakeGitRepo(t, parent, "my-repo")

	assert.Equal(t, filepath.Join(parent, "my-repo"), repoDir)
	fi, err := os.Stat(filepath.Join(repoDir, ".git"))
	require.NoError(t, err)
	assert.True(t, fi.IsDir())
}

func TestSubcommandRunner(t *testing.T) {
	t.Parallel()

	r := testutil.NewSubcommandRunner()
	r.On("status", func(_ []string, stdout, _ io.Writer) error {
		_, _ = io.WriteString(stdout, "clean")
		return nil
	})

	var stdout bytes.Buffer
	err := r.Run(context.Background(), "/dir", []string{"git", "status", "--short"}, &stdout, io.Discard)
	require.NoError(t, err)
	assert.Equal(t, "clean", stdout.String())

	// Unhandled subcommand
	err = r.Run(context.Background(), "/dir", []string{"git", "diff"}, &stdout, io.Discard)
	require.NoError(t, err)

	// Short command (no subcommand)
	err = r.Run(context.Background(), "/dir", []string{"git"}, &stdout, io.Discard)
	require.NoError(t, err)

	calls := r.Calls()
	require.Len(t, calls, 3)
	assert.Equal(t, []string{"git", "status", "--short"}, calls[0])
	assert.Equal(t, []string{"git", "diff"}, calls[1])
	assert.Equal(t, []string{"git"}, calls[2])
}
