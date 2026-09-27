package cmd

import (
	"fmt"

	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/spf13/cobra"
)

// NewDoCmd returns the "do" sub-command.
func NewDoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "do [REPO_REGEX...] -- COMMAND [ARGS...]",
		Short: "Run a command in each matching repository",
		Long: LongDoc{
			Header:     "Run COMMAND in every repository whose path matches all REPO_REGEX patterns.",
			RepoFilter: true,
			DashSeparation: `Arguments before "--" are treated as REPO_REGEX filters.
Arguments after  "--" are the COMMAND to execute.`,
			ExecutionNote: true,
			OutputModes:   DefaultOutputModeDoc(),
			Examples: []Example{
				{
					Desc:    `Run "git status" in every repository under /repos`,
					Command: "git-iter -r /repos do -- git status",
				},
				{
					Desc:    `Run "git fetch" only in repositories whose path contains "myorg"`,
					Command: "git-iter -r /repos do myorg -- git fetch",
				},
				{
					Desc:    ExampleDescJSONLines,
					Command: "git-iter -r /repos -o json do -- git log -1 --oneline",
				},
			},
		}.Build(),
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())

			doArgs, err := ParseDoArgsWithDash(args, cmd.ArgsLenAtDash())
			if err != nil {
				return err
			}

			finder, err := buildFinder(cfg, doArgs.Patterns)
			if err != nil {
				return fmt.Errorf("do: invalid pattern: %w", err)
			}

			return RunDo(
				cmd.Context(), cfg, OutModeFromContext(cmd.Context()),
				doArgs, finder, runner.ProcessRunner{},
				cmd.OutOrStdout(), cmd.ErrOrStderr(),
			)
		},
	}
}
