package markdown_test

import (
	"bytes"
	"testing"

	"github.com/berquerant/git-iter-go/internal/markdown"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetails(t *testing.T) {
	t.Parallel()

	t.Run("FormatDetails", func(t *testing.T) {
		t.Parallel()
		got := markdown.FormatDetails("Files", "- a.go\n- b.go")
		want := "<details>\n<summary>Files</summary>\n\n- a.go\n- b.go\n\n</details>\n\n"
		assert.Equal(t, want, got)
	})

	t.Run("FormatCodeDetails", func(t *testing.T) {
		t.Parallel()
		got := markdown.FormatCodeDetails("stdout", "hello world\n")
		want := "<details>\n<summary>stdout</summary>\n\n```\nhello world\n```\n\n</details>\n\n"
		assert.Equal(t, want, got)
	})

	t.Run("FormatCodeDetails without trailing newline", func(t *testing.T) {
		t.Parallel()
		got := markdown.FormatCodeDetails("stderr", "warning text")
		want := "<details>\n<summary>stderr</summary>\n\n```\nwarning text\n```\n\n</details>\n\n"
		assert.Equal(t, want, got)
	})

	t.Run("Details WriteTo", func(t *testing.T) {
		t.Parallel()
		d := markdown.Details{Summary: "Summary", Content: "Body"}
		var buf bytes.Buffer
		n, err := d.WriteTo(&buf)
		require.NoError(t, err)
		assert.Equal(t, int64(buf.Len()), n)
		assert.Equal(t, "<details>\n<summary>Summary</summary>\n\nBody\n\n</details>\n\n", buf.String())
	})
}

