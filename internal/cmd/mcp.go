package cmd

import (
	"github.com/berquerant/git-iter-go/internal/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// NewMcpCmd returns the "mcp" sub-command.
func NewMcpCmd() *cobra.Command {
	var (
		enableDo       bool
		enableGet      bool
		enableWorktree bool
	)
	cmd := &cobra.Command{
		Use:   "mcp [FLAGS]",
		Short: "Run as an MCP (Model Context Protocol) server over stdio",
		Long: `Run git-iter as a Model Context Protocol (MCP) server communicating over stdin/stdout.

Exposes tools for LLMs/MCP clients to discover repositories and worktrees, inspect status, diff, and grep code.
Mutating tools ("git_iter_do", "git_iter_get", and "git_iter_worktree_add/remove/prune")
are disabled by default for security, and can be enabled with --enable-do, --enable-get, and --enable-worktree.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := ConfigFromContext(cmd.Context())
			server := mcp.NewServer(
				cfg,
				mcp.WithEnableDo(enableDo),
				mcp.WithEnableGet(enableGet),
				mcp.WithEnableWorktree(enableWorktree),
			)
			return server.Run(cmd.Context(), &mcpsdk.StdioTransport{})
		},
	}

	cmd.Flags().BoolVar(&enableDo, "enable-do", false, "enable git_iter_do tool (arbitrary command execution)")
	cmd.Flags().BoolVar(&enableGet, "enable-get", false, "enable git_iter_get tool (repository cloning and discovery)")
	cmd.Flags().BoolVar(&enableWorktree, "enable-worktree", false, "enable mutating worktree tools (add, remove, prune)")

	return cmd
}

