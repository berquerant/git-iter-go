package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewListCmd returns the "list" (alias: "ls") sub-command.
func NewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list [REPO_REGEX...]",
		Aliases: []string{"ls"},
		Short:   "List matching repository paths",
		Long: LongDoc{
			Header:     "Print the absolute path of every repository that matches REPO_REGEX, one per line.",
			RepoFilter: true,
			Sections: []Section{
				{
					Body: `This subcommand is also available as "ls".`,
				},
			},
			OutputModes: OutputModeDoc{
				Text: "Print one absolute path per line (default).",
				JSON: "Emit one JSON object per repository (JSON Lines):",
				JSONRows: []Item{
					{Term: JSONRowRepoAbsPath, Desc: JSONRowRepoAbsPathDesc},
					{Term: JSONRowRepoRelPath, Desc: JSONRowRepoRelPathDesc},
				},
				Markdown: "Emit a formatted Markdown document listing repositories (or md).",
			},
			Examples: []Example{
				{
					Desc:    "List all repositories under /repos",
					Command: "git-iter -r /repos list",
				},
				{
					Desc:    `List only repositories whose path contains "myorg"`,
					Command: "git-iter -r /repos list myorg",
				},
				{
					Desc:    `Pipe into "git-iter read" to run a command on the selected set`,
					Command: "git-iter -r /repos list myorg | git-iter read git status",
				},
				{
					Desc:    "List as JSON Lines",
					Command: "git-iter -r /repos -o json list",
				},
			},
		}.Build(),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())

			listArgs := ParseListArgs(args)

			finder, err := buildFinder(cfg, listArgs.Patterns)
			if err != nil {
				return fmt.Errorf("list: %w", err)
			}

			return RunList(cmd.Context(), cfg, OutModeFromContext(cmd.Context()), listArgs, finder, cmd.OutOrStdout())
		},
	}
}
