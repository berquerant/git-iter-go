package config_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestParseLogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantLevel slog.Level
		wantErr   bool
	}{
		{name: "empty defaults to info", input: "", wantLevel: slog.LevelInfo},
		{name: "info lowercase", input: "info", wantLevel: slog.LevelInfo},
		{name: "INFO uppercase", input: "INFO", wantLevel: slog.LevelInfo},
		{name: "debug", input: "debug", wantLevel: slog.LevelDebug},
		{name: "DEBUG", input: "DEBUG", wantLevel: slog.LevelDebug},
		{name: "warn", input: "warn", wantLevel: slog.LevelWarn},
		{name: "warning", input: "warning", wantLevel: slog.LevelWarn},
		{name: "error", input: "error", wantLevel: slog.LevelError},
		{name: "invalid level", input: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := config.ParseLogLevel(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantLevel, got)
		})
	}
}

func resolveConfig(t *testing.T, flags []string, envKey, envVal string) *config.Config {
	t.Helper()
	if envKey != "" {
		t.Setenv(envKey, envVal)
	}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	require.NoError(t, config.SetFlags(fs))
	require.NoError(t, fs.Parse(flags))

	cfg, err := config.Resolve(fs)
	require.NoError(t, err)
	return cfg
}

func TestResolve_GitCommand(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		envKey    string
		envVal    string
		wantValue string
	}{
		{
			name:      "default is git",
			wantValue: "git",
		},
		{
			name:      "flag overrides default",
			flags:     []string{"--git-command", "/usr/local/bin/git"},
			wantValue: "/usr/local/bin/git",
		},
		{
			name:      "env overrides default",
			envKey:    "GIT_ITER_GIT_COMMAND",
			envVal:    "/opt/bin/git",
			wantValue: "/opt/bin/git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := resolveConfig(t, tt.flags, tt.envKey, tt.envVal)
			assert.Equal(t, tt.wantValue, cfg.GitCommand)
		})
	}
}

func TestResolve_Timeout(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		envKey    string
		envVal    string
		wantValue time.Duration
	}{
		{
			name:      "default is 0s",
			wantValue: time.Duration(0),
		},
		{
			name:      "flag overrides default",
			flags:     []string{"--timeout", "10s"},
			wantValue: 10 * time.Second,
		},
		{
			name:      "env overrides default",
			envKey:    "GIT_ITER_TIMEOUT",
			envVal:    "1m",
			wantValue: 1 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := resolveConfig(t, tt.flags, tt.envKey, tt.envVal)
			assert.Equal(t, tt.wantValue, cfg.Timeout)
		})
	}
}

func TestResolve_MarkdownHeaderLevel(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		envKey    string
		envVal    string
		wantValue int
	}{
		{
			name:      "default is 3",
			wantValue: 3,
		},
		{
			name:      "flag overrides default",
			flags:     []string{"--markdown-header-level", "2"},
			wantValue: 2,
		},
		{
			name:      "env overrides default",
			envKey:    "GIT_ITER_MARKDOWN_HEADER_LEVEL",
			envVal:    "4",
			wantValue: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := resolveConfig(t, tt.flags, tt.envKey, tt.envVal)
			assert.Equal(t, tt.wantValue, cfg.MarkdownHeaderLevel)
		})
	}
}

func TestResolve_Sort(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		envKey    string
		envVal    string
		wantValue bool
	}{
		{
			name:      "default is false",
			wantValue: false,
		},
		{
			name:      "flag overrides default",
			flags:     []string{"--sort"},
			wantValue: true,
		},
		{
			name:      "env overrides default",
			envKey:    "GIT_ITER_SORT",
			envVal:    "true",
			wantValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := resolveConfig(t, tt.flags, tt.envKey, tt.envVal)
			assert.Equal(t, tt.wantValue, cfg.Sort)
		})
	}
}

func TestResolve_SortReverse(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		envKey    string
		envVal    string
		wantValue bool
	}{
		{
			name:      "default is false",
			wantValue: false,
		},
		{
			name:      "flag overrides default",
			flags:     []string{"--sort-reverse"},
			wantValue: true,
		},
		{
			name:      "env overrides default",
			envKey:    "GIT_ITER_SORT_REVERSE",
			envVal:    "true",
			wantValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := resolveConfig(t, tt.flags, tt.envKey, tt.envVal)
			assert.Equal(t, tt.wantValue, cfg.SortReverse)
		})
	}
}

func TestResolve_AbsPath(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		envKey    string
		envVal    string
		wantValue bool
	}{
		{
			name:      "default is false",
			wantValue: false,
		},
		{
			name:      "--abs-path flag overrides default",
			flags:     []string{"--abs-path"},
			wantValue: true,
		},
		{
			name:      "-a flag overrides default",
			flags:     []string{"-a"},
			wantValue: true,
		},
		{
			name:      "env overrides default",
			envKey:    "GIT_ITER_ABS_PATH",
			envVal:    "true",
			wantValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := resolveConfig(t, tt.flags, tt.envKey, tt.envVal)
			assert.Equal(t, tt.wantValue, cfg.AbsPath)
		})
	}
}

