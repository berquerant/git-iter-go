package markdown

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/berquerant/git-iter-go/internal/output/common"
)

// ParseResultFiles parses stdout lines into non-empty file paths.
func ParseResultFiles(stdout string) []string {
	var files []string
	for line := range strings.SplitSeq(stdout, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			files = append(files, trimmed)
		}
	}
	return files
}

// WriteLsFiles writes an individual repository's ls-files Result formatted as Markdown.
func WriteLsFiles(w io.Writer, r common.Result, headerLevel int) error {
	repo := r.RepoRelPath
	if repo == "" {
		repo = r.RepoAbsPath
	}

	h := HeadingPrefix(headerLevel)
	files := ParseResultFiles(r.Stdout)

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s %s\n\n", h, repo)
	fmt.Fprintf(&sb, "Total: %d\n\n", len(files))

	if len(files) > 0 {
		var content strings.Builder
		for _, f := range files {
			fmt.Fprintf(&content, "- %s\n", f)
		}
		sb.WriteString(FormatDetails("Files", content.String()))
	}

	if r.Stderr != "" {
		sb.WriteString(FormatCodeDetails("stderr", r.Stderr))
	}

	_, err := io.WriteString(w, sb.String())
	return err
}

// WriteLsFilesSummary writes an aggregate summary for multiple ls-files Results.
func WriteLsFilesSummary(w io.Writer, results []common.Result, headerLevel int) error {
	sumLevel := max(headerLevel-1, 1)
	h := HeadingPrefix(sumLevel)

	totalFiles := 0
	counts := make(map[int]int)
	var failed []common.Result
	for _, r := range results {
		counts[r.ExitCode]++
		if r.ExitCode != 0 {
			failed = append(failed, r)
		}
		totalFiles += len(ParseResultFiles(r.Stdout))
	}

	codes := make([]int, 0, len(counts))
	for c := range counts {
		codes = append(codes, c)
	}
	sort.Ints(codes)

	var sb strings.Builder
	sb.WriteString("---\n\n")
	fmt.Fprintf(&sb, "%s Summary\n\n", h)
	fmt.Fprintf(&sb, "- **Total Repositories**: %d\n", len(results))
	fmt.Fprintf(&sb, "- **Total Files**: %d\n", totalFiles)
	sb.WriteString("- **Exit Codes**:\n")
	for _, c := range codes {
		fmt.Fprintf(&sb, "  - Exit Code `%d`: %d\n", c, counts[c])
	}
	sb.WriteString("- **Failed Repositories**:\n")
	if len(failed) == 0 {
		sb.WriteString("  - None (All succeeded)\n")
	} else {
		for _, f := range failed {
			name := f.RepoRelPath
			if name == "" {
				name = f.RepoAbsPath
			}
			fmt.Fprintf(&sb, "  - `%s` (Exit Code `%d`)\n", name, f.ExitCode)
		}
	}
	sb.WriteString("\n")

	_, err := io.WriteString(w, sb.String())
	return err
}
