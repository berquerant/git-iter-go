package cmd

import (
	"fmt"

	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/spf13/cobra"
)

// NewStatusCmd returns the "status" (alias: "st") sub-command.
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status [REPO_REGEX...] [FLAGS] -- [STATUS_FLAGS]",
		Aliases: []string{"st"},
		Short:   "Run git status in each matching repository",
		Long: LongDoc{
			Header: `Run git status in every repository whose path matches REPO_REGEX.

By default, "git status --short" is executed.`,
			RepoFilter:     true,
			DashSeparation: `Any flags or arguments passed after "--" are forwarded verbatim to "git status".`,
			ExecutionNote:  true,
			OutputModes:    DefaultOutputModeDoc(),
		}.Build(),
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())

			statusArgs := ParseStatusArgsWithDash(args, cmd.ArgsLenAtDash())

			finder, err := buildFinder(cfg, statusArgs.Patterns)
			if err != nil {
				return fmt.Errorf("status: %w", err)
			}

			return RunStatus(
				cmd.Context(),
				cfg,
				OutModeFromContext(cmd.Context()),
				statusArgs,
				finder,
				runner.ProcessRunner{},
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			)
		},
	}

	return cmd
}
