package cmd

import (
	"fmt"

	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/spf13/cobra"
)

// NewGrepCmd returns the "grep" sub-command.
func NewGrepCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "grep [REPO_REGEX...] -- GIT_GREP_ARGS...",
		Short: "Run git grep in each matching repository",
		Long: LongDoc{
			Header: `Run "git grep -H GIT_GREP_ARGS" in every repository whose path matches REPO_REGEX.

Each output line is prefixed with the repository path so results from different
repositories are distinguishable:

  <repo-path>/<file>:<line>:<match>

Path prefix format:
  By default the prefix is the repository path relative to --repos-root.
  Set --abs-path / -a (or GIT_ITER_ABS_PATH=1) to use absolute paths.`,
			RepoFilter: true,
			DashSeparation: `Arguments before "--" are treated as REPO_REGEX filters (Go regular expressions,
joined with "|").
Arguments after  "--" are passed verbatim to "git grep".`,
			Sections: []Section{
				{
					Body: `Repositories that produce no matches are silently skipped (git grep exits 1 for
no matches — treated as non-fatal).`,
				},
			},
			ExecutionNote: true,
			OutputModes: OutputModeDoc{
				Text: "Prefix-annotated grep lines streamed to stdout (default).",
				JSON: DefaultJSONDesc,
				JSONRows: []Item{
					{Term: JSONRowRepoAbsPath, Desc: JSONRowRepoAbsPathDesc},
					{Term: JSONRowRepoRelPath, Desc: JSONRowRepoRelPathDesc},
					{Term: JSONRowCommand, Desc: "argv passed to git grep"},
					{Term: JSONRowExitCode, Desc: "OS exit code (1 = no matches, non-fatal)"},
					{Term: JSONRowStdout, Desc: "captured prefixed grep output"},
					{Term: JSONRowStderr, Desc: JSONRowStderrDesc},
				},
				Markdown: DefaultMarkdownDesc,
			},
			Examples: []Example{
				{
					Desc:    `Search for "TODO" in all repositories`,
					Command: "git-iter -r /repos grep -- TODO",
				},
				{
					Desc:    ExampleDescJSONLines,
					Command: "git-iter -r /repos -o json grep -- FIXME",
				},
			},
		}.Build(),
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())

			grepArgs, err := ParseGrepArgsWithDash(args, cmd.ArgsLenAtDash())
			if err != nil {
				return err
			}

			finder, err := buildFinder(cfg, grepArgs.Patterns)
			if err != nil {
				return fmt.Errorf("grep: invalid pattern: %w", err)
			}

			return RunGrep(
				cmd.Context(), cfg, OutModeFromContext(cmd.Context()),
				grepArgs, finder, runner.ProcessRunner{},
				cmd.OutOrStdout(), cmd.ErrOrStderr(),
			)
		},
	}
}
