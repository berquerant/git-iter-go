// Package output provides types and helpers for structured (JSONL) and Markdown output.
package output

import (
	"github.com/berquerant/git-iter-go/internal/output/common"
	"github.com/berquerant/git-iter-go/internal/output/jsonl"
	"github.com/berquerant/git-iter-go/internal/output/markdown"
)

// DTO and task type aliases from common subpackage.
type (
	Result     = common.Result
	RepoResult = common.RepoResult
	FileResult = common.FileResult
	Task       = common.Task
)

// Helpers from common subpackage.
var (
	ExitCodeFrom     = common.ExitCodeFrom
	RelPath          = common.RelPath
	ExecuteTask      = common.ExecuteTask
	ExecuteTasksCore = common.ExecuteTasksCore
)

// JSONL types and functions from jsonl subpackage.
type (
	JSONLExecutor     = jsonl.JSONLExecutor
	FileJSONLExecutor = jsonl.FileJSONLExecutor
)

var (
	WriteJSONL     = jsonl.WriteJSONL
	WriteRepoJSONL = jsonl.WriteRepoJSONL
	WriteFileJSONL = jsonl.WriteFileJSONL
)

// Markdown types and functions from markdown subpackage.
type (
	MarkdownExecutor        = markdown.MarkdownExecutor
	LsFilesMarkdownExecutor = markdown.LsFilesMarkdownExecutor
)

var (
	HeadingPrefix               = markdown.HeadingPrefix
	WriteMarkdownHeader         = markdown.WriteMarkdownHeader
	WriteMarkdown               = markdown.WriteMarkdown
	WriteMarkdownSummary        = markdown.WriteMarkdownSummary
	WriteRepoMarkdown           = markdown.WriteRepoMarkdown
	WriteListMarkdown           = markdown.WriteListMarkdown
	ParseResultFiles            = markdown.ParseResultFiles
	WriteLsFilesMarkdown        = markdown.WriteLsFilesMarkdown
	WriteLsFilesMarkdownSummary = markdown.WriteLsFilesMarkdownSummary
)
