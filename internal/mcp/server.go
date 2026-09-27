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
	enableDo bool
}

// WithEnableDo enables the git_iter_do tool on the MCP server.
func WithEnableDo(enable bool) ServerOption {
	return func(c *serverConfig) {
		c.enableDo = enable
	}
}

// NewServer constructs and registers tools for an MCP server instance.
// By default, the git_iter_do tool is disabled unless WithEnableDo(true) is provided.
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
	if optConfig.enableDo {
		tool.RegisterDo(server, cfg)
	}
	tool.RegisterGrep(server, cfg)
	tool.RegisterLsFiles(server, cfg)
	tool.RegisterStatus(server, cfg)
	tool.RegisterDiff(server, cfg)

	return server
}
