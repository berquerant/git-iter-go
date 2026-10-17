package git_test

import (
	"context"
	"fmt"
	"io"
	"slices"
	"testing"

	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRunner struct {
	*testutil.SubcommandRunner
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{
		SubcommandRunner: testutil.NewSubcommandRunner(),
	}
}

func (f *fakeRunner) on(subcmd string, h func(cmd []string, stdout, stderr io.Writer) error) {
	f.On(subcmd, h)
}

func TestCommand_Builders(t *testing.T) {
	t.Parallel()

	c := git.New("/usr/bin/git", nil)
	assert.Equal(t, "/usr/bin/git", c.GitBin())
	assert.Equal(t, []string{"/usr/bin/git", "grep", "-H", "foo", "bar"}, c.GrepCmd("foo", "bar"))
	assert.Equal(t, []string{"/usr/bin/git", "pull", "--rebase"}, c.PullCmd("--rebase"))
	assert.Equal(t, []string{"/usr/bin/git", "fetch", "--prune"}, c.FetchCmd("--prune"))
	assert.Equal(t, []string{"/usr/bin/git", "ls-files", "-c"}, c.LsFilesCmd("-c"))
	assert.Equal(t, []string{"/usr/bin/git", "diff", "--stat"}, c.DiffCmd(true, "--stat"))
	assert.Equal(t, []string{"/usr/bin/git", "--no-pager", "diff", "--stat"}, c.DiffCmd(false, "--stat"))
	assert.Equal(t, []string{"/usr/bin/git", "status", "--short"}, c.StatusCmd())
	assert.Equal(t, []string{"/usr/bin/git", "status", "--porcelain"}, c.StatusCmd("--porcelain"))
	assert.Equal(t, []string{"/usr/bin/git", "clone", "https://github.com/org/repo.git", "/repos/github.com/org/repo"}, c.CloneCmd("https://github.com/org/repo.git", "/repos/github.com/org/repo"))

	def := git.New("", nil)
	assert.Equal(t, git.DefaultCommand, def.GitBin())
}

func TestCommand_IsDirty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		output    string
		runErr    error
		wantDirty bool
		wantErr   bool
	}{
		{
			name:      "dirty with modified file",
			output:    " M file.go\n",
			wantDirty: true,
		},
		{
			name:      "clean",
			output:    "",
			wantDirty: false,
		},
		{
			name:    "error running git status",
			runErr:  fmt.Errorf("git error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fake := newFakeRunner()
			fake.on("status", func(_ []string, stdout, _ io.Writer) error {
				if tt.runErr != nil {
					return tt.runErr
				}
				_, _ = fmt.Fprint(stdout, tt.output)
				return nil
			})

			c := git.New("git", fake)
			dirty, err := c.IsDirty(context.Background(), "/repo")
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantDirty, dirty)
		})
	}
}

func TestCommand_CurrentBranch(t *testing.T) {
	t.Parallel()

	fake := newFakeRunner()
	fake.on("branch", func(_ []string, stdout, _ io.Writer) error {
		_, _ = fmt.Fprintln(stdout, "feat/issue-1")
		return nil
	})

	c := git.New("git", fake)
	branch, err := c.CurrentBranch(context.Background(), "/repo")
	require.NoError(t, err)
	assert.Equal(t, "feat/issue-1", branch)
	assert.Equal(t, [][]string{{"git", "branch", "--show-current"}}, fake.Calls())
}

func TestCommand_Checkout(t *testing.T) {
	t.Parallel()

	fake := newFakeRunner()
	c := git.New("git", fake)
	err := c.Checkout(context.Background(), "/repo", "main", io.Discard)
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"git", "checkout", "main"}}, fake.Calls())
}

func TestCommand_ResolveDefaultBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		originHead   string
		initBranch   string
		mainExists   bool
		masterExists bool
		wantBranch   string
		wantErr      bool
	}{
		{
			name:       "via origin/HEAD",
			originHead: "origin/main",
			wantBranch: "main",
		},
		{
			name:       "via init.defaultBranch",
			initBranch: "trunk",
			wantBranch: "trunk",
		},
		{
			name:       "fallback to main",
			mainExists: true,
			wantBranch: "main",
		},
		{
			name:         "fallback to master",
			masterExists: true,
			wantBranch:   "master",
		},
		{
			name:    "none exists errors",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fake := newFakeRunner()
			fake.on("symbolic-ref", func(_ []string, stdout, _ io.Writer) error {
				if tt.originHead != "" {
					_, _ = fmt.Fprintln(stdout, tt.originHead)
					return nil
				}
				return fmt.Errorf("no symbolic-ref")
			})
			fake.on("config", func(_ []string, stdout, _ io.Writer) error {
				if tt.initBranch != "" {
					_, _ = fmt.Fprintln(stdout, tt.initBranch)
					return nil
				}
				return fmt.Errorf("no config")
			})
			fake.on("show-ref", func(cmd []string, _, _ io.Writer) error {
				if slices.Contains(cmd, "refs/heads/main") && tt.mainExists {
					return nil
				}
				if slices.Contains(cmd, "refs/heads/master") && tt.masterExists {
					return nil
				}
				return fmt.Errorf("no ref")
			})

			c := git.New("git", fake)
			branch, err := c.ResolveDefaultBranch(context.Background(), "/repo")
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantBranch, branch)
		})
	}
}

