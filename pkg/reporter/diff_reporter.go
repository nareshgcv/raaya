package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/nareshgcv/raaya/pkg/analysis"
)

// Diff prints a capability diff for the terminal.
func Diff(w io.Writer, d *analysis.Diff, color bool) {
	s := style(color)
	fmt.Fprintln(w, s.wrap("1", "Raaya capability diff"))
	fmt.Fprintln(w)

	section := func(title string, changes []analysis.ReachChange, mark, code string) {
		if len(changes) == 0 {
			return
		}
		fmt.Fprintln(w, title+":")
		for _, c := range changes {
			var access string
			switch mark {
			case "~":
				access = permLabel(c.Before) + " → " + permLabel(c.After)
			case "-":
				access = permLabel(c.Before)
			default:
				access = permLabel(c.After)
			}
			fmt.Fprintf(w, "  %s %s → %s %q [%s]\n", s.wrap(code, mark), c.AgentID, strings.ToLower(string(c.NodeType)), c.NodeName, access)
		}
		fmt.Fprintln(w)
	}
	section("Escalated", d.Escalations, "~", "31")
	section("Newly reachable", d.NewlyReachable, "+", "33")
	section("No longer reachable", d.NoLongerReachable, "-", "32")

	if len(d.NewFindings) > 0 {
		fmt.Fprintln(w, "New findings:")
		for _, f := range d.NewFindings {
			fmt.Fprintf(w, "  %s %s  %s\n", s.severity(f.Severity), f.RuleID, f.Message)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "Graph: +%d/-%d nodes, +%d/-%d edges\n", len(d.AddedNodes), len(d.RemovedNodes), len(d.AddedEdges), len(d.RemovedEdges))
	if d.Regressions > 0 {
		fmt.Fprintln(w, s.wrap("31;1", fmt.Sprintf("%d regression(s)", d.Regressions)))
	} else {
		fmt.Fprintln(w, s.wrap("32", "No regressions."))
	}
}
