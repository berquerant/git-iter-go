package cmd

import (
	"fmt"

	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/spf13/cobra"
)

// NewLsFilesCmd returns the "ls-files" (alias: "lsf") sub-command.
func NewLsFilesCmd() *cobra.Command {
	var showAbsPath bool

	cmd := &cobra.Command{
		Use:     "ls-files [REPO_REGEX...] [FLAGS] -- [LS_FILES_FLAGS]",
		Aliases: []string{"lsf"},
		Short:   "Run git ls-files in each matching repository",
		Long: LongDoc{
			Header: `Run "git ls-files LS_FILES_FLAGS" in every repository whose path matches REPO_REGEX.

Each output line is prefixed with the repository path:

  <repo-path>/<file>

Path prefix format:
  By default the prefix is the repository path relative to --repos-root.
  Set --show-abs-path (or global --abs-path / -a / GIT_ITER_ABS_PATH=1) to use absolute paths.`,
			RepoFilter: true,
			DashSeparation: `Arguments before "--" are treated as REPO_REGEX filters (Go regular expressions,
joined with "|").
Arguments after  "--" are flags and arguments forwarded to "git ls-files".`,
			ExecutionNote: true,
			OutputModes: OutputModeDoc{
				Text: "Stream stdout/stderr directly (default).",
				JSON: "Emit one JSON object per file (JSON Lines):",
				JSONRows: []Item{
					{Term: JSONRowRepoAbsPath, Desc: JSONRowRepoAbsPathDesc},
					{Term: JSONRowRepoRelPath, Desc: JSONRowRepoRelPathDesc},
					{Term: "file_rel_path", Desc: "file path relative to repository"},
					{Term: "file_abs_path", Desc: "absolute path of the file"},
				},
				Markdown: DefaultMarkdownDesc,
			},
		}.Build(),
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())

			lsFilesArgs := ParseLsFilesArgsWithDash(args, cmd.ArgsLenAtDash())

			finder, err := buildFinder(cfg, lsFilesArgs.Patterns)
			if err != nil {
				return fmt.Errorf("ls-files: invalid pattern: %w", err)
			}

			return RunLsFiles(
				cmd.Context(),
				cfg,
				OutModeFromContext(cmd.Context()),
				lsFilesArgs,
				finder,
				runner.ProcessRunner{},
				showAbsPath,
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			)
		},
	}

	cmd.Flags().BoolVar(&showAbsPath, "show-abs-path", false,
		"show absolute paths in ls-files output")

	return cmd
}
