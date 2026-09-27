package mcp

import "github.com/berquerant/git-iter-go/internal/mcp/tool"

// Type aliases for tool input and output DTOs moved to internal/mcp/tool.
type (
	ListToolsInput    = tool.ListToolsInput
	ListToolsOutput   = tool.ListToolsOutput
	DoToolInput       = tool.DoToolInput
	DoToolOutput      = tool.DoToolOutput
	GrepToolInput     = tool.GrepToolInput
	GrepToolOutput    = tool.GrepToolOutput
	LsFilesToolInput  = tool.LsFilesToolInput
	LsFilesToolOutput = tool.LsFilesToolOutput
	StatusToolInput   = tool.StatusToolInput
	StatusToolOutput  = tool.StatusToolOutput
	DiffToolInput     = tool.DiffToolInput
	DiffToolOutput    = tool.DiffToolOutput
)
