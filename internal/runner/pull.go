package runner

import (
	"context"
	"fmt"
	"io"

	"github.com/berquerant/git-iter-go/internal/git"
)

// PullRunner handles git pull execution with optional default branch checkout
// and branch restoration.
type PullRunner struct {
	// GitCommand is the git binary name/path. Defaults to "git" if empty.
	GitCommand string
	// SwitchDefault checks out the repository's default branch before pulling.
	SwitchDefault bool
	// SwitchBack restores the initial branch after pulling if SwitchDefault changed branches.
	SwitchBack bool
	// PullFlags are additional arguments passed directly to "git pull".
	PullFlags []string
	// Inner executes git commands. Defaults to ProcessRunner{} if nil.
	Inner Runner
}

func (p *PullRunner) git() *git.Command {
	return git.New(p.GitCommand, p.Inner)
}

// prepareBranchSwitch switches to the default branch if needed.
// Returns the original branch and whether a switch occurred.
func (p *PullRunner) prepareBranchSwitch(ctx context.Context, dir string, stderr io.Writer) (string, bool, error) {
	if !p.SwitchDefault {
		return "", false, nil
	}

	g := p.git()
	origBranch, err := g.CurrentBranch(ctx, dir)
	if err != nil {
		return "", false, err
	}

	defaultBranch, err := g.ResolveDefaultBranch(ctx, dir)
	if err != nil {
		return "", false, err
	}

	if origBranch == defaultBranch {
		return origBranch, false, nil
	}

	if err := g.Checkout(ctx, dir, defaultBranch, stderr); err != nil {
		return "", false, fmt.Errorf("switch to default branch %q: %w", defaultBranch, err)
	}

	return origBranch, true, nil
}

// Run executes the pull sequence.
func (p *PullRunner) Run(ctx context.Context, dir string, _ []string, stdout, stderr io.Writer) error {
	origBranch, switchedBranch, err := p.prepareBranchSwitch(ctx, dir, stderr)
	if err != nil {
		return err
	}

	g := p.git()
	pullCmd := g.PullCmd(p.PullFlags...)
	pullErr := g.Run(ctx, dir, pullCmd[1:], stdout, stderr)

	if switchedBranch && p.SwitchBack && origBranch != "" {
		if switchBackErr := g.Checkout(ctx, dir, origBranch, stderr); switchBackErr != nil {
			if pullErr != nil {
				return fmt.Errorf("pull error: %w; switch back to %q failed: %v", pullErr, origBranch, switchBackErr)
			}
			return fmt.Errorf("switch back to %q: %w", origBranch, switchBackErr)
		}
	}

	return pullErr
}
