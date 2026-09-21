package graph

import (
	"fmt"
	"strings"
)

func ExportToDOT(sg *SecurityGraph) string {
	var sb strings.Builder
	sb.WriteString("digraph SecurityGraph {\n")
	sb.WriteString("  rankdir=LR;\n")
	sb.WriteString("  node [shape=box, style=filled, color=lightgrey];\n\n")

	for _, node := range sg.Nodes {
		sb.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\\n(%s)\"];\n", node.ID, node.Name, node.Type))
	}

	sb.WriteString("\n")
	for _, edge := range sg.Edges {
		sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"];\n", edge.SourceID, edge.TargetID, edge.Type))
	}

	sb.WriteString("}\n")
	return sb.String()
}
