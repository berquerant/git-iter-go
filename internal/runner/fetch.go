package runner

import (
	"context"
	"io"

	"github.com/berquerant/git-iter-go/internal/git"
)

// FetchRunner handles git fetch execution across repositories.
type FetchRunner struct {
	// GitCommand is the git binary name/path. Defaults to "git" if empty.
	GitCommand string
	// Default fetches only the repository's default branch.
	Default bool
	// FetchFlags are additional arguments passed directly to "git fetch".
	FetchFlags []string
	// Inner executes git commands. Defaults to ProcessRunner{} if nil.
	Inner Runner
}

func (f *FetchRunner) git() *git.Command {
	return git.New(f.GitCommand, f.Inner)
}

// buildFetchFlags constructs the arguments slice passed to "git fetch".
func (f *FetchRunner) buildFetchFlags(ctx context.Context, dir string) ([]string, error) {
	flags := append([]string{}, f.FetchFlags...)

	if f.Default {
		g := f.git()
		defaultBranch, err := g.ResolveDefaultBranch(ctx, dir)
		if err != nil {
			return nil, err
		}
		// fetch origin <default-branch>
		flags = append(flags, "origin", defaultBranch)
	}

	return flags, nil
}

// Run executes the fetch command.
func (f *FetchRunner) Run(ctx context.Context, dir string, _ []string, stdout, stderr io.Writer) error {
	flags, err := f.buildFetchFlags(ctx, dir)
	if err != nil {
		return err
	}

	g := f.git()
	fetchCmd := g.FetchCmd(flags...)
	return g.Run(ctx, dir, fetchCmd[1:], stdout, stderr)
}
