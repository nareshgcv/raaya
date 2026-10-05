package reporter

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/nareshgcv/raaya/pkg/analysis"
	"github.com/nareshgcv/raaya/pkg/blastradius"
	"github.com/nareshgcv/raaya/pkg/graph"
)

// IsTerminal reports whether colour output makes sense on f.
func IsTerminal(f *os.File) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

type style bool

func (s style) wrap(code, text string) string {
	if !s {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func (s style) severity(sev analysis.Severity) string {
	codes := map[analysis.Severity]string{analysis.SevHigh: "31;1", analysis.SevMedium: "33", analysis.SevLow: "36"}
	return s.wrap(codes[sev], fmt.Sprintf("%-6s", sev))
}

// Check prints a findings report.
func Check(w io.Writer, g *graph.Graph, findings []analysis.Finding, color bool) {
	s := style(color)
	fmt.Fprintln(w, s.wrap("1", "Raaya security analysis"))
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  Agents       %d\n", len(g.NodesOfType(graph.NodeAgent)))
	fmt.Fprintf(w, "  MCP servers  %d\n", len(g.NodesOfType(graph.NodeMCPServer)))
	fmt.Fprintf(w, "  Tools        %d\n", len(g.NodesOfType(graph.NodeTool)))
	fmt.Fprintf(w, "  Resources    %d\n\n", len(g.NodesOfType(graph.NodeResource)))

	if len(findings) == 0 {
		fmt.Fprintln(w, s.wrap("32", "No findings."))
		return
	}
	fmt.Fprintln(w, "Findings:")
	for _, f := range findings {
		fmt.Fprintf(w, "  %s %s  %s%s\n", s.severity(f.Severity), f.RuleID, f.Message, s.wrap("2", location(f)))
	}
	fmt.Fprintf(w, "\n%d finding(s)\n", len(findings))
}

// Blast prints a blast radius.
func Blast(w io.Writer, g *graph.Graph, r blastradius.Radius, color bool) {
	s := style(color)
	fmt.Fprintln(w, s.wrap("1", "Blast radius of "+r.SourceID))
	list := func(title string, entries []blastradius.Entry) {
		fmt.Fprintf(w, "\n%s (%d):\n", title, len(entries))
		for _, e := range entries {
			n := g.Nodes[e.NodeID]
			fmt.Fprintf(w, "  %-9s %s %q [%s]\n", strings.ToLower(string(n.Type)), e.NodeID, n.Name, permLabel(e.Permission))
		}
	}
	list("Direct", r.Direct)
	list("Transitive", r.Transitive)
	fmt.Fprintf(w, "\nScore: %d\n", r.Score)
}

func location(f analysis.Finding) string {
	if f.File == "" {
		return ""
	}
	if f.Line > 0 {
		return "  " + f.File + ":" + strconv.Itoa(f.Line)
	}
	return "  " + f.File
}

func permLabel(p graph.PermissionLevel) string {
	if p == graph.PermNone {
		return "reachable"
	}
	return string(p)
}
