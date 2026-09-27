package graph

import (
	"fmt"
	"strings"
)

func (g *SecurityGraph) ToMermaid() string {
	var builder strings.Builder
	builder.WriteString("graph TD\n")

	for _, node := range g.Nodes {
		label := fmt.Sprintf("%s\\n(%s)", node.Name, node.Kind)
		switch node.Kind {
		case KindMCPServer:
			builder.WriteString(fmt.Sprintf("  %s[\"%s\"]:::%s\n", sanitizeID(node.ID), label, "server"))
		case KindToolDef:
			builder.WriteString(fmt.Sprintf("  %s(\"%s\"):::%s\n", sanitizeID(node.ID), label, "tool"))
		case KindSecret:
			builder.WriteString(fmt.Sprintf("  %s{{\"%s\"}}:::%s\n", sanitizeID(node.ID), label, "secret"))
		default:
			builder.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", sanitizeID(node.ID), label))
		}
	}

	for _, edge := range g.Edges {
		builder.WriteString(fmt.Sprintf("  %s -->|%s| %s\n", sanitizeID(edge.FromID), edge.Relation, sanitizeID(edge.ToID)))
	}

	builder.WriteString("\nclassDef server fill:#f9f,stroke:#333,stroke-width:2px;\n")
	builder.WriteString("classDef tool fill:#bbf,stroke:#333,stroke-width:1px;\n")
	builder.WriteString("classDef secret fill:#f88,stroke:#333,stroke-width:2px;\n")

	return builder.String()
}

func sanitizeID(id string) string {
	r := strings.NewReplacer(":", "_", "/", "_", ".", "_", "-", "_")
	return r.Replace(id)
}
