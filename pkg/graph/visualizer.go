package graph

import (
	"fmt"
	"strconv"
	"strings"
)

// Mermaid renders the graph as a Mermaid flowchart. Inferred edges are dashed.
func (g *Graph) Mermaid() string {
	var b strings.Builder
	b.WriteString("graph LR\n")
	ids := map[string]string{}
	for i, n := range g.SortedNodes() {
		id := fmt.Sprintf("n%d", i)
		ids[n.ID] = id
		label := strings.NewReplacer(`"`, "#quot;", "\n", " ").Replace(n.Name)
		switch n.Type {
		case NodeAgent:
			fmt.Fprintf(&b, "  %s([\"%s\"]):::agent\n", id, label)
		case NodeMCPServer:
			fmt.Fprintf(&b, "  %s[\"%s\"]:::server\n", id, label)
		case NodeTool:
			fmt.Fprintf(&b, "  %s(\"%s\"):::tool\n", id, label)
		default:
			fmt.Fprintf(&b, "  %s[(\"%s\")]:::resource\n", id, label)
		}
	}
	for _, e := range g.SortedEdges() {
		from, okFrom := ids[e.SourceID]
		to, okTo := ids[e.TargetID]
		if !okFrom || !okTo {
			continue
		}
		arrow := "-->"
		if e.Inferred {
			arrow = "-.->"
		}
		if e.Permission != PermNone {
			fmt.Fprintf(&b, "  %s %s|%s| %s\n", from, arrow, e.Permission, to)
		} else {
			fmt.Fprintf(&b, "  %s %s %s\n", from, arrow, to)
		}
	}
	b.WriteString("  classDef agent fill:#e0e7ff,stroke:#4338ca\n")
	b.WriteString("  classDef server fill:#fae8ff,stroke:#a21caf\n")
	b.WriteString("  classDef tool fill:#dbeafe,stroke:#1d4ed8\n")
	b.WriteString("  classDef resource fill:#fee2e2,stroke:#b91c1c\n")
	return b.String()
}

// DOT renders the graph in Graphviz format; render SVG with `dot -Tsvg`.
func (g *Graph) DOT() string {
	shapes := map[NodeType]string{NodeAgent: "oval", NodeMCPServer: "box", NodeTool: "component", NodeResource: "cylinder"}
	var b strings.Builder
	b.WriteString("digraph raaya {\n  rankdir=LR;\n")
	for _, n := range g.SortedNodes() {
		shape := shapes[n.Type]
		if shape == "" {
			shape = "ellipse"
		}
		fmt.Fprintf(&b, "  %s [label=%s, shape=%s];\n", strconv.Quote(n.ID), strconv.Quote(n.Name), shape)
	}
	for _, e := range g.SortedEdges() {
		if _, ok := g.Nodes[e.TargetID]; !ok {
			continue
		}
		var attrs []string
		if e.Permission != PermNone {
			attrs = append(attrs, "label="+strconv.Quote(string(e.Permission)))
		}
		if e.Inferred {
			attrs = append(attrs, "style=dashed")
		}
		fmt.Fprintf(&b, "  %s -> %s", strconv.Quote(e.SourceID), strconv.Quote(e.TargetID))
		if len(attrs) > 0 {
			fmt.Fprintf(&b, " [%s]", strings.Join(attrs, ", "))
		}
		b.WriteString(";\n")
	}
	b.WriteString("}\n")
	return b.String()
}
