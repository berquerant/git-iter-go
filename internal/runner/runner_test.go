package runner_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrepRunner(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		repoPath   string
		absPath    bool
		root       string
		fakeOutput string
		want       string
		wantErr    bool
	}{
		{
			name:       "abs path prepends repoPath",
			repoPath:   "/home/user/repos/myrepo",
			absPath:    true,
			root:       "/home/user/repos",
			fakeOutput: "main.go:1:func main() {}",
			want:       "/home/user/repos/myrepo/main.go:1:func main() {}\n",
		},
		{
			name:       "relative path strips root prefix",
			repoPath:   "/home/user/repos/org/myrepo",
			absPath:    false,
			root:       "/home/user/repos",
			fakeOutput: "main.go:1:hello",
			want:       "org/myrepo/main.go:1:hello\n",
		},
		{
			name:       "empty root falls back to repoPath",
			repoPath:   "/abs/repo",
			absPath:    false,
			root:       "",
			fakeOutput: "file.go:2:world",
			want:       "/abs/repo/file.go:2:world\n",
		},
		{
			name:       "multiple output lines",
			repoPath:   "/root/r",
			absPath:    true,
			fakeOutput: "a.go:1:x\nb.go:2:y",
			want:       "/root/r/a.go:1:x\n/root/r/b.go:2:y\n",
		},
		{
			name:       "inner runner error is propagated",
			repoPath:   "/root/r",
			absPath:    true,
			fakeOutput: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fake := &testutil.FakeRunner{Output: tt.fakeOutput}
			if tt.wantErr {
				fake.Err = assert.AnError
			}
			gr := &runner.GrepRunner{
				Inner:    fake,
				RepoPath: tt.repoPath,
				AbsPath:  tt.absPath,
				Root:     tt.root,
			}
			var stdout bytes.Buffer
			err := gr.Run(context.Background(), tt.repoPath, []string{"git", "grep", "."}, &stdout, &bytes.Buffer{})
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, stdout.String())
		})
	}
}

func TestProcessRunner(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		cmd     []string
		wantErr bool
	}{
		{
			name:    "empty command returns error",
			cmd:     []string{},
			wantErr: true,
		},
		{
			name:    "nonexistent binary returns error",
			cmd:     []string{"nonexistent-command-12345"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := runner.ProcessRunner{}
			err := r.Run(context.Background(), t.TempDir(), tt.cmd, &bytes.Buffer{}, &bytes.Buffer{})
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProcessRunner_InheritEnvironment(t *testing.T) {
	// Note: Do not run t.Parallel() when modifying process environment via t.Setenv.
	const (
		envKey = "GIT_ITER_TEST_ENV"
		envVal = "custom-env-value-12345"
	)
	t.Setenv(envKey, envVal)

	r := runner.ProcessRunner{}
	var stdout, stderr bytes.Buffer
	err := r.Run(
		context.Background(),
		t.TempDir(),
		[]string{"sh", "-c", "printf %s \"$" + envKey + "\""},
		&stdout,
		&stderr,
	)
	require.NoError(t, err)
	assert.Equal(t, envVal, stdout.String())
}


