package cmd

import (
	"fmt"

	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/spf13/cobra"
)

// NewDiffCmd returns the "diff" sub-command.
func NewDiffCmd() *cobra.Command {
	var interactive bool
	cmd := &cobra.Command{
		Use:   "diff [REPO_REGEX...] [FLAGS] -- [DIFF_FLAGS]",
		Short: "Run git diff in each matching repository",
		Long: LongDoc{
			Header:     "Run git diff in every repository whose path matches REPO_REGEX.",
			RepoFilter: true,
			DashSeparation: `Any flags or arguments passed after "--" are forwarded verbatim to "git diff"
(e.g. --stat, --cached, main..HEAD).`,
			ExecutionNote: true,
			OutputModes:   DefaultOutputModeDoc(),
		}.Build(),
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())

			diffArgs := ParseDiffArgsWithDash(args, cmd.ArgsLenAtDash())
			diffArgs.Interactive = interactive

			finder, err := buildFinder(cfg, diffArgs.Patterns)
			if err != nil {
				return fmt.Errorf("diff: %w", err)
			}

			return RunDiff(
				cmd.Context(),
				cfg,
				OutModeFromContext(cmd.Context()),
				diffArgs,
				finder,
				runner.ProcessRunner{},
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			)
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "enable pager and interactive git diff")

	return cmd
}
