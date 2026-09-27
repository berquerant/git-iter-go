package cmd

import (
	"strings"
)

// Common output mode description constants.
const (
	DefaultTextDesc     = "Stream stdout/stderr of each process directly (default)."
	DefaultJSONDesc     = "Emit one JSON object per repository (JSON Lines):"
	DefaultMarkdownDesc = "Emit a formatted Markdown document with summaries (or md)."

	JSONRowRepoAbsPath = "repo_abs_path"
	JSONRowRepoRelPath = "repo_rel_path"
	JSONRowCommand     = "command"
	JSONRowExitCode    = "exit_code"
	JSONRowStdout      = "stdout"
	JSONRowStderr      = "stderr"

	JSONRowRepoAbsPathDesc = "absolute path of the repository"
	JSONRowRepoRelPathDesc = "path relative to --repos-root"
	JSONRowStderrDesc      = "captured standard error"

	ExampleDescJSONLines = "Collect results as JSON Lines"
)

// Item represents a term and its description (e.g. flag or field).
type Item struct {
	Term string // Term or name (e.g. "-d, --dirty-only", "repo_abs_path")
	Desc string // Description of the item
}

// Section represents a structured section in help documentation.
type Section struct {
	Title string // Section title (e.g. "Flags:", "Filter flags (global):")
	Body  string // Free-form body text (optional)
	Items []Item // Structured key-value items with automatic alignment (optional)
}

// Format renders the section to a formatted string.
func (s Section) Format() string {
	var b strings.Builder
	if title := strings.TrimSpace(s.Title); title != "" {
		b.WriteString(title)
	}

	if body := strings.TrimSpace(s.Body); body != "" {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(body)
	}

	if len(s.Items) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(FormatItems(s.Items, 2, 2))
	}

	return b.String()
}

// FormatItems aligns items into a 2-column list with indent and spacing.
func FormatItems(items []Item, indent, spacing int) string {
	if len(items) == 0 {
		return ""
	}

	maxTermLen := 0
	for _, it := range items {
		if l := len(it.Term); l > maxTermLen {
			maxTermLen = l
		}
	}

	indentStr := strings.Repeat(" ", indent)
	spaceStr := strings.Repeat(" ", spacing)

	var lines []string
	for _, it := range items {
		if it.Desc == "" {
			lines = append(lines, indentStr+it.Term)
			continue
		}
		pad := strings.Repeat(" ", maxTermLen-len(it.Term))
		lines = append(lines, indentStr+it.Term+pad+spaceStr+it.Desc)
	}
	return strings.Join(lines, "\n")
}

// OutputModeDoc describes the output modes supported by a subcommand.
type OutputModeDoc struct {
	Text     string // Description for text mode
	JSON     string // Description for json mode
	JSONRows []Item // Field descriptions for json mode
	Markdown string // Description for markdown/md mode
}

// DefaultOutputModeDoc returns the standard output mode documentation for process-executing commands.
func DefaultOutputModeDoc() OutputModeDoc {
	return OutputModeDoc{
		Text: DefaultTextDesc,
		JSON: DefaultJSONDesc,
		JSONRows: []Item{
			{Term: JSONRowRepoAbsPath, Desc: JSONRowRepoAbsPathDesc},
			{Term: JSONRowRepoRelPath, Desc: JSONRowRepoRelPathDesc},
			{Term: JSONRowCommand, Desc: "argv slice that was executed"},
			{Term: JSONRowExitCode, Desc: "OS exit code of the process"},
			{Term: JSONRowStdout, Desc: "captured standard output"},
			{Term: JSONRowStderr, Desc: JSONRowStderrDesc},
		},
		Markdown: DefaultMarkdownDesc,
	}
}

// LongDoc defines the structured sections of a subcommand's Long help text.
// Build() concatenates these sections in order with standardized formatting.
type LongDoc struct {
	// Header is the opening synopsis or description of what the subcommand does.
	Header string
	// RepoFilter indicates whether to include the standard REPO_REGEX filter explanation.
	RepoFilter bool
	// DashSeparation explains how arguments before and after "--" are separated (if applicable).
	DashSeparation string
	// Sections contains additional structured sections (flags, options, notes) rendered with Title/Body/Items.
	Sections []Section
	// ExecutionNote indicates whether to include standard notes about working directory and parallelism controls.
	ExecutionNote bool
	// OutputModes specifies the explanations for text, json, and markdown output formats.
	OutputModes OutputModeDoc
	// Examples contains command usage examples appended at the end of help text.
	Examples []Example
}

// Example represents a usage example consisting of a description and a command.
type Example struct {
	Desc    string // Description of what the example does (e.g. "Run git status in all repositories")
	Command string // The shell command invocation (e.g. "git-iter -r /repos do -- git status")
}

// Format formats the example as a comment description followed by the command.
func (e Example) Format() string {
	var b strings.Builder
	if desc := strings.TrimSpace(e.Desc); desc != "" {
		b.WriteString("# " + desc + "\n")
	}
	b.WriteString(strings.TrimSpace(e.Command))
	return b.String()
}

const (
	commonRepoFilterDoc = `REPO_REGEX is a Go regular expression applied to the absolute repository path.
Multiple patterns are joined with "|" (logical OR).
If no patterns are given, the command runs in every discovered repository.`

	commonExecutionNoteDoc = `The working directory is set to the repository root for each invocation.

Parallelism is controlled by --max-procs / GIT_ITER_MAX_PROCS (default: 1).`
)

// Build constructs the formatted Long documentation string.
func (d LongDoc) Build() string {
	var sections []string

	if h := strings.TrimSpace(d.Header); h != "" {
		sections = append(sections, h)
	}

	if d.RepoFilter {
		sections = append(sections, commonRepoFilterDoc)
	}

	if ds := strings.TrimSpace(d.DashSeparation); ds != "" {
		sections = append(sections, ds)
	}

	for _, s := range d.Sections {
		if formatted := s.Format(); formatted != "" {
			sections = append(sections, formatted)
		}
	}

	if d.ExecutionNote {
		sections = append(sections, commonExecutionNoteDoc)
	}

	sections = append(sections, formatOutputModes(d.OutputModes))

	if len(d.Examples) > 0 {
		var b strings.Builder
		b.WriteString("Examples:\n")
		for i, ex := range d.Examples {
			if i > 0 {
				b.WriteString("\n\n")
			}
			b.WriteString(ex.Format())
		}
		sections = append(sections, b.String())
	}

	return strings.Join(sections, "\n\n")
}

func formatOutputModes(o OutputModeDoc) string {
	var b strings.Builder
	b.WriteString("Output modes (global --out / -o):\n")

	textDesc := o.Text
	if textDesc == "" {
		textDesc = "Stream output directly (default)."
	}
	b.WriteString("  text      " + textDesc + "\n")

	jsonDesc := o.JSON
	if jsonDesc == "" {
		jsonDesc = "Emit one JSON object per repository (JSON Lines)."
	}
	b.WriteString("  json      " + jsonDesc)

	if len(o.JSONRows) > 0 {
		formattedRows := FormatItems(o.JSONRows, 14, 2)
		b.WriteString("\n" + formattedRows)
	}
	b.WriteString("\n")

	mdDesc := o.Markdown
	if mdDesc == "" {
		mdDesc = "Emit a formatted Markdown document with summaries (or md)."
	}
	b.WriteString("  markdown  " + mdDesc)

	return b.String()
}
