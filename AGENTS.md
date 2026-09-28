# AGENTS.md — git-iter-go

This document provides guidance for AI agents working on this codebase.

## Overview

`git-iter-go` is a high-performance Go CLI tool designed to run commands, search, list, or pipe operations across multiple local git repositories concurrently, with capabilities such as parallel execution controls, process timeouts, structured JSON Lines (`json`) and Markdown (`markdown`/`md`) output formats.

---

## Repository Layout

.
├── .github/workflows/                 # GitHub Actions (lint, test, release)
├── .goreleaser.yml                    # GoReleaser configuration for cross-platform releases
├── LICENSE                            # License file
├── NOTICE                             # Third-party notices generated via go-licenses
├── notice-template.md                 # Template used to generate NOTICE
├── mise.toml                          # Tool definitions (Go, golangci-lint, go-licenses) and task runner
├── cmd/
│   └── git-iter/
│       └── main.go                    # Entrypoint (signal handling, root command invocation)
├── internal/
│   ├── cmd/                           # Cobra CLI layer & command coordinators
│   │   ├── root.go                    # Root command, global flags, slog setup, config injection
│   │   ├── do.go                      # "do" subcommand: run commands across matched repos
│   │   ├── grep.go                    # "grep" subcommand: git grep with repository prefix
│   │   ├── list.go                    # "list" / "ls" subcommand: discover and list repo paths
│   │   ├── read.go                    # "read" subcommand: execute commands for repos fed via stdin
│   │   ├── pull.go                    # "pull" subcommand: git pull across matched repos with branch switching
│   │   ├── fetch.go                   # "fetch" subcommand: git fetch across matched repos with default branch filter
│   │   ├── ls_files.go                # "ls-files" / "lsf" subcommand: git ls-files with repository prefix
│   │   ├── status.go                  # "status" / "st" subcommand: git status across matched repos
│   │   ├── diff.go                    # "diff" subcommand: git diff across matched repos
│   │   ├── mcp.go                     # "mcp" subcommand: Model Context Protocol (MCP) server CLI wiring
│   │   ├── version.go                 # "version" subcommand: display build version, commit, date
│   │   ├── parse.go                   # Finder builder & parse package type aliases
│   │   ├── runner.go                  # Aliases to internal/cmd/runner functions
│   │   ├── parse/                     # CLI argument parsing package
│   │   │   ├── dash.go                # Dash separation logic (SplitWithDash)
│   │   │   ├── do.go                  # Do subcommand argument parsing (DoArgs, ParseDoArgs)
│   │   │   ├── grep.go                # Grep subcommand argument parsing (GrepArgs, ParseGrepArgs)
│   │   │   ├── repo.go                # Repo listing & reading argument parsing (ListArgs, ReadArgs)
│   │   │   ├── git_sync.go            # Remote sync argument parsing (PullArgs, FetchArgs)
│   │   │   └── git_inspect.go         # Git status, diff, ls-files argument parsing
│   │   └── runner/                    # Subcommand runner orchestration package
│   │       ├── common.go              # Shared task construction and streaming execution logic
│   │       ├── do.go                  # RunDo, RunRead command runners
│   │       ├── repo.go                # RunList repository runner
│   │       ├── grep.go                # RunGrep runner with path prefixing
│   │       ├── git_sync.go            # RunPull, RunFetch runners
│   │       └── git_inspect.go         # RunLsFiles, RunStatus, RunDiff runners
│   ├── mcp/                           # Model Context Protocol (MCP) server implementation
│   │   ├── server.go                  # MCP server instance creation, options, and tool wiring
│   │   ├── types.go                   # Aliases to internal/mcp/tool DTOs
│   │   └── tool/                      # MCP tool implementations and DTOs
│   │       ├── common.go              # Shared finder, timeout, and execution helpers
│   │       ├── git_common.go          # Shared git command tool registration helper
│   │       ├── list.go                # git_iter_list tool
│   │       ├── do.go                  # git_iter_do tool
│   │       ├── grep.go                # git_iter_grep tool
│   │       ├── ls_files.go            # git_iter_ls_files tool
│   │       ├── status.go              # git_iter_status tool
│   │       └── diff.go                # git_iter_diff tool
│   ├── config/                        # Configuration resolution via structconfig
│   │   └── config.go                  # Config struct, env/flag mapping (GIT_ITER_*), log levels
│   ├── executor/                      # Unstructured/streaming concurrent task execution
│   │   └── executor.go                # Executor: concurrency control (semaphore), timeout, fail-fast
│   ├── git/                           # Centralized git CLI command builder and operations
│   │   └── git.go                     # git.Command: branch operations, default branch resolution, grep/pull builder
│   ├── markdown/                      # Pure Markdown document rendering and formatting
│   │   ├── details.go                 # Details section formatting (Details, FormatDetails, FormatCodeDetails)
│   │   ├── heading.go                 # Heading prefix and header generation (HeadingPrefix, WriteHeader)
│   │   ├── section.go                 # Repository result, summary, and list formatting (WriteResult, WriteSummary, WriteList)
│   │   └── ls_files.go                # ls-files specific formatting (WriteLsFiles, WriteLsFilesSummary, ParseResultFiles)
│   ├── output/                        # Structured output handling (JSONL & Markdown)
│   │   ├── output.go                  # Aliases to subpackages (common, jsonl, markdown)
│   │   ├── common/                    # Common DTOs and execution primitives (Result, Task, RelPath, ExecuteTasksCore)
│   │   │   ├── types.go               # Result, RepoResult, FileResult, Task DTOs
│   │   │   ├── path.go                # Path and exit code helpers (RelPath, ExitCodeFrom)
│   │   │   └── exec.go                # Concurrent task execution primitives (ExecuteTask, ExecuteTasksCore)
│   │   ├── jsonl/                     # JSON Lines output serialization and executors
│   │   │   ├── write.go               # WriteJSONL, WriteRepoJSONL, WriteFileJSONL
│   │   │   └── executor.go            # JSONLExecutor, FileJSONLExecutor
│   │   └── markdown/                  # Markdown executors orchestrating internal/markdown formatting
│   │       ├── heading.go             # Aliases to internal/markdown functions
│   │       └── executor.go            # MarkdownExecutor, LsFilesMarkdownExecutor
│   ├── repo/                          # Repository discovery and path filtering
│   │   ├── repo.go                    # Finder interface, StdinFinder, scanning helpers
│   │   ├── fs.go                      # FilesystemFinder, CommandFinder, NewFinder
│   │   ├── filter.go                  # FilteredFinder, StatusFilterFinder, RemoteURLFilterFinder, DefaultBranchFilterFinder, BranchFilterFinder
│   │   └── sort_limit.go              # SortedFinder, LimitedFinder
│   └── runner/                        # Process execution abstraction
│       ├── runner.go                  # Runner interface, ProcessRunner, PathPrefixRunner (GrepRunner)
│       ├── pull.go                    # PullRunner: git pull with default branch checkout & restoration
│       ├── fetch.go                   # FetchRunner: git fetch with optional default branch restriction
│       └── error.go                   # TaskError: structured execution error with repo/cmd/exit code
├── testutil/
│   └── testutil.go                    # FakeRunner and FakeFinder test doubles
└── version/
    └── version.go                     # Version, Commit, and Date build metadata

