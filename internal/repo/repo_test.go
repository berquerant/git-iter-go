package repo_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/berquerant/git-iter-go/internal/repo"
	"github.com/stretchr/testify/require"
)

func TestStdinFinder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single line",
			input: "/repo\n",
			want:  []string{"/repo"},
		},
		{
			name:  "multiple lines",
			input: "/a\n/b\n/c\n",
			want:  []string{"/a", "/b", "/c"},
		},
		{
			name:  "blank lines ignored",
			input: "/a\n\n/b\n",
			want:  []string{"/a", "/b"},
		},
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := &repo.StdinFinder{Reader: strings.NewReader(tt.input)}
			got, err := f.Find(context.Background())
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}

	t.Run("context cancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		f := &repo.StdinFinder{Reader: strings.NewReader("/a\n/b\n")}
		_, err := f.Find(ctx)
		require.ErrorIs(t, err, context.Canceled)
	})
}

type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestStdinFinder_Error(t *testing.T) {
	t.Parallel()
	f := &repo.StdinFinder{Reader: errReader{}}
	_, err := f.Find(context.Background())
	require.Error(t, err)
}