func TestCommand_RemoteURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		output  string
		runErr  error
		wantURL string
	}{
		{
			name:    "origin remote url found",
			output:  "https://github.com/berquerant/git-iter-go.git\n",
			wantURL: "https://github.com/berquerant/git-iter-go.git",
		},
		{
			name:    "no origin remote url",
			runErr:  fmt.Errorf("fatal: No such remote 'origin'"),
			wantURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fake := newFakeRunner()
			fake.on("remote", func(_ []string, stdout, _ io.Writer) error {
				if tt.runErr != nil {
					return tt.runErr
				}
				_, _ = fmt.Fprint(stdout, tt.output)
				return nil
			})

			c := git.New("git", fake)
			url, err := c.RemoteURL(context.Background(), "/repo")
			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, url)
		})
	}
}

func TestURI2RelPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		uri     string
		want    string
		wantErr bool
	}{
		{
			name: "https with .git",
			uri:  "https://github.com/berquerant/git-iter-go.git",
			want: "github.com/berquerant/git-iter-go",
		},
		{
			name: "https without .git",
			uri:  "https://github.com/berquerant/git-iter-go",
			want: "github.com/berquerant/git-iter-go",
		},
		{
			name: "https with trailing slash",
			uri:  "https://github.com/berquerant/git-iter-go/",
			want: "github.com/berquerant/git-iter-go",
		},
		{
			name: "http URL",
			uri:  "http://example.com/org/repo.git",
			want: "example.com/org/repo",
		},
		{
			name: "scp-like ssh with .git",
			uri:  "git@github.com:berquerant/git-iter-go.git",
			want: "github.com/berquerant/git-iter-go",
		},
		{
			name: "scp-like ssh without .git",
			uri:  "git@github.com:berquerant/git-iter-go",
			want: "github.com/berquerant/git-iter-go",
		},
		{
			name: "scp-like ssh with nested groups",
			uri:  "git@gitlab.com:group/sub/repo.git",
			want: "gitlab.com/group/sub/repo",
		},
		{
			name: "git:// protocol",
			uri:  "git://github.com/berquerant/git-iter-go.git",
			want: "github.com/berquerant/git-iter-go",
		},
		{
			name: "ssh:// protocol with user and port",
			uri:  "ssh://git@github.com:22/berquerant/git-iter-go.git",
			want: "github.com/berquerant/git-iter-go",
		},
		{
			name: "repo name containing .git in middle",
			uri:  "https://github.com/berquerant/my.git.repo.git",
			want: "github.com/berquerant/my.git.repo",
		},
		{
			name: "scheme-less host/user/repo",
			uri:  "github.com/berquerant/git-iter-go",
			want: "github.com/berquerant/git-iter-go",
		},
		{
			name:    "empty string",
			uri:     "",
			wantErr: true,
		},
		{
			name:    "host only",
			uri:     "https://github.com",
			wantErr: true,
		},
		{
			name:    "single word",
			uri:     "notauri",
			wantErr: true,
		},
		{
			name:    "path traversal",
			uri:     "https://github.com/berquerant/..",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := git.URI2RelPath(tt.uri)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCommand_Clone(t *testing.T) {
	t.Parallel()

	fake := newFakeRunner()
	c := git.New("git", fake)
	err := c.Clone(context.Background(), "https://github.com/org/repo.git", "/repos/github.com/org/repo", io.Discard)
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"git", "clone", "https://github.com/org/repo.git", "/repos/github.com/org/repo"}}, fake.Calls())
}

func TestResolveCloneDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		reposRoot string
		uri       string
		want      string
		wantErr   bool
	}{
		{
			name:      "valid root and URI",
			reposRoot: "/repos",
			uri:       "https://github.com/org/repo.git",
			want:      "/repos/github.com/org/repo",
		},
		{
			name:      "empty reposRoot",
			reposRoot: "",
			uri:       "https://github.com/org/repo.git",
			wantErr:   true,
		},
		{
			name:      "invalid URI",
			reposRoot: "/repos",
			uri:       "invalid",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := git.ResolveCloneDir(tt.reposRoot, tt.uri)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}


