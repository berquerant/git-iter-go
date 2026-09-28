// Package config defines the global configuration for git-iter-go.
// Settings are resolved in priority order: default < environment variable < CLI flag,
// using github.com/berquerant/structconfig.
package config

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/berquerant/structconfig"
	"github.com/spf13/pflag"
)

// Config holds all global settings for git-iter.
type Config struct {
	// MaxProcs is the maximum number of parallel processes.
	MaxProcs int `name:"max-procs" default:"1" usage:"maximum number of parallel processes" short:"p"`
	// ReposRoot is the root directory scanned for git repositories when ListRepos is empty.
	ReposRoot string `name:"repos-root" default:"" usage:"root directory of local repositories" short:"r"`
	// ListRepos is an optional shell command whose stdout lists absolute repository paths.
	// When set, ReposRoot is ignored for discovery.
	ListRepos string `name:"list-repos" default:"" usage:"command to list absolute paths of repositories (one per line)"`
	// AbsPath controls whether grep/path-annotated output shows absolute paths (true) or paths
	// relative to ReposRoot (false).
	AbsPath bool `name:"abs-path" default:"false" usage:"show absolute paths in command output" short:"a"`
	// FailFast aborts execution immediately when an error occurs in any repository.
	FailFast bool `name:"fail-fast" default:"false" usage:"abort immediately when any repository command fails"`
	// LogLevel sets the logging level (debug, info, warn, error).
	LogLevel string `name:"log-level" default:"info" usage:"logging level (debug, info, warn, error)" short:"l"`
	// GitCommand is the path or name of the git binary.
	GitCommand string `name:"git-command" default:"git" usage:"path to git executable"`
	// Timeout is the maximum execution duration for each process.
	Timeout time.Duration `name:"timeout" default:"0s" usage:"timeout for each process execution (e.g. 5s, 1m)"`
	// MarkdownHeaderLevel is the heading level (1-6) for repository sections in markdown output.
	MarkdownHeaderLevel int `name:"markdown-header-level" default:"3" usage:"heading level for markdown sections (1-6)"`
	// Sort sorts repository paths in ascending order.
	Sort bool `name:"sort" default:"false" usage:"sort repository paths in ascending order"`
	// SortReverse sorts repository paths in descending order.
	SortReverse bool `name:"sort-reverse" default:"false" usage:"sort repository paths in descending order"`
	// Limit limits the number of processed repositories to the first n repos (after sorting if applicable).
	Limit int `name:"limit" default:"0" usage:"limit the number of processed repositories" short:"n"`
	// DirtyOnly filters repositories to only those with uncommitted changes, staged changes, or untracked files.
	DirtyOnly bool `name:"dirty-only" default:"false" usage:"only include dirty repositories" short:"d"`
	// CleanOnly filters repositories to only those without uncommitted changes or untracked files.
	CleanOnly bool `name:"clean-only" default:"false" usage:"only include clean repositories" short:"c"`
	// RemoteURL filters repositories to only those whose origin remote URL matches the specified regex pattern.
	RemoteURL string `name:"remote-url" default:"" usage:"filter repositories by origin remote URL regex pattern"`
	// DefaultBranchOnly filters repositories to only those whose current branch is the default branch.
	DefaultBranchOnly bool `name:"default-branch-only" default:"false" usage:"only include repositories whose current branch is the default branch"`
	// NotDefaultBranchOnly filters repositories to only those whose current branch is not the default branch.
	NotDefaultBranchOnly bool `name:"not-default-branch-only" default:"false" usage:"only include repositories whose current branch is not the default branch"`
	// Branch filters repositories to only those whose current branch matches the specified regex pattern.
	Branch string `name:"branch" default:"" usage:"filter repositories by current branch regex pattern"`
}

const envPrefix = "GIT_ITER_"

// SetFlags registers all Config fields as flags on fs.
func SetFlags(fs *pflag.FlagSet) error {
	sc := structconfig.New[Config](structconfig.WithEnvPrefix(envPrefix))
	return sc.SetFlags(fs)
}

// Resolve builds a Config by merging: default < env < already-parsed flags.
// fs must already be parsed before calling Resolve.
func Resolve(fs *pflag.FlagSet) (*Config, error) {
	sc := structconfig.New[Config](structconfig.WithEnvPrefix(envPrefix))
	merger := structconfig.NewMerger[Config]()

	// Step 1: defaults
	var def Config
	if err := sc.FromDefault(&def); err != nil {
		return nil, err
	}

	// Step 2: env vars (fall back to default tag when env is unset)
	var fromEnv Config
	if err := sc.FromEnv(&fromEnv); err != nil {
		return nil, err
	}

	// Step 3: flags (already parsed)
	var fromFlags Config
	if err := sc.FromFlags(&fromFlags, fs); err != nil {
		return nil, err
	}

	// Merge: default < env < flags
	merged, err := merger.Merge(def, fromEnv)
	if err != nil {
		return nil, err
	}
	merged, err = merger.Merge(merged, fromFlags)
	if err != nil {
		return nil, err
	}
	return &merged, nil
}

// GitCommandOrFallback returns GitCommand if set, otherwise "git".
func (c *Config) GitCommandOrFallback() string {
	if c != nil && c.GitCommand != "" {
		return c.GitCommand
	}
	return "git"
}

// ParseLogLevel converts a level string (e.g. "debug", "info", "warn", "error")
// to slog.Level. Defaults to slog.LevelInfo if empty.
func ParseLogLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("config: unknown log level: %q", s)
	}
}
