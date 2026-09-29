package graph

import (
	"fmt"
	"strings"
)

type NodeType string

const (
	NodeAgent     NodeType = "AGENT"
	NodeMCPServer NodeType = "MCP_SERVER"
	NodeTool      NodeType = "TOOL"
	NodeResource  NodeType = "RESOURCE"
)

type Capability string

const (
	CapReadDatabase  Capability = "READ_DATABASE"
	CapWriteDatabase Capability = "WRITE_DATABASE"
	CapExecCommand   Capability = "EXEC_COMMAND"
	CapFileSystem    Capability = "FILE_SYSTEM"
	CapNetworkAccess Capability = "NETWORK_ACCESS"
)

type Node struct {
	ID           string       `json:"id"`
	Type         NodeType     `json:"type"`
	Name         string       `json:"name"`
	Capabilities []Capability `json:"capabilities"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type SecurityGraph struct {
	Nodes map[string]Node `json:"nodes"`
	Edges []Edge          `json:"edges"`
}

func NewSecurityGraph() *SecurityGraph {
	return &SecurityGraph{
		Nodes: make(map[string]Node),
		Edges: make([]Edge, 0),
	}
}

func (g *SecurityGraph) AddNode(node Node) {
	g.Nodes[node.ID] = node
}

func (g *SecurityGraph) AddEdge(from, to string) {
	for _, e := range g.Edges {
		if e.From == from && e.To == to {
			return
		}
	}
	g.Edges = append(g.Edges, Edge{From: from, To: to})
}

// PropagateTransitiveCapabilities bubbles downstream capabilities (Tools/Resources) up to Agent nodes
func (g *SecurityGraph) PropagateTransitiveCapabilities() {
	adj := make(map[string][]string)
	for _, edge := range g.Edges {
		adj[edge.From] = append(adj[edge.From], edge.To)
	}

	for _, node := range g.Nodes {
		if node.Type == NodeAgent {
			propagated := g.collectCapsDFS(node.ID, adj, make(map[string]bool))
			node.Capabilities = mergeCaps(node.Capabilities, propagated)
			g.Nodes[node.ID] = node
		}
	}
}

func (g *SecurityGraph) collectCapsDFS(currID string, adj map[string][]string, visited map[string]bool) []Capability {
	if visited[currID] {
		return nil
	}
	visited[currID] = true

	var caps []Capability
	targetNode := g.Nodes[currID]
	caps = append(caps, targetNode.Capabilities...)

	for _, neighborID := range adj[currID] {
		caps = append(caps, g.collectCapsDFS(neighborID, adj, visited)...)
	}

	return caps
}

func mergeCaps(a, b []Capability) []Capability {
	set := make(map[Capability]bool)
	for _, c := range a {
		set[c] = true
	}
	for _, c := range b {
		set[c] = true
	}
	var res []Capability
	for c := range set {
		res = append(res, c)
	}
	return res
}

// ExportMermaid exports DAG for PR markdown comments
func (g *SecurityGraph) ExportMermaid() string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")
	for _, node := range g.Nodes {
		sb.WriteString(fmt.Sprintf("  %s[\"%s (%s)\"]\n", node.ID, node.Name, node.Type))
	}
	for _, edge := range g.Edges {
		sb.WriteString(fmt.Sprintf("  %s --> %s\n", edge.From, edge.To))
	}
	return sb.String()
}
