package parse

// PullArgs is the parsed argument set for the "pull" subcommand.
type PullArgs struct {
	Patterns  []string // repository filter regexps; empty means all
	PullFlags []string // optional arguments forwarded to "git pull"
}

// ParsePullArgs parses raw cobra args into PullArgs.
func ParsePullArgs(args []string) PullArgs {
	return ParsePullArgsWithDash(args, -1)
}

// ParsePullArgsWithDash parses args taking into account Cobra's ArgsLenAtDash().
func ParsePullArgsWithDash(args []string, dashIndex int) PullArgs {
	patterns, pullFlags := SplitWithDash(args, dashIndex)
	return PullArgs{Patterns: patterns, PullFlags: pullFlags}
}

// FetchArgs is the parsed argument set for the "fetch" subcommand.
type FetchArgs struct {
	Patterns   []string // repository filter regexps; empty means all
	FetchFlags []string // optional arguments forwarded to "git fetch"
}

// ParseFetchArgs parses raw cobra args into FetchArgs.
func ParseFetchArgs(args []string) FetchArgs {
	return ParseFetchArgsWithDash(args, -1)
}

// ParseFetchArgsWithDash parses args taking into account Cobra's ArgsLenAtDash().
func ParseFetchArgsWithDash(args []string, dashIndex int) FetchArgs {
	patterns, fetchFlags := SplitWithDash(args, dashIndex)
	return FetchArgs{Patterns: patterns, FetchFlags: fetchFlags}
}
