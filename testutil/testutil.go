// Package testutil provides shared test helpers: fake implementations of
// internal/repo.Finder and internal/runner.Runner interfaces.
package testutil

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

// FakeFinder is a test double for repo.Finder that returns a pre-configured list of paths.
type FakeFinder struct {
	Paths []string
	Err   error
}

func (f *FakeFinder) Find(_ context.Context) ([]string, error) {
	return f.Paths, f.Err
}

// RunCall records a single invocation of FakeRunner.Run.
type RunCall struct {
	Dir     string
	Command []string
}

// FakeRunner is a test double for runner.Runner.
// It records every call and writes Output to stdout; returns Err on each call.
type FakeRunner struct {
	mu     sync.Mutex
	Calls  []RunCall
	Output string
	Err    error
}

func (r *FakeRunner) Run(_ context.Context, dir string, command []string, stdout, _ io.Writer) error {
	r.mu.Lock()
	r.Calls = append(r.Calls, RunCall{Dir: dir, Command: command})
	r.mu.Unlock()
	if r.Output != "" {
		_, _ = io.WriteString(stdout, r.Output)
	}
	return r.Err
}

// MakeGitRepo creates a minimal git repository directory (with a .git dir) under parent and returns its path.
func MakeGitRepo(t testing.TB, parent, name string) string {
	t.Helper()
	dir := filepath.Join(parent, name)
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o750); err != nil {
		t.Fatalf("MakeGitRepo failed: %v", err)
	}
	return dir
}

// SubcommandHandler is invoked by SubcommandRunner when a matching subcommand is executed.
type SubcommandHandler func(cmd []string, stdout, stderr io.Writer) error

// SubcommandRunner is a mock runner that dispatches to handlers based on command[1].
type SubcommandRunner struct {
	mu       sync.Mutex
	calls    [][]string
	handlers map[string]SubcommandHandler
}

// NewSubcommandRunner creates a new SubcommandRunner.
func NewSubcommandRunner() *SubcommandRunner {
	return &SubcommandRunner{
		handlers: make(map[string]SubcommandHandler),
	}
}

// On registers a handler for a given git subcommand (e.g. "rev-parse", "symbolic-ref").
func (r *SubcommandRunner) On(subcmd string, h SubcommandHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[subcmd] = h
}

// Run records the call and executes the registered handler if found.
func (r *SubcommandRunner) Run(_ context.Context, _ string, command []string, stdout, stderr io.Writer) error {
	r.mu.Lock()
	r.calls = append(r.calls, slices.Clone(command))
	var h SubcommandHandler
	if len(command) > 1 {
		h = r.handlers[command[1]]
	}
	r.mu.Unlock()

	if h != nil {
		return h(command, stdout, stderr)
	}
	return nil
}

// Calls returns a copy of all recorded calls.
func (r *SubcommandRunner) Calls() [][]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.calls)
}

// RunTestMain unsets all GIT_ITER_* environment variables, runs the test suite m,
// restores the original environment variables, and terminates the process with m's exit code.
// TestMain implementations in packages can simply call testutil.RunTestMain(m).
func RunTestMain(m *testing.M) {
	orig := make(map[string]string)
	for _, env := range os.Environ() {
		idx := strings.IndexByte(env, '=')
		if idx <= 0 {
			continue
		}
		key := env[:idx]
		if strings.HasPrefix(key, "GIT_ITER_") {
			orig[key] = env[idx+1:]
			_ = os.Unsetenv(key)
		}
	}
	code := m.Run()
	for key, val := range orig {
		_ = os.Setenv(key, val)
	}
	os.Exit(code)
}
