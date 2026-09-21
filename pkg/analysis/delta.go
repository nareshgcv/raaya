package analysis

import "github.com/raaya/pkg/graph"

type GraphDelta struct {
	AddedNodes   []*graph.Node `json:"added_nodes"`
	RemovedNodes []*graph.Node `json:"removed_nodes"`
	AddedEdges   []*graph.Edge `json:"added_edges"`
	RemovedEdges []*graph.Edge `json:"removed_edges"`
}

func ComputeDelta(base, head *graph.SecurityGraph) *GraphDelta {
	delta := &GraphDelta{
		AddedNodes:   make([]*graph.Node, 0),
		RemovedNodes: make([]*graph.Node, 0),
		AddedEdges:   make([]*graph.Edge, 0),
		RemovedEdges: make([]*graph.Edge, 0),
	}

	for id, node := range head.Nodes {
		if _, exists := base.Nodes[id]; !exists {
			delta.AddedNodes = append(delta.AddedNodes, node)
		}
	}

	for id, node := range base.Nodes {
		if _, exists := head.Nodes[id]; !exists {
			delta.RemovedNodes = append(delta.RemovedNodes, node)
		}
	}

	baseEdgeMap := make(map[string]bool)
	for _, edge := range base.Edges {
		baseEdgeMap[edge.ID] = true
	}

	for _, edge := range head.Edges {
		if !baseEdgeMap[edge.ID] {
			delta.AddedEdges = append(delta.AddedEdges, edge)
		}
	}

	return delta
}
