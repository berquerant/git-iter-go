package cmd

import (
	"fmt"

	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/spf13/cobra"
)

// NewPullCmd returns the "pull" sub-command.
func NewPullCmd() *cobra.Command {
	var (
		switchDefault bool
		switchBack    bool
	)

	cmd := &cobra.Command{
		Use:   "pull [REPO_REGEX...] [FLAGS] -- [PULL_FLAGS]",
		Short: "Run git pull in each matching repository",
		Long: LongDoc{
			Header:     "Run git pull in every repository whose path matches all REPO_REGEX patterns.",
			RepoFilter: true,
			DashSeparation: `Arguments before "--" are treated as REPO_REGEX filters.
Arguments after  "--" are flags and arguments forwarded to "git pull".`,
			ExecutionNote: true,
			OutputModes:   DefaultOutputModeDoc(),
		}.Build(),
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())

			pullArgs := ParsePullArgsWithDash(args, cmd.ArgsLenAtDash())

			finder, err := buildFinder(cfg, pullArgs.Patterns)
			if err != nil {
				return fmt.Errorf("pull: invalid pattern: %w", err)
			}

			pullRunner := &runner.PullRunner{
				GitCommand:    cfg.GitCommand,
				SwitchDefault: switchDefault,
				SwitchBack:    switchBack,
				PullFlags:     pullArgs.PullFlags,
				Inner:         runner.ProcessRunner{},
			}

			return RunPull(
				cmd.Context(),
				cfg,
				OutModeFromContext(cmd.Context()),
				pullArgs,
				finder,
				pullRunner,
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			)
		},
	}

	cmd.Flags().BoolVarP(&switchDefault, "switch-default", "s", false,
		"checkout default branch before pulling")
	cmd.Flags().BoolVarP(&switchBack, "switch-back", "b", false,
		"switch back to initial branch after pulling (requires --switch-default)")

	return cmd
}
