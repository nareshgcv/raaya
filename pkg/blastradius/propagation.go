package blastradius

import (
	"raaya/pkg/graph"
)

// ReachableImpact contains the computed blast radius for a given starting node.
type ReachableImpact struct {
	SourceID        string                 `json:"source_id"`
	DirectNodes     []*graph.Node          `json:"direct_nodes"`
	TransitiveNodes []*graph.Node          `json:"transitive_nodes"`
	MaxPermissions  []graph.PermissionLevel `json:"max_permissions"`
	Score           float64                `json:"impact_score"`
}

// CapabilityEngine handles transitive reachability propagation.
type CapabilityEngine struct {
	graph *graph.Graph
}

func NewCapabilityEngine(g *graph.Graph) *CapabilityEngine {
	return &CapabilityEngine{graph: g}
}

// CalculateBlastRadius performs BFS traversal to determine all downstream exposed entities.
func (e *CapabilityEngine) CalculateBlastRadius(startNodeID string) *ReachableImpact {
	impact := &ReachableImpact{
		SourceID:        startNodeID,
		DirectNodes:     make([]*graph.Node, 0),
		TransitiveNodes: make([]*graph.Node, 0),
		MaxPermissions:  make([]graph.PermissionLevel, 0),
	}

	visited := make(map[string]bool)
	queue := []string{startNodeID}
	visited[startNodeID] = true
	depth := make(map[string]int)
	depth[startNodeID] = 0

	// Adjacency list setup
	adj := make(map[string][]graph.Edge)
	for _, edge := range e.graph.Edges {
		adj[edge.SourceID] = append(adj[edge.SourceID], edge)
	}

	permSet := make(map[graph.PermissionLevel]bool)

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		currDepth := depth[curr]

		for _, edge := range adj[curr] {
			targetNode, exists := e.graph.Nodes[edge.TargetID]
			if !exists {
				continue
			}

			permSet[edge.Permission] = true

			if !visited[edge.TargetID] {
				visited[edge.TargetID] = true
				depth[edge.TargetID] = currDepth + 1
				queue = append(queue, edge.TargetID)

				if currDepth == 0 {
					impact.DirectNodes = append(impact.DirectNodes, targetNode)
				} else {
					impact.TransitiveNodes = append(impact.TransitiveNodes, targetNode)
				}
			}
		}
	}

	for perm := range permSet {
		impact.MaxPermissions = append(impact.MaxPermissions, perm)
	}

	// Calculate weighted blast radius score based on node types and reachable counts
	impact.Score = float64(len(impact.DirectNodes)*2 + len(impact.TransitiveNodes)*5)
	return impact
}
