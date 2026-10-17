package cmd

import (
	"fmt"

	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/spf13/cobra"
)

// NewGetCmd returns the "get" sub-command.
func NewGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [URI]",
		Short: "Resolve or clone a repository by URI, or list all repository paths",
		Args:  cobra.MaximumNArgs(1),
		Long: LongDoc{
			Header: `Without URI, print the absolute path of every discovered repository under --repos-root.
With URI, check whether the corresponding local repository directory exists under --repos-root (containing .git); if not, clone URI into that directory, and print its absolute path.`,
			OutputModes: RepoListOutputModeDoc(),
			Examples: []Example{
				{
					Desc:    "List all repositories under /repos",
					Command: "git-iter -r /repos get",
				},
				{
					Desc:    "Resolve or clone a repository and print its local path",
					Command: "git-iter -r /repos get https://github.com/berquerant/git-iter-go.git",
				},
			},
		}.Build(),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())
			getArgs := ParseGetArgs(args)

			finder, err := buildFinder(cfg, nil)
			if err != nil {
				return fmt.Errorf("get: %w", err)
			}

			return RunGet(
				cmd.Context(),
				cfg,
				OutModeFromContext(cmd.Context()),
				getArgs,
				finder,
				runner.ProcessRunner{},
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			)
		},
	}
}
