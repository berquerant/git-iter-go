package parse

import (
	"fmt"
)

// GrepArgs is the parsed argument set for the "grep" subcommand.
type GrepArgs struct {
	Patterns    []string // repository filter regexps
	GitGrepArgs []string // arguments forwarded verbatim to "git grep"
}

// ParseGrepArgs parses raw cobra args into GrepArgs.
// Returns an error when no GIT_GREP_ARGS are provided after "--".
func ParseGrepArgs(args []string) (GrepArgs, error) {
	return ParseGrepArgsWithDash(args, -1)
}

// ParseGrepArgsWithDash parses args taking into account Cobra's ArgsLenAtDash().
func ParseGrepArgsWithDash(args []string, dashIndex int) (GrepArgs, error) {
	patterns, grepArgs := SplitWithDash(args, dashIndex)
	if len(grepArgs) == 0 {
		return GrepArgs{}, fmt.Errorf("grep: arguments required (separate from repo patterns with --)")
	}
	return GrepArgs{Patterns: patterns, GitGrepArgs: grepArgs}, nil
}
