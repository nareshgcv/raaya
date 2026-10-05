package rules

import (
	"fmt"

	"github.com/nareshgcv/raaya/pkg/blastradius"
	"github.com/nareshgcv/raaya/pkg/graph"
)

// RAAYA001 and RAAYA002: capabilities an agent holds transitively, through
// servers and tools, rather than through an explicit grant.
func reachableCapabilities(g *graph.Graph) []Finding {
	var out []Finding
	reaches := map[string]blastradius.Reach{}
	for _, agent := range g.NodesOfType(graph.NodeAgent) {
		reaches[agent.ID] = blastradius.From(g, agent.ID)
	}
	for _, agent := range g.NodesOfType(graph.NodeAgent) {
		parent := reaches[agent.Metadata[graph.MetaParentAgent]]
		for _, r := range reaches[agent.ID].Sorted() {
			// A subagent reaching what its parent already reaches adds noise, not risk.
			if p, ok := parent[r.NodeID]; ok && p.Rank() >= r.Permission.Rank() {
				continue
			}
			n := g.Nodes[r.NodeID]
			switch {
			case r.Permission.Rank() >= graph.PermExecute.Rank():
				out = append(out, newFinding(g, "RAAYA001", SevHigh, agent.ID, n.ID,
					fmt.Sprintf("%s can reach %s %q with %s%s", agent.Name, typeLabel(n.Type), n.Name, r.Permission, inferredNote(g, n.ID))))
			case r.Permission == graph.PermWrite && n.Type == graph.NodeResource:
				out = append(out, newFinding(g, "RAAYA002", SevMedium, agent.ID, n.ID,
					fmt.Sprintf("%s can modify resource %q%s", agent.Name, n.Name, inferredNote(g, n.ID))))
			}
		}
	}
	return out
}

// RAAYA005: a server handed the filesystem root or a home directory.
func broadFilesystemAccess(g *graph.Graph) []Finding {
	var out []Finding
	for _, e := range g.SortedEdges() {
		res, server := g.Nodes[e.TargetID], g.Nodes[e.SourceID]
		if res == nil || server == nil || res.Type != graph.NodeResource || res.Metadata[graph.MetaBroadPath] != "true" {
			continue
		}
		out = append(out, newFinding(g, "RAAYA005", SevMedium, "", server.ID,
			fmt.Sprintf("Server %q is given %q, which exposes the whole filesystem or home directory", server.Name, res.Name)))
	}
	return out
}

func inferredNote(g *graph.Graph, nodeID string) string {
	for _, e := range g.Edges {
		if e.TargetID == nodeID && e.Inferred {
			return " (inferred from server configuration)"
		}
	}
	return ""
}
