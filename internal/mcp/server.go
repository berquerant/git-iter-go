package mcp

import (
	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/mcp/tool"
	"github.com/berquerant/git-iter-go/version"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerOption configures an MCP server instance.
type ServerOption func(*serverConfig)

type serverConfig struct {
	enableDo       bool
	enableGet      bool
	enableWorktree bool
}

// WithEnableDo enables the git_iter_do tool on the MCP server.
func WithEnableDo(enable bool) ServerOption {
	return func(c *serverConfig) {
		c.enableDo = enable
	}
}

// WithEnableGet enables the git_iter_get tool on the MCP server.
func WithEnableGet(enable bool) ServerOption {
	return func(c *serverConfig) {
		c.enableGet = enable
	}
}

// WithEnableWorktree enables mutating worktree tools (add, remove, prune) on the MCP server.
func WithEnableWorktree(enable bool) ServerOption {
	return func(c *serverConfig) {
		c.enableWorktree = enable
	}
}

// NewServer constructs and registers tools for an MCP server instance.
// By default, mutating tools (git_iter_do, git_iter_get, and git_iter_worktree_add/remove/prune)
// are disabled unless enabled via options.
func NewServer(cfg *config.Config, opts ...ServerOption) *mcpsdk.Server {
	var optConfig serverConfig
	for _, opt := range opts {
		if opt != nil {
			opt(&optConfig)
		}
	}

	server := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "git-iter",
		Version: version.Version,
	}, nil)

	tool.RegisterList(server, cfg)
	tool.RegisterWorktreeList(server, cfg)
	if optConfig.enableDo {
		tool.RegisterDo(server, cfg)
	}
	if optConfig.enableGet {
		tool.RegisterGet(server, cfg)
	}
	if optConfig.enableWorktree {
		tool.RegisterWorktreeMutations(server, cfg)
	}
	tool.RegisterGrep(server, cfg)
	tool.RegisterLsFiles(server, cfg)
	tool.RegisterStatus(server, cfg)
	tool.RegisterDiff(server, cfg)

	return server
}

