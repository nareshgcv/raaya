package analysis

import (
	"sort"

	"github.com/nareshgcv/raaya/pkg/blastradius"
	"github.com/nareshgcv/raaya/pkg/graph"
)

// ReachChange is a change in what one agent can reach.
type ReachChange struct {
	AgentID  string                `json:"agent_id"`
	NodeID   string                `json:"node_id"`
	NodeName string                `json:"node_name"`
	NodeType graph.NodeType        `json:"node_type"`
	Before   graph.PermissionLevel `json:"before,omitempty"`
	After    graph.PermissionLevel `json:"after,omitempty"`
}

// Diff is the security-surface change between two graphs.
type Diff struct {
	NewlyReachable    []ReachChange `json:"newly_reachable"`
	Escalations       []ReachChange `json:"escalations"`
	NoLongerReachable []ReachChange `json:"no_longer_reachable"`
	NewFindings       []Finding     `json:"new_findings"`
	AddedNodes        []*graph.Node `json:"added_nodes"`
	RemovedNodes      []*graph.Node `json:"removed_nodes"`
	AddedEdges        []graph.Edge  `json:"added_edges"`
	RemovedEdges      []graph.Edge  `json:"removed_edges"`
	Regressions       int           `json:"regressions"`
}

// ComputeDiff compares the effective reach of every agent in base and head,
// so a change three hops away (a new tool on an existing server, a read-only
// tool losing its annotation) is reported against the agents it affects.
//
// Regressions are: newly reachable nodes at or above threshold, any
// escalation, and new HIGH findings not already covered by reach changes.
func ComputeDiff(base, head *graph.Graph, threshold graph.PermissionLevel) *Diff {
	d := &Diff{
		NewlyReachable: []ReachChange{}, Escalations: []ReachChange{}, NoLongerReachable: []ReachChange{},
		NewFindings: []Finding{}, AddedNodes: []*graph.Node{}, RemovedNodes: []*graph.Node{},
		AddedEdges: []graph.Edge{}, RemovedEdges: []graph.Edge{},
	}

	for _, n := range head.SortedNodes() {
		if _, ok := base.Nodes[n.ID]; !ok {
			d.AddedNodes = append(d.AddedNodes, n)
		}
	}
	for _, n := range base.SortedNodes() {
		if _, ok := head.Nodes[n.ID]; !ok {
			d.RemovedNodes = append(d.RemovedNodes, n)
		}
	}
	baseEdges, headEdges := edgeSet(base), edgeSet(head)
	for _, e := range head.SortedEdges() {
		if !baseEdges[e.Key()] {
			d.AddedEdges = append(d.AddedEdges, e)
		}
	}
	for _, e := range base.SortedEdges() {
		if !headEdges[e.Key()] {
			d.RemovedEdges = append(d.RemovedEdges, e)
		}
	}

	for _, agentID := range agentIDs(base, head) {
		before, after := reachOf(base, agentID), reachOf(head, agentID)
		for _, r := range after.Sorted() {
			n := head.Nodes[r.NodeID]
			prev, had := before[r.NodeID]
			change := ReachChange{AgentID: agentID, NodeID: n.ID, NodeName: n.Name, NodeType: n.Type, Before: prev, After: r.Permission}
			switch {
			case !had:
				d.NewlyReachable = append(d.NewlyReachable, change)
				if r.Permission != graph.PermNone && r.Permission.Rank() >= threshold.Rank() {
					d.Regressions++
				}
			case r.Permission.Rank() > prev.Rank():
				d.Escalations = append(d.Escalations, change)
				d.Regressions++
			}
		}
		for _, r := range before.Sorted() {
			if _, still := after[r.NodeID]; !still {
				n := base.Nodes[r.NodeID]
				d.NoLongerReachable = append(d.NoLongerReachable, ReachChange{
					AgentID: agentID, NodeID: n.ID, NodeName: n.Name, NodeType: n.Type, Before: r.Permission,
				})
			}
		}
	}

	}
	return d
}

func reachOf(g *graph.Graph, agentID string) blastradius.Reach {
	if _, ok := g.Nodes[agentID]; !ok {
		return blastradius.Reach{}
	}
	return blastradius.From(g, agentID)
}

func agentIDs(graphs ...*graph.Graph) []string {
	set := map[string]bool{}
	for _, g := range graphs {
		for _, n := range g.NodesOfType(graph.NodeAgent) {
			set[n.ID] = true
		}
	}
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func edgeSet(g *graph.Graph) map[string]bool {
	set := make(map[string]bool, len(g.Edges))
	for _, e := range g.Edges {
		set[e.Key()] = true
	}
	return set
}
