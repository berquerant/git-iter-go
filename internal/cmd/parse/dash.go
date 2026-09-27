package parse

// SplitWithDash splits args into patterns (before dash) and command/flags (after dash).
// If dashIndex >= 0 (provided by Cobra ArgsLenAtDash), it splits at that index.
// Otherwise, it falls back to looking for a literal "--" element in args.
func SplitWithDash(args []string, dashIndex int) (patterns, command []string) {
	if dashIndex >= 0 && dashIndex <= len(args) {
		return args[:dashIndex], args[dashIndex:]
	}
	for i, a := range args {
		if a == "--" {
			return args[:i], args[i+1:]
		}
	}
	return args, nil
}
