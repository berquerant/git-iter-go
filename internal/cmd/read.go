package cmd

import (
	"github.com/berquerant/git-iter-go/internal/repo"
	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/spf13/cobra"
)

// NewReadCmd returns the "read" sub-command.
func NewReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "read COMMAND [ARGS...]",
		Short: "Run a command in repositories listed on stdin",
		Long: LongDoc{
			Header: `Run COMMAND in each repository path read from stdin (one absolute path per line).

This subcommand is useful for composing git-iter with other tools:
  - pipe the output of "git-iter list" to narrow down repositories
  - pipe any newline-delimited list of paths from an external source

Blank lines in stdin are ignored.
The working directory is set to each repository path for every invocation.`,
			Sections: []Section{
				{
					Body: "Parallelism is controlled by --max-procs / GIT_ITER_MAX_PROCS (default: 1).",
				},
				{
					Body: `If COMMAND or its arguments include flags (such as "bash -c"), precede COMMAND
with "--" so that flags are not parsed as git-iter global options:
  git-iter read -- bash -c "git status"`,
				},
			},
			ExecutionNote: false, // Execution note without repos-root discovery
			OutputModes: OutputModeDoc{
				Text: DefaultTextDesc,
				JSON: DefaultJSONDesc,
				JSONRows: []Item{
					{Term: JSONRowRepoAbsPath, Desc: JSONRowRepoAbsPathDesc},
					{Term: JSONRowRepoRelPath, Desc: "path relative to --repos-root (or absolute path if not specified)"},
					{Term: JSONRowCommand, Desc: "argv slice that was executed"},
					{Term: JSONRowExitCode, Desc: "OS exit code of the process"},
					{Term: JSONRowStdout, Desc: "captured standard output"},
					{Term: JSONRowStderr, Desc: JSONRowStderrDesc},
				},
				Markdown: DefaultMarkdownDesc,
			},
			Examples: []Example{
				{
					Desc:    `Run "git status" in repos piped from "git-iter list"`,
					Command: "git-iter -r /repos list myorg | git-iter read git status",
				},
				{
					Desc:    `Execute a shell command using bash -c (use -- to prevent -c being parsed as git-iter flag)`,
					Command: `git-iter -r /repos list myorg | git-iter read -- bash -c "git status -s"`,
				},
				{
					Desc:    ExampleDescJSONLines,
					Command: "git-iter -r /repos list myorg | git-iter -o json read git log -1 --oneline",
				},
			},
		}.Build(),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())

			readArgs := ParseReadArgs(args)
			var finder repo.Finder = &repo.StdinFinder{Reader: cmd.InOrStdin()}
			finder = repo.NewSortedFinder(finder, cfg.Sort, cfg.SortReverse)
			finder = repo.NewLimitedFinder(finder, cfg.Limit)

			return RunRead(
				cmd.Context(),
				cfg,
				OutModeFromContext(cmd.Context()),
				readArgs,
				finder,
				runner.ProcessRunner{},
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			)
		},
	}
}
