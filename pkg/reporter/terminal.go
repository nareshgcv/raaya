package reporter

import (
	"fmt"
	"strings"

	"github.com/raaya/pkg/graph"
)

func RenderTerminalTree(sg *graph.SecurityGraph) {
	fmt.Println("\n=== RAAYA Dependency Graph ===")

	for _, node := range sg.Nodes {
		if node.Type == graph.NodeAgent || node.Type == graph.NodeServer {
			fmt.Printf("● [%s] %s\n", node.Type, node.Name)
			renderOutgoingEdges(sg, node.ID, "  ")
		}
	}
	fmt.Println()
}

func renderOutgoingEdges(sg *graph.SecurityGraph, nodeID string, indent string) {
	for _, edge := range sg.Edges {
		if edge.SourceID == nodeID {
			target := sg.Nodes[edge.TargetID]
			if target != nil {
				fmt.Printf("%s└── (%s) ──> [%s] %s\n", indent, edge.Type, target.Type, target.Name)
				renderOutgoingEdges(sg, target.ID, indent+"    ")
			}
		}
	}
}
