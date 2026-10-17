package common_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/berquerant/git-iter-go/internal/output/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExitCodeFrom(t *testing.T) {
	t.Parallel()

	// Run a real subprocess to get a genuine *exec.ExitError.
	cmd := exec.Command("sh", "-c", "exit 42")
	runErr := cmd.Run()
	require.Error(t, runErr)

	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil error", err: nil, want: 0},
		{name: "exec.ExitError code 42", err: runErr, want: 42},
		{name: "generic error", err: assert.AnError, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, common.ExitCodeFrom(tt.err))
		})
	}
}

func TestRelPath(t *testing.T) {
	t.Parallel()
	absCwd, err := filepath.Abs(".")
	require.NoError(t, err)

	tests := []struct {
		name        string
		reposRoot   string
		repoAbsPath string
		want        string
	}{
		{
			name:        "empty root returns absPath",
			reposRoot:   "",
			repoAbsPath: "/repos/org/repo",
			want:        "/repos/org/repo",
		},
		{
			name:        "nested path",
			reposRoot:   "/repos",
			repoAbsPath: "/repos/org/repo",
			want:        "org/repo",
		},
		{
			name:        "direct child",
			reposRoot:   "/repos",
			repoAbsPath: "/repos/myrepo",
			want:        "myrepo",
		},
		{
			name:        "root equals repoPath",
			reposRoot:   "/repos/r",
			repoAbsPath: "/repos/r",
			want:        ".",
		},
		{
			name:        "relative reposRoot with absolute repoAbsPath",
			reposRoot:   ".",
			repoAbsPath: filepath.Join(absCwd, "org", "repo"),
			want:        filepath.Join("org", "repo"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, common.RelPath(tt.reposRoot, tt.repoAbsPath))
		})
	}
}