---

## Architecture

The project follows a clean, modular architecture separating CLI definition, configuration, discovery, execution, output rendering, and build metadata.

### Architecture Diagram

```mermaid
flowchart TD
    CLI[cmd/git-iter/main.go] --> RootCmd[internal/cmd: NewRootCmd]
    RootCmd --> Conf[internal/config: Config via structconfig]
    
    subgraph CLI Commands [Subcommands]
        DoCmd[do]
        GrepCmd[grep]
        ListCmd[list / ls]
        LsFilesCmd["ls-files / lsf"]
        ReadCmd[read]
        PullCmd[pull]
        FetchCmd[fetch]
        StatusCmd["status / st"]
        DiffCmd[diff]
        McpCmd["mcp (MCP Stdio Server)"]
        VersionCmd[version]
    end
    RootCmd --> CLI Commands
    VersionCmd --> VersionPkg[version package]
    
    subgraph Discovery [Repository Discovery: internal/repo]
        Finder[Finder Interface]
        FSFinder[FilesystemFinder: WalkDir .git]
        CmdFinder[CommandFinder: list-repos shell cmd]
        StdinFinder[StdinFinder: pipe from stdin]
        FilterFinder[FilteredFinder: regex filter]
        
        FSFinder --> Finder
        CmdFinder --> Finder
        StdinFinder --> Finder
        Finder --> FilterFinder
    end
    CLI Commands --> Discovery
    
    subgraph Execution & Output [Runners & Executors]
        Runner[Runner Interface: internal/runner]
        PRunner[ProcessRunner: os/exec]
        PPRunner["PathPrefixRunner (GrepRunner): path prefixing"]
        PullRunner[PullRunner: branch switch & pull]
        FetchRunner[FetchRunner: default branch fetch]
        PRunner --> Runner
        PPRunner --> Runner
        PullRunner --> Runner
        FetchRunner --> Runner

        ExecDispatch{Output Mode?}
        TextExec["Executor (internal/executor)<br/>Stream stdout/stderr directly"]
        JSONLExec["JSONLExecutor (internal/output)<br/>Structured JSON Lines"]
        MDExec["MarkdownExecutor (internal/output)<br/>Collapsible sections + Summary"]
        
        ExecDispatch -->|text| TextExec
        ExecDispatch -->|json| JSONLExec
        ExecDispatch -->|markdown / md| MDExec
    end
    
    CLI Commands --> ExecDispatch
    TextExec --> Runner
    JSONLExec --> Runner
    MDExec --> Runner
```

