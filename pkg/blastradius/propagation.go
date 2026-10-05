// Package blastradius propagates capabilities through the security graph.
package blastradius

import (
	"sort"

	"github.com/nareshgcv/raaya/pkg/graph"
)

// Reach maps every node reachable from a start node to the strongest
// permission obtainable on it.
//
// Along one path the effective permission is the weakest permissioned edge
// (a read-only tool caps everything behind it); structural edges limit
// nothing. Across paths the strongest result wins. Nodes reachable only
// through structural edges map to graph.PermNone.
type Reach map[string]graph.PermissionLevel

// Entry is one reachable node and its effective permission.
type Entry struct {
	NodeID     string                `json:"node_id"`
	Permission graph.PermissionLevel `json:"permission,omitempty"`
}

const unbounded = 1 << 20

// From computes the Reach of startID. It terminates on cyclic graphs because
// a node is revisited only when its permission strictly improves, and there
// are finitely many levels.
func From(g *graph.Graph, startID string) Reach {
	adj := make(map[string][]graph.Edge)
	for _, e := range g.Edges {
		adj[e.SourceID] = append(adj[e.SourceID], e)
	}

	best := map[string]int{startID: unbounded}
	queue := []string{startID}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, e := range adj[cur] {
			if _, ok := g.Nodes[e.TargetID]; !ok {
				continue
			}
			limit := best[cur]
			if r := e.Permission.Rank(); r > 0 && r < limit {
				limit = r
			}
			if prev, seen := best[e.TargetID]; !seen || limit > prev {
				best[e.TargetID] = limit
				queue = append(queue, e.TargetID)
			}
		}
	}

	delete(best, startID)
	out := make(Reach, len(best))
	for id, r := range best {
		out[id] = rankToPermission(r)
	}
	return out
}

// Sorted returns the entries ordered by node ID.
func (r Reach) Sorted() []Entry {
	out := make([]Entry, 0, len(r))
	for id, p := range r {
		out = append(out, Entry{NodeID: id, Permission: p})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].NodeID < out[j].NodeID })
	return out
}

func rankToPermission(r int) graph.PermissionLevel {
	switch r {
	case 1:
		return graph.PermRead
	case 2:
		return graph.PermWrite
	case 3:
		return graph.PermExecute
	case 4:
		return graph.PermAdmin
	default:
		return graph.PermNone
	}
}
