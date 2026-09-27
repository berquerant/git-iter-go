package markdown

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/berquerant/git-iter-go/internal/output/common"
)

// WriteResult writes Result formatted as a Markdown section.
// The heading is "<h> <repo>: exit code = <code>" using headerLevel (clamped to 1-6).
// stdout/stderr are wrapped in collapsible <details> blocks.
func WriteResult(w io.Writer, r common.Result, headerLevel int) error {
	repo := r.RepoRelPath
	if repo == "" {
		repo = r.RepoAbsPath
	}

	h := HeadingPrefix(headerLevel)
	var sb strings.Builder

	fmt.Fprintf(&sb, "%s %s: exit code = %d\n\n", h, repo, r.ExitCode)

	if r.Stdout != "" {
		sb.WriteString(FormatCodeDetails("stdout", r.Stdout))
	}

	if r.Stderr != "" {
		sb.WriteString(FormatCodeDetails("stderr", r.Stderr))
	}

	_, err := io.WriteString(w, sb.String())
	return err
}

// WriteSummary writes an aggregate summary for multiple Results.
func WriteSummary(w io.Writer, results []common.Result, headerLevel int) error {
	sumLevel := max(headerLevel-1, 1)
	h := HeadingPrefix(sumLevel)

	counts := make(map[int]int)
	var failed []common.Result
	for _, r := range results {
		counts[r.ExitCode]++
		if r.ExitCode != 0 {
			failed = append(failed, r)
		}
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

// WriteRepo writes RepoResult formatted as a Markdown section.
func WriteRepo(w io.Writer, r common.RepoResult, headerLevel int) error {
	repo := r.RepoRelPath
	if repo == "" {
		repo = r.RepoAbsPath
	}

	h := HeadingPrefix(headerLevel)
	_, err := fmt.Fprintf(w, "%s %s\n\n", h, repo)
	return err
}

// WriteList writes the complete list output in Markdown format:
// a top heading, a bullet list of repositories, and a summary section.
func WriteList(w io.Writer, results []common.RepoResult, reposRoot string, headerLevel int) error {
	topLevel := max(headerLevel-1, 1)
	hTop := HeadingPrefix(topLevel)
	hSum := HeadingPrefix(headerLevel)

	var sb strings.Builder
	if reposRoot != "" {
		fmt.Fprintf(&sb, "%s List: root = %s\n\n", hTop, reposRoot)
	} else {
		fmt.Fprintf(&sb, "%s List\n\n", hTop)
	}

	for _, r := range results {
		repo := r.RepoRelPath
		if repo == "" {
			repo = r.RepoAbsPath
		}
		fmt.Fprintf(&sb, "- %s\n", repo)
	}
	if len(results) > 0 {
		sb.WriteString("\n")
	}

	sb.WriteString("---\n\n")
	fmt.Fprintf(&sb, "%s Summary\n\n", hSum)
	fmt.Fprintf(&sb, "- **Total Repositories**: %d\n\n", len(results))

	_, err := io.WriteString(w, sb.String())
	return err
}
