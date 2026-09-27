package common

import (
	"github.com/berquerant/git-iter-go/internal/runner"
)

// Result holds the outcome of running a single command in one repository.
// It is serialised as one JSON object per line in json output mode.
type Result struct {
	RepoAbsPath string   `json:"repo_abs_path"` // absolute path of the repository
	RepoRelPath string   `json:"repo_rel_path"` // path relative to repos root
	Command     []string `json:"command"`       // command that was executed
	ExitCode    int      `json:"exit_code"`     // OS exit code (0 = success)
	Stdout      string   `json:"stdout"`        // captured standard output
	Stderr      string   `json:"stderr"`        // captured standard error
}

// RepoResult is the JSON representation for subcommands that list repositories
// without running a command (e.g. "list").
type RepoResult struct {
	RepoAbsPath string `json:"repo_abs_path"` // absolute path of the repository
	RepoRelPath string `json:"repo_rel_path"` // path relative to repos root
}

// FileResult is the JSON representation for subcommands that list files
// across repositories (e.g. "ls-files").
type FileResult struct {
	RepoAbsPath string `json:"repo_abs_path"` // absolute path of the repository
	RepoRelPath string `json:"repo_rel_path"` // path relative to repos root
	FileRelPath string `json:"file_rel_path"` // file path relative to repository
	FileAbsPath string `json:"file_abs_path"` // absolute path of the file
}

// Task describes a single unit of work for Executors.
// Each Task carries its own Runner so callers can inject per-repo runners
// (e.g. a GrepRunner configured for a specific repository).
type Task struct {
	// RepoAbsPath is the absolute path of the repository.
	RepoAbsPath string
	// ReposRoot is used to compute RepoRelPath; may be empty.
	ReposRoot string
	// Dir is the working directory passed to Runner (usually equals RepoAbsPath).
	Dir string
	// Command is the argv slice to execute.
	Command []string
	// Runner executes the command. If nil, runner.ProcessRunner{} is used.
	Runner runner.Runner
}
