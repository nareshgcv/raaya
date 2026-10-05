package blastradius

import "github.com/nareshgcv/raaya/pkg/graph"

// Radius is everything a compromised node could reach.
type Radius struct {
	SourceID   string  `json:"source_id"`
	Direct     []Entry `json:"direct"`
	Transitive []Entry `json:"transitive"`
	Score      int     `json:"score"`
}

var permissionWeight = map[graph.PermissionLevel]int{
	graph.PermNone: 1, graph.PermRead: 1, graph.PermWrite: 3, graph.PermExecute: 5, graph.PermAdmin: 8,
}

// Calculate treats nodeID as fully compromised and reports everything
// downstream of it. The score weights each node by its permission and counts
// resources double, so a few ADMIN resources outscore many READ tools.
func Calculate(g *graph.Graph, nodeID string) Radius {
	direct := map[string]bool{}
	for _, e := range g.Edges {
		if e.SourceID == nodeID {
			direct[e.TargetID] = true
		}
	}
	r := Radius{SourceID: nodeID, Direct: []Entry{}, Transitive: []Entry{}}
	for _, entry := range From(g, nodeID).Sorted() {
		if direct[entry.NodeID] {
			r.Direct = append(r.Direct, entry)
		} else {
			r.Transitive = append(r.Transitive, entry)
		}
		weight := permissionWeight[entry.Permission]
		if g.Nodes[entry.NodeID].Type == graph.NodeResource {
			weight *= 2
		}
		r.Score += weight
	}
	return r
}
