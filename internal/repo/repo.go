// Package repo provides types for discovering local git repositories.
package repo

import (
	"bufio"
	"context"
	"io"
	"strings"
)

// Finder returns a list of absolute paths to local git repositories.
type Finder interface {
	Find(ctx context.Context) ([]string, error)
}

// StdinFinder reads repository paths (one per line) from Reader.
type StdinFinder struct {
	Reader io.Reader
}

func (f *StdinFinder) Find(ctx context.Context) ([]string, error) {
	return scanLinesWithContext(ctx, f.Reader)
}

// scanLinesWithContext reads non-empty, trimmed lines from r while checking ctx.
func scanLinesWithContext(ctx context.Context, r io.Reader) ([]string, error) {
	sc := bufio.NewScanner(r)
	var lines []string
	for sc.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return lines, sc.Err()
}

// splitLines splits s by newlines, trimming whitespace and ignoring blank lines.
func splitLines(ctx context.Context, s string) ([]string, error) {
	return scanLinesWithContext(ctx, strings.NewReader(s))
}
