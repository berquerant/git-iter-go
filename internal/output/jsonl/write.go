package jsonl

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/berquerant/git-iter-go/internal/output/common"
)

// WriteJSONL encodes r as a single compact JSON line (terminated by '\n') to w.
func WriteJSONL(w io.Writer, r common.Result) error {
	return writeJSON(w, r)
}

// WriteRepoJSONL encodes r as a single compact JSON line to w.
func WriteRepoJSONL(w io.Writer, r common.RepoResult) error {
	return writeJSON(w, r)
}

// WriteFileJSONL encodes r as a single compact JSON line to w.
func WriteFileJSONL(w io.Writer, r common.FileResult) error {
	return writeJSON(w, r)
}

func writeJSON(w io.Writer, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("output: marshal: %w", err)
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}
