package parse

// ListArgs is the parsed argument set for the "list" subcommand.
type ListArgs struct {
	Patterns []string // repository filter regexps; empty means all
}

// ParseListArgs parses raw cobra args into ListArgs. Never errors.
func ParseListArgs(args []string) ListArgs {
	return ListArgs{Patterns: args}
}

// ReadArgs is the parsed argument set for the "read" subcommand.
type ReadArgs struct {
	Command []string // command to run in each repo read from stdin
}

// ParseReadArgs parses raw cobra args into ReadArgs. Never errors;
// cobra MinimumNArgs(1) ensures args is non-empty before RunE is called.
func ParseReadArgs(args []string) ReadArgs {
	return ReadArgs{Command: args}
}
