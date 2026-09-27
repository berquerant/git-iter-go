package markdown

import (
	"github.com/berquerant/git-iter-go/internal/markdown"
)

// Forwarding aliases to internal/markdown.
var (
	HeadingPrefix               = markdown.HeadingPrefix
	WriteMarkdownHeader         = markdown.WriteHeader
	WriteMarkdown               = markdown.WriteResult
	WriteMarkdownSummary        = markdown.WriteSummary
	WriteRepoMarkdown           = markdown.WriteRepo
	WriteListMarkdown           = markdown.WriteList
	ParseResultFiles            = markdown.ParseResultFiles
	WriteLsFilesMarkdown        = markdown.WriteLsFiles
	WriteLsFilesMarkdownSummary = markdown.WriteLsFilesSummary
)
