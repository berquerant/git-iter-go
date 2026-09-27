// Package cmd defines all cobra commands for git-iter.
package cmd

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"slices"
	"sync"

	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// syncWriter wraps an io.Writer with a mutex to ensure thread-safe writes.
type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (s *syncWriter) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}

// ErrNoRepoSource is returned when neither --repos-root nor --list-repos is configured.
var ErrNoRepoSource = errors.New("either --repos-root (-r) or --list-repos must be set")

// configKey is the context key used to pass the resolved Config to sub-commands.
type configKey struct{}

// outModeKey is the context key used to pass the output mode to sub-commands.
type outModeKey struct{}

// ConfigFromContext retrieves the Config stored in ctx by NewRootCmd's PersistentPreRunE.
func ConfigFromContext(ctx context.Context) *config.Config {
	v, _ := ctx.Value(configKey{}).(*config.Config)
	return v
}

// OutModeFromContext retrieves the output mode ("text" or "json") from ctx.
// Returns "text" when not set.
func OutModeFromContext(ctx context.Context) string {
	v, _ := ctx.Value(outModeKey{}).(string)
	if v == "" {
		return "text"
	}
	return v
}

// NewRootCmd constructs the root cobra command with all sub-commands registered.
// Each call creates a fresh pflag.FlagSet so that tests can create multiple
// independent command trees without flag-redefinition panics.
func NewRootCmd() *cobra.Command {
	fs := pflag.NewFlagSet("git-iter", pflag.ContinueOnError)
	if err := config.SetFlags(fs); err != nil {
		panic(err)
	}

	var outMode string

	root := &cobra.Command{
		Use:          "git-iter",
		Short:        "Run commands on multiple local git repositories",
		SilenceUsage: true,
		Long: `git-iter — run commands across multiple local git repositories.

Repository source (one of the following is required):
  --repos-root / -r   Walk this directory tree and find every git repository
                      (a directory that contains a .git entry).
  --list-repos        Shell command whose stdout lists absolute repository paths,
                      one per line. When set, --repos-root is ignored.

Environment variables (all prefixed with GIT_ITER_):
  GIT_ITER_MAX_PROCS      Maximum parallel processes (default: 1).
  GIT_ITER_REPOS_ROOT     Equivalent to --repos-root.
  GIT_ITER_LIST_REPOS     Equivalent to --list-repos.
  GIT_ITER_ABS_PATH       Equivalent to --abs-path (set to any non-empty
                          value to enable).
  GIT_ITER_FAIL_FAST      Equivalent to --fail-fast (set to any non-empty
                          value to enable).
  GIT_ITER_LOG_LEVEL      Equivalent to --log-level (debug, info, warn, error).
  GIT_ITER_GIT_COMMAND          Equivalent to --git-command (default: git).
  GIT_ITER_TIMEOUT              Equivalent to --timeout (e.g. 5s, 1m).
  GIT_ITER_MARKDOWN_HEADER_LEVEL Equivalent to --markdown-header-level (1-6, default: 3).

Flag resolution order (highest priority wins):
  CLI flag  >  environment variable  >  built-in default`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Resolve(fs)
			if err != nil {
				return err
			}
			if cfg.ReposRoot == "" && cfg.ListRepos == "" && !slices.Contains([]string{
				"version",
				"help",
				"read",
				"mcp",
			}, cmd.Name()) {
				return ErrNoRepoSource
			}

			logLevel, err := config.ParseLogLevel(cfg.LogLevel)
			if err != nil {
				return err
			}
			handler := slog.NewTextHandler(&syncWriter{w: cmd.ErrOrStderr()}, &slog.HandlerOptions{Level: logLevel})
			slog.SetDefault(slog.New(handler))

			ctx := context.WithValue(cmd.Context(), configKey{}, cfg)
			ctx = context.WithValue(ctx, outModeKey{}, outMode)
			cmd.SetContext(ctx)
			return nil
		},
	}

	// Mirror our FlagSet into cobra's PersistentFlags so cobra parses them.
	root.PersistentFlags().AddFlagSet(fs)
	root.PersistentFlags().StringVarP(&outMode, "out", "o", "text",
		`output mode: "text" (stream), "json" (JSON Lines), or "markdown"/"md"`)

	root.AddCommand(
		NewDoCmd(),
		NewGrepCmd(),
		NewListCmd(),
		NewReadCmd(),
		NewPullCmd(),
		NewFetchCmd(),
		NewLsFilesCmd(),
		NewStatusCmd(),
		NewDiffCmd(),
		NewMcpCmd(),
		NewVersionCmd(),
	)

	return root
}
