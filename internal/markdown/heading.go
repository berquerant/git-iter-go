package markdown

import (
	"fmt"
	"io"
	"strings"
)

// HeadingPrefix returns the markdown heading prefix for a given level (clamped to 1-6).
func HeadingPrefix(level int) string {
	if level < 1 {
		level = 1
	} else if level > 6 {
		level = 6
	}
	return strings.Repeat("#", level)
}

// WriteHeader writes a top-level Markdown heading for a command execution.
// titleLevel is typically headerLevel - 1 (clamped to 1-6).
func WriteHeader(w io.Writer, title string, headerLevel int) error {
	topLevel := max(headerLevel-1, 1)
	h := HeadingPrefix(topLevel)
	_, err := fmt.Fprintf(w, "%s %s\n\n", h, title)
	return err
}
