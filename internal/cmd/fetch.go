package cmd

import (
	"fmt"

	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/spf13/cobra"
)

// NewFetchCmd returns the "fetch" sub-command.
func NewFetchCmd() *cobra.Command {
	var defaultBranch bool

	cmd := &cobra.Command{
		Use:   "fetch [REPO_REGEX...] [FLAGS] -- [FETCH_FLAGS]",
		Short: "Run git fetch in each matching repository",
		Long: LongDoc{
			Header:     "Run git fetch in every repository whose path matches all REPO_REGEX patterns.",
			RepoFilter: true,
			DashSeparation: `Arguments before "--" are treated as REPO_REGEX filters.
Arguments after  "--" are flags and arguments forwarded to "git fetch".`,
			ExecutionNote: true,
			OutputModes:   DefaultOutputModeDoc(),
		}.Build(),
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())

			fetchArgs := ParseFetchArgsWithDash(args, cmd.ArgsLenAtDash())

			finder, err := buildFinder(cfg, fetchArgs.Patterns)
			if err != nil {
				return fmt.Errorf("fetch: invalid pattern: %w", err)
			}

			fetchRunner := &runner.FetchRunner{
				GitCommand: cfg.GitCommand,
				Default:    defaultBranch,
				FetchFlags: fetchArgs.FetchFlags,
				Inner:      runner.ProcessRunner{},
			}

			return RunFetch(
				cmd.Context(),
				cfg,
				OutModeFromContext(cmd.Context()),
				fetchArgs,
				finder,
				fetchRunner,
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			)
		},
	}

	cmd.Flags().BoolVarP(&defaultBranch, "default", "s", false,
		"fetch only the repository's default branch")

	return cmd
}
