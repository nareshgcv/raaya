package analysis

import (
	"raaya/pkg/graph"
)

// CapabilityDiff captures security posture shifts between commits or PRs.
type CapabilityDiff struct {
	AddedNodes     []*graph.Node `json:"added_nodes"`
	RemovedNodes   []*graph.Node `json:"removed_nodes"`
	AddedEdges     []graph.Edge  `json:"added_edges"`
	Escalations    []graph.Edge  `json:"escalations"`
	HasRegressions bool          `json:"has_regressions"`
}

// ComputeDiff compares base vs head security graphs.
func ComputeDiff(base, head *graph.Graph) *CapabilityDiff {
	diff := &CapabilityDiff{
		AddedNodes:   make([]*graph.Node, 0),
		RemovedNodes: make([]*graph.Node, 0),
		AddedEdges:   make([]graph.Edge, 0),
		Escalations:  make([]graph.Edge, 0),
	}

	// Detect Node additions
	for id, headNode := range head.Nodes {
		if _, exists := base.Nodes[id]; !exists {
			diff.AddedNodes = append(diff.AddedNodes, headNode)
		}
	}

	// Detect Node removals
	for id, baseNode := range base.Nodes {
		if _, exists := head.Nodes[id]; !exists {
			diff.RemovedNodes = append(diff.RemovedNodes, baseNode)
		}
	}

	// Map base edges for O(1) comparison
	baseEdgeMap := make(map[string]graph.Edge)
	for _, edge := range base.Edges {
		key := edge.SourceID + "->" + edge.TargetID
		baseEdgeMap[key] = edge
	}

	// Compare Head edges against Base
	for _, headEdge := range head.Edges {
		key := headEdge.SourceID + "->" + headEdge.TargetID
		if baseEdge, exists := baseEdgeMap[key]; !exists {
			diff.AddedEdges = append(diff.AddedEdges, headEdge)
			diff.HasRegressions = true
		} else if isEscalation(baseEdge.Permission, headEdge.Permission) {
			diff.Escalations = append(diff.Escalations, headEdge)
			diff.HasRegressions = true
		}
	}

	return diff
}

func isEscalation(basePerm, headPerm graph.PermissionLevel) bool {
	weights := map[graph.PermissionLevel]int{
		graph.PermRead:    1,
		graph.PermExecute: 2,
		graph.PermWrite:   3,
		graph.PermAdmin:   4,
	}
	return weights[headPerm] > weights[basePerm]
}
