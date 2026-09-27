package cmd

import (
	"github.com/berquerant/git-iter-go/version"
	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version info",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return version.Write(cmd.OutOrStdout())
		},
	}
}
