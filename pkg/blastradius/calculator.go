package blastradius

import "github.com/raaya/pkg/graph"

type RadiusMap map[string][]string

func CalculateTransitiveReach(sg *graph.SecurityGraph, startNodeID string) []string {
	visited := make(map[string]bool)
	reachable := make([]string, 0)

	var dfs func(currID string)
	dfs = func(currID string) {
		if visited[currID] {
			return
		}
		visited[currID] = true

		for _, edge := range sg.Edges {
			if edge.SourceID == currID {
				if !visited[edge.TargetID] {
					reachable = append(reachable, edge.TargetID)
					dfs(edge.TargetID)
				}
			}
		}
	}

	dfs(startNodeID)
	return reachable
}
