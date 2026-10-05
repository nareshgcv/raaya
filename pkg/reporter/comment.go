package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/nareshgcv/raaya/pkg/analysis"
	"github.com/nareshgcv/raaya/pkg/graph"
)

// CommentMarker is embedded in PR comments so a bot can find and update its
// previous comment instead of posting a new one on every push.
const CommentMarker = "<!-- raaya-capability-diff -->"

func mdCell(s string) string {
	return strings.NewReplacer("|", `\|`, "\n", " ").Replace(s)
}

// CheckMarkdown writes findings as Markdown.
func CheckMarkdown(w io.Writer, g *graph.Graph, findings []analysis.Finding) {
	fmt.Fprintln(w, "## Raaya security analysis")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%d agent(s), %d MCP server(s), %d tool(s), %d resource(s).\n\n",
		len(g.NodesOfType(graph.NodeAgent)), len(g.NodesOfType(graph.NodeMCPServer)),
		len(g.NodesOfType(graph.NodeTool)), len(g.NodesOfType(graph.NodeResource)))
	if len(findings) == 0 {
		fmt.Fprintln(w, "No findings.")
		return
	}
	fmt.Fprintln(w, "| Severity | Rule | Finding | Location |")
	fmt.Fprintln(w, "|---|---|---|---|")
	for _, f := range findings {
		fmt.Fprintf(w, "| %s | %s | %s | %s |\n", f.Severity, f.RuleID, mdCell(f.Message), mdCell(strings.TrimSpace(location(f))))
	}
}

// DiffMarkdown writes a capability diff as a GitHub PR comment.
func DiffMarkdown(w io.Writer, d *analysis.Diff) {
	fmt.Fprintln(w, CommentMarker)
	fmt.Fprintln(w, "## Raaya capability diff")
	fmt.Fprintln(w)
	if len(d.NewlyReachable)+len(d.Escalations)+len(d.NoLongerReachable)+len(d.NewFindings) == 0 {
		fmt.Fprintln(w, "No change to what agents can reach.")
		return
	}
	fmt.Fprintf(w, "**%d regression(s)** · %d newly reachable · %d escalated · %d no longer reachable\n\n",
		d.Regressions, len(d.NewlyReachable), len(d.Escalations), len(d.NoLongerReachable))

	table := func(title string, changes []analysis.ReachChange, withBefore bool) {
		if len(changes) == 0 {
			return
		}
		fmt.Fprintf(w, "### %s\n\n", title)
		if withBefore {
			fmt.Fprintln(w, "| Agent | Target | Before | After |\n|---|---|---|---|")
		} else {
			fmt.Fprintln(w, "| Agent | Target | Access |\n|---|---|---|")
		}
		for _, c := range changes {
			target := fmt.Sprintf("%s `%s`", strings.ToLower(string(c.NodeType)), mdCell(c.NodeName))
			if withBefore {
				fmt.Fprintf(w, "| %s | %s | %s | %s |\n", c.AgentID, target, permLabel(c.Before), permLabel(c.After))
				continue
			}
			access := c.After
			if access == graph.PermNone {
				access = c.Before
			}
			fmt.Fprintf(w, "| %s | %s | %s |\n", c.AgentID, target, permLabel(access))
		}
		fmt.Fprintln(w)
	}
	table("Escalated", d.Escalations, true)
	table("Newly reachable", d.NewlyReachable, false)

	if len(d.NewFindings) > 0 {
		fmt.Fprintln(w, "### New findings\n\n| Severity | Rule | Finding |\n|---|---|---|")
		for _, f := range d.NewFindings {
			fmt.Fprintf(w, "| %s | %s | %s |\n", f.Severity, f.RuleID, mdCell(f.Message))
		}
		fmt.Fprintln(w)
	}
	if len(d.NoLongerReachable) > 0 {
		fmt.Fprintln(w, "<details><summary>No longer reachable</summary>")
		fmt.Fprintln(w)
		table("Removed", d.NoLongerReachable, false)
		fmt.Fprintln(w, "</details>")
	}
}
