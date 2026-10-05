package graph

import "sort"

// NewGraph returns an empty graph.
func NewGraph() *Graph {
	return &Graph{Nodes: map[string]*Node{}, Edges: []Edge{}}
}

// AddNode inserts n, or merges it into an existing node with the same ID
// (the existing name and metadata values win). It returns the stored node.
func (g *Graph) AddNode(n *Node) *Node {
	if n.Metadata == nil {
		n.Metadata = map[string]string{}
	}
	existing, ok := g.Nodes[n.ID]
	if !ok {
		g.Nodes[n.ID] = n
		return n
	}
	if existing.Name == "" {
		existing.Name = n.Name
	}
	if existing.Type == "" {
		existing.Type = n.Type
	}
	if existing.Metadata == nil {
		existing.Metadata = map[string]string{}
	}
	for k, v := range n.Metadata {
		if _, set := existing.Metadata[k]; !set {
			existing.Metadata[k] = v
		}
	}
	return existing
}

// AddEdge appends e unless an identical edge exists, and reports whether it did.
func (g *Graph) AddEdge(e Edge) bool {
	if g.index == nil {
		g.rebuildIndex()
	}
	k := e.Key()
	if _, dup := g.index[k]; dup {
		return false
	}
	g.index[k] = struct{}{}
	g.Edges = append(g.Edges, e)
	return true
}

// SetPermission sets the permission of every edge matching match and returns
// how many edges matched. Edges that become identical are merged.
func (g *Graph) SetPermission(match func(Edge) bool, p PermissionLevel) int {
	n := 0
	for i := range g.Edges {
		if match(g.Edges[i]) {
			g.Edges[i].Permission = p
			n++
		}
	}
	if n > 0 {
		g.rebuildIndex()
	}
	return n
}

// Normalize sorts edges so that output is deterministic.
func (g *Graph) Normalize() {
	sort.Slice(g.Edges, func(i, j int) bool { return edgeLess(g.Edges[i], g.Edges[j]) })
	g.rebuildIndex()
}

// rebuildIndex rebuilds the duplicate index, dropping duplicate edges.
func (g *Graph) rebuildIndex() {
	g.index = make(map[string]struct{}, len(g.Edges))
	kept := g.Edges[:0]
	for _, e := range g.Edges {
		k := e.Key()
		if _, dup := g.index[k]; dup {
			continue
		}
		g.index[k] = struct{}{}
		kept = append(kept, e)
	}
	g.Edges = kept
}
