# git-iter

[![CI](https://github.com/berquerant/git-iter-go/actions/workflows/test.yml/badge.svg)](https://github.com/berquerant/git-iter-go/actions/workflows/test.yml)
[![Lint](https://github.com/berquerant/git-iter-go/actions/workflows/lint.yml/badge.svg)](https://github.com/berquerant/git-iter-go/actions/workflows/lint.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`git-iter` is a high-performance CLI tool to discover, search, and run commands concurrently across multiple local git repositories.

It supports parallel execution controls, per-process timeouts, fail-fast behavior, and structured output formats (JSON Lines and Markdown reports), alongside a built-in Model Context Protocol (MCP) server for LLMs and agents.

---

## Features

- **Concurrent Execution**: Run commands across multiple repositories simultaneously (`-p, --max-procs`).
- **Flexible Discovery**: Recursively discover git repositories (`-r, --repos-root`) or provide arbitrary discovery shell commands (`--list-repos`).
- **Filtering & Limiting**:
  - Filter repositories using regex patterns.
  - Filter by repository state (`-d, --dirty-only`, `-c, --clean-only`).
  - Filter by origin remote URL regex pattern (`--remote-url`).
  - Sort (`--sort`, `--sort-reverse`) and limit (`-n, --limit`).
- **Multiple Output Formats**:
  - `text` (default): Real-time streaming output to stdout/stderr.
  - `json`: Structured JSON Lines (`output.Result` / `output.RepoResult` / `output.FileResult`).
  - `markdown` / `md`: Markdown reports with collapsible details and summary tables.
- **Dedicated Subcommands**:
  - `do`: Run arbitrary commands in matching repositories.
  - `grep`: Run `git grep` with repository path prefixing.
  - `diff`: Run `git diff` non-interactively (`--no-pager`) by default, with opt-in pager (`-i, --interactive`).
  - `status` (`st`): Inspect status across repositories.
  - `ls-files` (`lsf`): List tracked/untracked files with repository path prefixing.
  - `pull`: Pull with optional default branch checkout and switch-back (`-s, --switch-default`, `-b, --switch-back`).
  - `fetch`: Fetch with optional default branch restriction (`-s, --default`).
  - `list` (`ls`): List matching repository paths.
  - `read`: Execute commands for repository paths fed via stdin.
  - `mcp`: Model Context Protocol server over stdio for LLM integration.

---

## Installation

### Via `go install`

```bash
go install github.com/berquerant/git-iter-go/cmd/git-iter@latest
```

### Pre-built Binaries

Download pre-compiled binaries for Linux, macOS, or Windows from [GitHub Releases](https://github.com/berquerant/git-iter-go/releases).

---

## Usage Overview

```text
git-iter [command] [flags]
```

### Global Flags & Configuration

Settings can be specified via CLI flags or environment variables (`GIT_ITER_*`):

| Flag | Environment Variable | Default | Description |
|---|---|---|---|
| `-r, --repos-root` | `GIT_ITER_REPOS_ROOT` | `""` | Root directory scanned recursively for repositories |
| `--list-repos` | `GIT_ITER_LIST_REPOS` | `""` | Shell command outputting repository paths (one per line) |
| `-p, --max-procs` | `GIT_ITER_MAX_PROCS` | `1` | Maximum parallel worker processes |
| `-o, --out` | - | `text` | Output mode (`text`, `json`, `markdown`/`md`) |
| `--timeout` | `GIT_ITER_TIMEOUT` | `0s` | Per-process execution timeout (e.g. `5s`, `1m`) |
| `--fail-fast` | `GIT_ITER_FAIL_FAST` | `false` | Abort immediately when any command fails |
| `-d, --dirty-only` | - | `false` | Only process repositories with uncommitted/untracked changes |
| `-c, --clean-only` | - | `false` | Only process repositories without uncommitted/untracked changes |
| `--remote-url` | `GIT_ITER_REMOTE_URL` | `""` | Filter repositories by origin remote URL regex pattern |
| `--sort` | - | `false` | Sort discovered repository paths in ascending order |
| `--sort-reverse` | - | `false` | Sort discovered repository paths in descending order |
| `-n, --limit` | - | `0` | Limit processed repositories to the first n repos |
| `-a, --abs-path` | `GIT_ITER_ABS_PATH` | `false` | Show absolute repository paths instead of relative prefixes |
| `-l, --log-level` | `GIT_ITER_LOG_LEVEL` | `info` | Logging level (`debug`, `info`, `warn`, `error`) to stderr |
| `--git-command` | `GIT_ITER_GIT_COMMAND` | `git` | Path or name of the git binary |
| `--markdown-header-level` | `GIT_ITER_MARKDOWN_HEADER_LEVEL` | `3` | Heading level for markdown repo sections (1-6) |

---

## Common Examples

### 1. Execute Commands Across Repositories (`do`)

Arguments before `--` are regex patterns filtering repository paths. Arguments after `--` are the command and arguments to run:

```bash
# Run git status in all repositories under ~/src
git-iter -r ~/src do -- git status -s

# Run in parallel with 4 workers
git-iter -r ~/src -p 4 do -- git fetch --prune

# Only target dirty repositories matching regex "backend"
git-iter -r ~/src -d do backend -- git status -s

# Output structured JSON Lines
git-iter -r ~/src -o json do -- git rev-parse HEAD
```

### 2. Search Code (`grep`)

Prefixes each match with the repository path:

```bash
# Search for TODO in all repositories
git-iter -r ~/src grep -- TODO

# Forward git grep flags
git-iter -r ~/src grep -- -n -i "fixme"

# Generate a Markdown report
git-iter -r ~/src -o markdown grep -- -n "TODO" > todo_report.md
```

### 3. Check Diff Across Repositories (`diff`)

By default, `diff` runs with `--no-pager` so output streams cleanly without blocking:

```bash
# Show summary stats of uncommitted changes
git-iter -r ~/src diff -- --stat

# Only inspect dirty repositories
git-iter -r ~/src -d diff

# Run interactively with pager enabled
git-iter -r ~/src diff -i
```

### 4. Repository Status (`status` / `st`)

```bash
# Show short status across all repositories
git-iter -r ~/src status

# Emit JSON Lines for automation
git-iter -r ~/src -o json status
```

### 5. Repository Discovery & Pipelining (`list` & `read`)

```bash
# List all discovered repositories
git-iter -r ~/src list

# Narrow down with patterns and pipe into read
git-iter -r ~/src list "service-.*" | git-iter read git pull

# Run shell commands with flags (use "--" so flags like "-c" are not parsed as git-iter options)
git-iter -r ~/src list | git-iter read -- bash -c "git status -s"
```

### 6. Pull and Fetch Across Repositories

```bash
# Pull only the default branch (e.g. main/master) and switch back
git-iter -r ~/src pull --switch-default --switch-back

# Fetch only the default branch from origin
git-iter -r ~/src fetch --default
```

---

## Model Context Protocol (MCP) Server

`git-iter` can run as an [MCP](https://modelcontextprotocol.io/) server communicating over standard input/output (`stdio`), exposing tools to AI assistants (Claude Desktop, Cursor, Antigravity, etc.):

```bash
git-iter -r ~/src mcp
```

### Available MCP Tools

- `git_iter_list`: Discover and list repositories under the configured root.
- `git_iter_status`: Run `git status` in matching repositories.
- `git_iter_diff`: Run `git diff` in matching repositories.
- `git_iter_grep`: Run `git grep` across matching repositories.
- `git_iter_ls_files`: Run `git ls-files` across matching repositories.
- `git_iter_do`: Execute arbitrary commands in matching repositories (**disabled by default for security**, opt-in via `--enable-do`).

To enable arbitrary command execution for the MCP server:

```bash
git-iter -r ~/src mcp --enable-do
```

---

## Development

Requires Go 1.24+ and [mise](https://mise.jdx.dev/).

```bash
# Run lint, test, and build (default)
mise run default

# Run tests with race detection and coverage
mise run test

# Run linters
mise run lint

# Check third-party license notices
mise run licenses
```

---

## License

MIT License. See [LICENSE](LICENSE) and [NOTICE](NOTICE) for third-party software notices.
