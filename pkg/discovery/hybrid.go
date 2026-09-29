package discovery

import (
	"context"
	"fmt"

	"raaya/pkg/graph"
)

// HybridAnalyzer merges AST, MCP Schema, and live configuration scanning.
type HybridAnalyzer struct {
	mcpScanner    *MCPScanner
	promptScanner *PromptScanner
	annotationScanner *AnnotationScanner
}

func NewHybridAnalyzer() *HybridAnalyzer {
	return &HybridAnalyzer{
		mcpScanner:        NewMCPScanner(),
		promptScanner:     NewPromptScanner(),
		annotationScanner: NewAnnotationScanner(),
	}
}

// DiscoverGraph parses source code, MCP configs, and AST tags to construct the initial security graph.
func (h *HybridAnalyzer) DiscoverGraph(ctx context.Context, projectRoot string) (*graph.Graph, error) {
	g := graph.NewGraph()

	// 1. Scan MCP Server and Tool definitions
	mcpNodes, mcpEdges, err := h.mcpScanner.Scan(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("mcp scan failed: %w", err)
	}

	for _, node := range mcpNodes {
		g.AddNode(node)
	}
	for _, edge := range mcpEdges {
		g.AddEdge(edge)
	}

	// 2. Scan Code AST for Prompt injection risks & tool invocation paths
	promptNodes, promptEdges, err := h.promptScanner.ScanAST(projectRoot)
	if err == nil {
		for _, node := range promptNodes {
			g.AddNode(node)
		}
		for _, edge := range promptEdges {
			g.AddEdge(edge)
		}
	}

	return g, nil
}
