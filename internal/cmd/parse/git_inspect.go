package parse

// LsFilesArgs is the parsed argument set for the "ls-files" subcommand.
type LsFilesArgs struct {
	Patterns     []string // repository filter regexps; empty means all
	LsFilesFlags []string // optional arguments forwarded to "git ls-files"
}

// ParseLsFilesArgs parses raw cobra args into LsFilesArgs.
func ParseLsFilesArgs(args []string) LsFilesArgs {
	return ParseLsFilesArgsWithDash(args, -1)
}

// ParseLsFilesArgsWithDash parses args taking into account Cobra's ArgsLenAtDash().
func ParseLsFilesArgsWithDash(args []string, dashIndex int) LsFilesArgs {
	patterns, lsFilesFlags := SplitWithDash(args, dashIndex)
	return LsFilesArgs{Patterns: patterns, LsFilesFlags: lsFilesFlags}
}

// StatusArgs is the parsed argument set for the "status" subcommand.
type StatusArgs struct {
	Patterns    []string // repository filter regexps; empty means all
	StatusFlags []string // optional arguments forwarded to "git status"
}

// ParseStatusArgs parses raw cobra args into StatusArgs.
func ParseStatusArgs(args []string) StatusArgs {
	return ParseStatusArgsWithDash(args, -1)
}

// ParseStatusArgsWithDash parses args taking into account Cobra's ArgsLenAtDash().
func ParseStatusArgsWithDash(args []string, dashIndex int) StatusArgs {
	patterns, statusFlags := SplitWithDash(args, dashIndex)
	return StatusArgs{Patterns: patterns, StatusFlags: statusFlags}
}

// DiffArgs is the parsed argument set for the "diff" subcommand.
type DiffArgs struct {
	Patterns    []string // repository filter regexps; empty means all
	DiffFlags   []string // optional arguments forwarded to "git diff"
	Interactive bool     // whether to run git diff interactively (with pager)
}

// ParseDiffArgs parses raw cobra args into DiffArgs.
func ParseDiffArgs(args []string) DiffArgs {
	return ParseDiffArgsWithDash(args, -1)
}

// ParseDiffArgsWithDash parses args taking into account Cobra's ArgsLenAtDash().
func ParseDiffArgsWithDash(args []string, dashIndex int) DiffArgs {
	patterns, diffFlags := SplitWithDash(args, dashIndex)
	return DiffArgs{Patterns: patterns, DiffFlags: diffFlags}
}