---

## Core Components

### 1. Configuration (`internal/config`)
- Powered by `github.com/berquerant/structconfig`.
- Resolves configuration in order of priority: **Default values < Environment variables (`GIT_ITER_*`) < CLI flags**.
- Handles flags:
  - `-r`, `--repos-root`: Root directory scanned for `.git` entries.
  - `--list-repos`: Shell command generating repo paths.
  - `-p`, `--max-procs`: Parallelism limit (default: `1`).
  - `--timeout`: Per-process execution timeout (e.g. `5s`, `1m`).
  - `--fail-fast`: Abort immediately if any command fails.
  - `--git-command`: Path to git executable (default: `git`).
  - `-a`, `--abs-path`: Grep and path-annotated outputs use absolute repo paths instead of relative.
  - `--markdown-header-level`: Heading level (1-6) for markdown repo sections (default: `3`).
  - `--sort`: Sort discovered repository paths in ascending order.
  - `--sort-reverse`: Sort discovered repository paths in descending (reverse) order.
  - `-n`, `--limit`: Limit the number of processed repositories to the first n repos (applied after sort).
  - `-d`, `--dirty-only`: Filter repositories to only those with uncommitted/untracked changes.
  - `-c`, `--clean-only`: Filter repositories to only those without uncommitted/untracked changes.
  - `--default-branch-only`: Filter repositories to only those whose current branch is the default branch.
  - `--not-default-branch-only`: Filter repositories to only those whose current branch is not the default branch.
  - `--branch`: Filter repositories to only those whose current branch matches the specified regex pattern.
  - `--remote-url`: Filter repositories to only those whose origin remote URL matches the specified regex pattern.
  - `-l`, `--log-level`: Logging level (`debug`, `info`, `warn`, `error`). Logs are always directed to stderr via `slog`.

