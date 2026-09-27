package markdown

import (
	"fmt"
	"io"
	"strings"
)

// Details represents an HTML <details><summary>...</summary>...</details> collapsible block.
type Details struct {
	Summary string
	Content string
}

// FormatDetails formats summary and content as a markdown <details> block.
func FormatDetails(summary, content string) string {
	return Details{Summary: summary, Content: content}.String()
}

// FormatCodeDetails formats summary and code text wrapped in a markdown code block within a <details> element.
func FormatCodeDetails(summary, code string) string {
	return NewCodeDetails(summary, code).String()
}

// NewCodeDetails creates a Details block with code content wrapped in triple backticks.
func NewCodeDetails(summary, code string) Details {
	var sb strings.Builder
	sb.WriteString("```\n")
	sb.WriteString(code)
	if !strings.HasSuffix(code, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString("```\n")
	return Details{
		Summary: summary,
		Content: sb.String(),
	}
}

// String formats the Details block into a string with a trailing double newline.
func (d Details) String() string {
	var sb strings.Builder
	_, _ = d.WriteTo(&sb)
	return sb.String()
}

// WriteTo writes the formatted <details> block to w with a trailing double newline,
// implementing io.WriterTo.
func (d Details) WriteTo(w io.Writer) (int64, error) {
	n, err := fmt.Fprintf(
		w,
		"<details>\n<summary>%s</summary>\n\n%s\n\n</details>\n\n",
		d.Summary,
		strings.TrimRight(d.Content, "\n"),
	)
	return int64(n), err
}
