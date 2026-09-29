package blastradius

import "raaya/pkg/graph"

type ReachabilityPath struct {
	AgentID      string             `json:"agent_id"`
	Path         []string           `json:"path"`
	TargetID     string             `json:"target_id"`
	TargetType   graph.NodeType     `json:"target_type"`
	Capabilities []graph.Capability `json:"capabilities"`
}

type ReachabilityEngine struct {
	Graph *graph.SecurityGraph
}

func NewReachabilityEngine(g *graph.SecurityGraph) *ReachabilityEngine {
	return &ReachabilityEngine{Graph: g}
}

// ComputePaths computes reachability across all agent boundaries
func (re *ReachabilityEngine) ComputePaths() []ReachabilityPath {
	adj := make(map[string][]string)
	for _, e := range re.Graph.Edges {
		adj[e.From] = append(adj[e.From], e.To)
	}

	var results []ReachabilityPath
	for id, n := range re.Graph.Nodes {
		if n.Type == graph.NodeAgent {
			visited := make(map[string]bool)
			re.dfs(id, id, []string{id}, adj, visited, &results)
		}
	}
	return results
}

func (re *ReachabilityEngine) dfs(agentID, curr string, path []string, adj map[string][]string, visited map[string]bool, results *[]ReachabilityPath) {
	visited[curr] = true
	defer func() { visited[curr] = false }()

	for _, nxt := range adj[curr] {
		if visited[nxt] { continue }
		nextPath := append([]string{}, path...)
		nextPath = append(nextPath, nxt)
		targetNode := re.Graph.Nodes[nxt]

		if len(targetNode.Capabilities) > 0 || targetNode.Type == graph.NodeResource {
			*results = append(*results, ReachabilityPath{
				AgentID:      agentID,
				Path:         nextPath,
				TargetID:     targetNode.ID,
				TargetType:   targetNode.Type,
				Capabilities: targetNode.Capabilities,
			})
		}
		re.dfs(agentID, nxt, nextPath, adj, visited, results)
	}
}
