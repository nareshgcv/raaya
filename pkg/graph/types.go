package graph

import "time"

// NodeType represents the entity level in the security graph.
type NodeType string

const (
	NodeAgent     NodeType = "AGENT"
	NodeMCPServer NodeType = "MCP_SERVER"
	NodeTool      NodeType = "TOOL"
	NodeResource  NodeType = "RESOURCE"
)

// PermissionLevel defines access rights for tools and resources.
type PermissionLevel string

const (
	PermRead      PermissionLevel = "READ"
	PermWrite     PermissionLevel = "WRITE"
	PermExecute   PermissionLevel = "EXECUTE"
	PermAdmin     PermissionLevel = "ADMIN"
)

// Node represents a single element in the MCP security topology.
type Node struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        NodeType          `json:"type"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Permissions []PermissionLevel `json:"permissions,omitempty"`
}

// Edge represents reachability or capability invocation between nodes.
type Edge struct {
	SourceID    string          `json:"source_id"`
	TargetID    string          `json:"target_id"`
	Capability  string          `json:"capability"`
	Permission  PermissionLevel `json:"permission"`
	Transitive  bool            `json:"transitive"`
}

// Graph holds the complete directed security reachability model.
type Graph struct {
	Nodes map[string]*Node `json:"nodes"`
	Edges []Edge           `json:"edges"`
}

// NewGraph initializes an empty security graph.
func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
		Edges: make([]Edge, 0),
	}
}

func (g *Graph) AddNode(node *Node) {
	g.Nodes[node.ID] = node
}

func (g *Graph) AddEdge(edge Edge) {
	g.Edges = append(g.Edges, edge)
}