### 2. Repository Discovery (`internal/repo`)
- `Finder`: Interface `Find(ctx context.Context) ([]string, error)` producing absolute repository paths.
- `FilesystemFinder`: Recursively searches `Root` for directories containing a `.git` entry (prunes descending into nested git repositories).
- `CommandFinder`: Executes a custom shell command to fetch repository paths (e.g., integrating with `ghq` or locate).
- `StdinFinder`: Reads newline-delimited repository paths from stdin (used by `read`).
- `FilteredFinder`: Filters repository paths using Go regex patterns combined with `|`.
- `SortedFinder`: Sorts repository paths ascending or descending.
- `LimitedFinder`: Limits repository paths to at most `Limit` items from the beginning.
- `StatusFilterFinder`: Filters repository paths based on dirty/clean state via git status.
- `DefaultBranchFilterFinder`: Filters repository paths based on whether current branch is or is not the default branch.
- `BranchFilterFinder`: Filters repository paths based on current branch matching a regex pattern.
- `RemoteURLFilterFinder`: Filters repository paths based on origin remote URL matching a regex pattern.

### 3. Execution Pipeline (`internal/runner`, `internal/executor`, `internal/output`)
- `Runner`: Interface `Run(ctx, dir, command, stdout, stderr) error`.
  - `ProcessRunner`: Executes subprocesses in the target repo's directory using `os/exec.CommandContext`.
  - `PathPrefixRunner` (`GrepRunner`): Wraps another `Runner` to execute commands (e.g. `git grep`, `git ls-files`) and prefix output lines with the repository relative/absolute path.
  - `PullRunner`: Executes `git pull` with optional default branch detection (`origin/HEAD` / `init.defaultBranch` / `main` / `master`), checkout, and branch restoration.
  - `FetchRunner`: Executes `git fetch` with optional default branch restriction (`git fetch origin <default-branch>`).
- Concurrency & Process Control:
  - Controlled by `golang.org/x/sync/semaphore.Weighted`.
  - Per-task timeout wrapping via `context.WithTimeout(ctx, Timeout)`.
  - Correlation logging: Each task execution logs start and completion with a unique UUID.
  - `runner.TaskError`: Structured error capturing `Dir`, `Command`, `ExitCode`, and the underlying `error`.
- Output Modes:
  - **`text`** (`internal/executor.Executor`): Direct streaming of stdout/stderr from concurrent processes.
  - **`json`** (`internal/output.JSONLExecutor` / `FileJSONLExecutor`): Outputs one JSONL object (`output.Result`, `output.RepoResult`, or `output.FileResult` for `ls-files`) per repository or file. Exit codes are captured as data and do not prematurely abort execution unless `FailFast` is enabled.
  - **`markdown`** / **`md`** (`internal/output.MarkdownExecutor`): Generates a Markdown document with heading levels specified by `--markdown-header-level`, collapsible `<details>` blocks for non-empty stdout/stderr, and an aggregate summary at the end reporting total count, exit code distribution, and failed repos.

### 4. MCP Server (`internal/mcp`)
- Runs as a Model Context Protocol (MCP) server over `mcp.StdioTransport` (stdin/stdout).
- Powered by `github.com/modelcontextprotocol/go-sdk`.
- Exposes tools to LLM/MCP clients:
  - `git_iter_list`: Discover repositories matching patterns under a root directory.
  - `git_iter_do`: Run arbitrary commands across matching repositories, returning exit codes, stdout, and stderr (disabled by default for security, enabled via `--enable-do`).
  - `git_iter_grep`: Run `git grep` across matching repositories with prefixed output paths.
  - `git_iter_ls_files`: Run `git ls-files` across matching repositories with prefixed output paths.
  - `git_iter_status`: Run `git status` across matching repositories, returning exit codes, stdout, and stderr.
  - `git_iter_diff`: Run `git diff` across matching repositories, returning exit codes, stdout, and stderr.
- Preserves output separation by keeping JSON-RPC transport clean on stdout/stdin while diagnostic `slog` output is routed to stderr.

### 5. Version & Build Metadata (`version`, `internal/cmd/version.go`)
- `version.Version`, `version.Commit`, `version.Date` are populated at build time via ldflags (e.g., `-X github.com/berquerant/git-iter-go/version.Version=...`).
- The `version` subcommand prints build metadata to stdout.
- Cross-platform release binaries and GitHub releases are managed with GoReleaser (`.goreleaser.yml`, triggered on tag push `v*`).

