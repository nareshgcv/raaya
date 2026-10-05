package graph

import (
	"encoding/json"
	"fmt"
	"io"
)

// WriteJSON writes the graph as indented JSON with edges in sorted order.
func (g *Graph) WriteJSON(w io.Writer) error {
	snapshot := Graph{Nodes: g.Nodes, Edges: g.SortedEdges()}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(snapshot)
}

// ReadJSON loads a graph written by WriteJSON (for example a saved baseline).
func ReadJSON(r io.Reader) (*Graph, error) {
	g := NewGraph()
	if err := json.NewDecoder(r).Decode(g); err != nil {
		return nil, fmt.Errorf("read graph: %w", err)
	}
	if g.Nodes == nil {
		g.Nodes = map[string]*Node{}
	}
	for id, n := range g.Nodes {
		if n == nil || n.ID != id {
			return nil, fmt.Errorf("read graph: node key %q does not match its id", id)
		}
		if n.Metadata == nil {
			n.Metadata = map[string]string{}
		}
	}
	g.Normalize()
	return g, nil
}
