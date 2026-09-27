package blastradius

import (
	"fmt"
	"sort"

	"raaya/pkg/graph"
)

// BlastRadiusReport summarizes downstream capabilities reachable from a root asset
type BlastRadiusReport struct {
	RootNodeID         string            `json:"root_node_id"`
	RootNodeName       string            `json:"root_node_name"`
	TotalReachableNodes int              `json:"total_reachable_nodes"`
	ReachableTools     []string          `json:"reachable_tools"`
	ReachableSecrets   []string          `json:"reachable_secrets"`
	MaxDepth           int               `json:"max_depth"`
	Paths              map[string][]string `json:"paths"` // Target ID -> Path trace array
}

// Calculate computes the transitive reachability and blast radius starting from a specific node
func Calculate(sg *graph.SecurityGraph, startNodeID string) (*BlastRadiusReport, error) {
	rootNode, exists := sg.Nodes[startNodeID]
	if !exists {
		return nil, fmt.Errorf("node with ID '%s' not found in Security Graph", startNodeID)
	}

	report := &BlastRadiusReport{
		RootNodeID:       rootNode.ID,
		RootNodeName:     rootNode.Name,
		ReachableTools:   make([]string, 0),
		ReachableSecrets: make([]string, 0),
		Paths:            make(map[string][]string),
	}

	// Build adjacency list for fast forward-traversal lookup
	adj := make(map[string][]string)
	for _, edge := range sg.Edges {
		adj[edge.FromID] = append(adj[edge.FromID], edge.ToID)
	}

	// BFS Traversal tracking distances and execution paths
	visited := make(map[string]bool)
	depthMap := make(map[string]int)
	queue := []string{startNodeID}

	visited[startNodeID] = true
	depthMap[startNodeID] = 0
	report.Paths[startNodeID] = []string{startNodeID}

	maxDepth := 0

	for len(queue) > 0 {
		currID := queue[0]
		queue = queue[1:]

		currDepth := depthMap[currID]
		if currDepth > maxDepth {
			maxDepth = currDepth
		}

		for _, neighborID := range adj[currID] {
			if !visited[neighborID] {
				visited[neighborID] = true
				depthMap[neighborID] = currDepth + 1

				// Record execution trace path
				currentPath := append([]string{}, report.Paths[currID]...)
				report.Paths[neighborID] = append(currentPath, neighborID)

				// Categorize reachable asset capabilities
				if targetNode, ok := sg.Nodes[neighborID]; ok {
					switch targetNode.Kind {
					case graph.KindToolDef:
						report.ReachableTools = append(report.ReachableTools, targetNode.Name)
					case graph.KindSecret:
						report.ReachableSecrets = append(report.ReachableSecrets, targetNode.Name)
					}
				}

				queue = append(queue, neighborID)
			}
		}
	}

	report.TotalReachableNodes = len(visited) - 1 // Exclude root node itself
	report.MaxDepth = maxDepth

	sort.Strings(report.ReachableTools)
	sort.Strings(report.ReachableSecrets)

	return report, nil
}

// CalculateAll returns a blast radius analysis for every root Agent or MCP Server in the graph
func CalculateAll(sg *graph.SecurityGraph) []*BlastRadiusReport {
	var reports []*BlastRadiusReport

	for _, node := range sg.Nodes {
		if node.Kind == graph.KindAgentPrompt || node.Kind == graph.KindMCPServer {
			if report, err := Calculate(sg, node.ID); err == nil {
				reports = append(reports, report)
			}
		}
	}

	return reports
}