func TestResolve_Limit(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		envKey    string
		envVal    string
		wantValue int
	}{
		{
			name:      "default is 0",
			wantValue: 0,
		},
		{
			name:      "--limit flag overrides default",
			flags:     []string{"--limit", "5"},
			wantValue: 5,
		},
		{
			name:      "-n flag overrides default",
			flags:     []string{"-n", "3"},
			wantValue: 3,
		},
		{
			name:      "env overrides default",
			envKey:    "GIT_ITER_LIMIT",
			envVal:    "10",
			wantValue: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := resolveConfig(t, tt.flags, tt.envKey, tt.envVal)
			assert.Equal(t, tt.wantValue, cfg.Limit)
		})
	}
}

func TestResolve_DirtyOnly(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		envKey    string
		envVal    string
		wantValue bool
	}{
		{
			name:      "default is false",
			wantValue: false,
		},
		{
			name:      "--dirty-only flag overrides default",
			flags:     []string{"--dirty-only"},
			wantValue: true,
		},
		{
			name:      "-d flag overrides default",
			flags:     []string{"-d"},
			wantValue: true,
		},
		{
			name:      "env overrides default",
			envKey:    "GIT_ITER_DIRTY_ONLY",
			envVal:    "true",
			wantValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := resolveConfig(t, tt.flags, tt.envKey, tt.envVal)
			assert.Equal(t, tt.wantValue, cfg.DirtyOnly)
		})
	}
}

func TestResolve_CleanOnly(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		envKey    string
		envVal    string
		wantValue bool
	}{
		{
			name:      "default is false",
			wantValue: false,
		},
		{
			name:      "--clean-only flag overrides default",
			flags:     []string{"--clean-only"},
			wantValue: true,
		},
		{
			name:      "-c flag overrides default",
			flags:     []string{"-c"},
			wantValue: true,
		},
		{
			name:      "env overrides default",
			envKey:    "GIT_ITER_CLEAN_ONLY",
			envVal:    "true",
			wantValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := resolveConfig(t, tt.flags, tt.envKey, tt.envVal)
			assert.Equal(t, tt.wantValue, cfg.CleanOnly)
		})
	}
}

func TestResolve_LogLevel(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		envKey    string
		envVal    string
		wantValue string
	}{
		{
			name:      "default is info",
			wantValue: "info",
		},
		{
			name:      "--log-level flag overrides default",
			flags:     []string{"--log-level", "debug"},
			wantValue: "debug",
		},
		{
			name:      "-l flag overrides default",
			flags:     []string{"-l", "warn"},
			wantValue: "warn",
		},
		{
			name:      "env overrides default",
			envKey:    "GIT_ITER_LOG_LEVEL",
			envVal:    "error",
			wantValue: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := resolveConfig(t, tt.flags, tt.envKey, tt.envVal)
			assert.Equal(t, tt.wantValue, cfg.LogLevel)
		})
	}
}

func TestResolve_RemoteURL(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		envKey    string
		envVal    string
		wantValue string
	}{
		{
			name:      "default is empty",
			wantValue: "",
		},
		{
			name:      "--remote-url flag overrides default",
			flags:     []string{"--remote-url", `github\.com`},
			wantValue: `github\.com`,
		},
		{
			name:      "env overrides default",
			envKey:    "GIT_ITER_REMOTE_URL",
			envVal:    `gitlab\.com`,
			wantValue: `gitlab\.com`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := resolveConfig(t, tt.flags, tt.envKey, tt.envVal)
			assert.Equal(t, tt.wantValue, cfg.RemoteURL)
		})
	}
}

func TestResolve_DefaultBranchOnly(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		envKey    string
		envVal    string
		wantValue bool
	}{
		{
			name:      "default is false",
			wantValue: false,
		},
		{
			name:      "--default-branch-only flag overrides default",
			flags:     []string{"--default-branch-only"},
			wantValue: true,
		},
		{
			name:      "env overrides default",
			envKey:    "GIT_ITER_DEFAULT_BRANCH_ONLY",
			envVal:    "true",
			wantValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := resolveConfig(t, tt.flags, tt.envKey, tt.envVal)
			assert.Equal(t, tt.wantValue, cfg.DefaultBranchOnly)
		})
	}
}

func TestResolve_NotDefaultBranchOnly(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		envKey    string
		envVal    string
		wantValue bool
	}{
		{
			name:      "default is false",
			wantValue: false,
		},
		{
			name:      "--not-default-branch-only flag overrides default",
			flags:     []string{"--not-default-branch-only"},
			wantValue: true,
		},
		{
			name:      "env overrides default",
			envKey:    "GIT_ITER_NOT_DEFAULT_BRANCH_ONLY",
			envVal:    "true",
			wantValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := resolveConfig(t, tt.flags, tt.envKey, tt.envVal)
			assert.Equal(t, tt.wantValue, cfg.NotDefaultBranchOnly)
		})
	}
}

func TestResolve_Branch(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		envKey    string
		envVal    string
		wantValue string
	}{
		{
			name:      "default is empty",
			wantValue: "",
		},
		{
			name:      "--branch flag overrides default",
			flags:     []string{"--branch", `feature/.*`},
			wantValue: `feature/.*`,
		},
		{
			name:      "env overrides default",
			envKey:    "GIT_ITER_BRANCH",
			envVal:    `bugfix/.*`,
			wantValue: `bugfix/.*`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := resolveConfig(t, tt.flags, tt.envKey, tt.envVal)
			assert.Equal(t, tt.wantValue, cfg.Branch)
		})
	}
}

