package parse

import (
	"fmt"
)

// DoArgs is the parsed argument set for the "do" subcommand.
type DoArgs struct {
	Patterns []string // repository filter regexps
	Command  []string // command to run in each matching repo
}

// ParseDoArgs parses raw cobra args into DoArgs.
// Returns an error when no COMMAND is provided after "--".
func ParseDoArgs(args []string) (DoArgs, error) {
	return ParseDoArgsWithDash(args, -1)
}

// ParseDoArgsWithDash parses args taking into account Cobra's ArgsLenAtDash().
func ParseDoArgsWithDash(args []string, dashIndex int) (DoArgs, error) {
	patterns, command := SplitWithDash(args, dashIndex)
	if len(command) == 0 {
		return DoArgs{}, fmt.Errorf("do: command is required (separate from repo patterns with --)")
	}
	return DoArgs{Patterns: patterns, Command: command}, nil
}