---

## Development & Testing Workflow

This project uses standard Go tooling managed via `mise`:

```bash
# Run lint, test, and build (default)
mise run default

# Run all linters (go fix, go vet, golangci-lint, licenses-check)
mise run lint

# Run golangci-lint directly
mise run golangci-lint

# Run go vet
mise run vet

# Check and update open-source licenses / NOTICE
mise run licenses        # runs licenses-check
mise run licenses-gen    # generate NOTICE report
mise run licenses-diff   # verify NOTICE is up to date

# Run unit tests with race detection and coverage
mise run test

# Build binary (outputs to bin/git-iter)
mise run build
```

Underlying test command:
```bash
go test -cover -race ./...
```

---

## Key Conventions & Guidelines

1. **Table-Driven Tests**:
   - All tests in `*_test.go` should follow Go's table-driven test idiom (`tests := []struct{...}{...}`) where possible.
   - Use `t.Parallel()` for tests that do not mutate global process state (note: tests mutating `slog.SetDefault` must not run in parallel).
   - Use Go's built-in `new(0)`, `new(1)`, etc., when pointers to scalar literals are needed in test tables.
2. **Output Separation**:
   - `stdout` is strictly reserved for command output, JSONL lines, and Markdown reports.
   - Informational and diagnostic logs (`slog`) must always go to `stderr`.
3. **Subcommand Separation with `--`**:
   - In `do` and `grep`, repository regex patterns precede `--` and command arguments follow `--`.
   - In `read`, use `--` before commands containing flags (e.g. `read -- bash -c "..."`) to prevent flags like `-c` from being intercepted by `git-iter`'s global options (`--clean-only`).
   - Arguments are parsed via `cmd.ParseDoArgsWithDash` and `cmd.ParseGrepArgsWithDash` taking `cmd.ArgsLenAtDash()` into account.
4. **Error Handling**:
   - Wrap task execution errors in `runner.TaskError` so context (dir, command, exit code) is preserved.
   - In `json` mode, process exit errors are treated as data captured in `Result.ExitCode` rather than fatal execution failures, preserving full visibility over all repositories.
5. **Investigation Time Limit (2 Minutes Max)**:
   - As a general rule, an agent must not spend more than 2 minutes investigating or gathering information to determine the next action.
   - If an investigation is expected to take, or has taken, 2 minutes without reaching a conclusion, the agent must pause at the 2-minute mark and request user approval or guidance before proceeding further.

---

## Post-Implementation Refactoring Guidelines

After completing feature implementations or bug fixes, perform refactoring following these rules to maintain code health and reliability:

1. **Separation of Production and Test Edits**:
   - Never modify production code and test code at the same time in a single step.
   - Keep behavioral assertions intact while refactoring production code, and vice versa.
2. **Verification Loop for Production Changes**:
   - Every production code change must leave the codebase with all tests and linters passing (`go test -cover -race ./...` and `golangci-lint run`).
3. **Careful Test Maintenance**:
   - When modifying tests, exercise extreme caution to prevent regressions, false positives, or accidental loss of test coverage.
4. **DRY (Don't Repeat Yourself)**:
   - Consolidate duplicated logic, boilerplate, and command-building patterns into reusable helpers or shared packages wherever possible.
5. **Function and Method Sizing**:
   - Break down overly long functions or methods into cohesive, single-responsibility units.
6. **Complexity and Branch Management**:
   - Simplify and decompose functions with high cyclomatic complexity or excessive branching/nesting into focused subroutines or early-return guards.
7. **File Sizing & Modularity**:
   - Split oversized files into smaller, logically organized source files within the package.
8. **Go Idioms & Best Practices**:
   - Follow standard Go conventions and guidelines:
     - [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
     - [Effective Go](https://go.dev/doc/effective_go)
