package version

import (
	"fmt"
	"io"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func Write(w io.Writer) error {
	_, err := fmt.Fprintf(w, `Version: %s
Commit : %s
Date   : %s
`, Version, Commit, Date)
	return err
}
