package markdown_test

import (
	"bytes"
	"testing"

	"github.com/berquerant/git-iter-go/internal/markdown"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadingPrefix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		level int
		want  string
	}{
		{level: 0, want: "#"},
		{level: 1, want: "#"},
		{level: 3, want: "###"},
		{level: 6, want: "######"},
		{level: 7, want: "######"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, markdown.HeadingPrefix(tt.level))
	}
}

func TestWriteHeader(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := markdown.WriteHeader(&buf, "test title", 3)
	require.NoError(t, err)
	assert.Equal(t, "## test title\n\n", buf.String())
}
