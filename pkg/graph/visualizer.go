package graph

import (
	"fmt"
	"strings"
)

// RenderMermaid exports the security graph as a Mermaid flowchart diagram.
func RenderMermaid(g *Graph) string {
	var sb strings.Builder
	sb.WriteString("flowchart TD\n")

	// Group/Style nodes by type
	for _, node := range g.Nodes {
		switch node.Type {
		case NodeAgent:
			sb.WriteString(fmt.Sprintf("    %s([🤖 Agent: %s])\n", node.ID, node.Name))
		case NodeMCPServer:
			sb.WriteString(fmt.Sprintf("    %s[🖥️ MCP Server: %s]\n", node.ID, node.Name))
		case NodeTool:
			sb.WriteString(fmt.Sprintf("    %s{{🛠️ Tool: %s}}\n", node.ID, node.Name))
		case NodeResource:
			sb.WriteString(fmt.Sprintf("    %s[(🗄️ Resource: %s)]\n", node.ID, node.Name))
		}
	}

	// Render reachability edges
	for _, edge := range g.Edges {
		label := string(edge.Permission)
		if edge.Capability != "" {
			label = fmt.Sprintf("%s: %s", edge.Capability, edge.Permission)
		}
		sb.WriteString(fmt.Sprintf("    %s -->|%s| %s\n", edge.SourceID, label, edge.TargetID))
	}

	return sb.String()
}

// RenderDOT exports the graph to DOT format for Graphviz visualization.
func RenderDOT(g *Graph) string {
	var sb strings.Builder
	sb.WriteString("digraph RaayaSecurityGraph {\n")
	sb.WriteString("  rankdir=LR;\n")
	sb.WriteString("  node [shape=box, style=rounded, fontname=\"Helvetica\"];\n\n")

	for _, node := range g.Nodes {
		sb.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\\n(%s)\"];\n", node.ID, node.Name, node.Type))
	}

	for _, edge := range g.Edges {
		sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"];\n", edge.SourceID, edge.TargetID, edge.Permission))
	}

	sb.WriteString("}\n")
	return sb.String()
}
