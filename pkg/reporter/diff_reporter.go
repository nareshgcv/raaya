package reporter

import (
	"encoding/json"
	"fmt"
	"io"

	"raaya/pkg/analysis"
)

// RenderDiffJSON outputs structured JSON for pipeline integrations.
func RenderDiffJSON(w io.Writer, diff *analysis.CapabilityDiff) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(diff)
}

// RenderDiffTerminal prints a human-readable capability diff summary to the console.
func RenderDiffTerminal(w io.Writer, diff *analysis.CapabilityDiff) {
	fmt.Fprintln(w, "\n--- Raaya Capability Diff Report ---")

	if !diff.HasRegressions {
		fmt.Fprintln(w, "✅ No security surface regressions detected.")
		return
	}

	fmt.Fprintln(w, "⚠️  Security Surface Changes Detected:")
	fmt.Fprintf(w, "  + Added Nodes: %d\n", len(diff.AddedNodes))
	fmt.Fprintf(w, "  - Removed Nodes: %d\n", len(diff.RemovedNodes))
	fmt.Fprintf(w, "  + New Capabilities: %d\n", len(diff.AddedEdges))
	fmt.Fprintf(w, "  ⚡ Escalations: %d\n\n", len(diff.Escalations))

	if len(diff.Escalations) > 0 {
		fmt.Fprintln(w, "Permission Escalations:")
		for _, esc := range diff.Escalations {
			fmt.Fprintf(w, "  * %s -> %s [%s]\n", esc.SourceID, esc.TargetID, esc.Permission)
		}
	}
}
