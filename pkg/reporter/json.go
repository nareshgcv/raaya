// Package reporter renders findings, diffs and blast radius results as
// terminal text, JSON, SARIF and Markdown.
package reporter

import (
	"encoding/json"
	"io"
)

// JSON writes v as indented JSON.
func JSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
