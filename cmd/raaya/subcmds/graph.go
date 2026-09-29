package main

import (
	"context"
	"fmt"

	"raaya/pkg/analysis"
	"raaya/pkg/graph"
	"raaya/pkg/reporter"
)

func runGraphDiffScan() {
	ctx := context.Background()

	// 1. Base Graph (simulated HEAD~1 scan)
	baseGraph := graph.NewSecurityGraph()
	baseGraph.AddNode(graph.Node{ID: "agent:support", Type: graph.NodeAgent, Name: "SupportAgent"})
	baseGraph.AddNode(graph.Node{ID: "mcp:db", Type: graph.NodeMCPServer, Name: "DBServer"})
	baseGraph.AddEdge("agent:support", "mcp:db")

	// 2. Current Graph (working tree scan with expanded capabilities)
	currentGraph := graph.NewSecurityGraph()
	currentGraph.AddNode(graph.Node{ID: "agent:support", Type: graph.NodeAgent, Name: "SupportAgent"})
	currentGraph.AddNode(graph.Node{ID: "mcp:db", Type: graph.NodeMCPServer, Name: "DBServer"})
	currentGraph.AddNode(graph.Node{ID: "tool:write_sql", Type: graph.NodeTool, Name: "WriteSQL", Capabilities: []graph.Capability{graph.CapWriteDatabase}})

	currentGraph.AddEdge("agent:support", "mcp:db")
	currentGraph.AddEdge("mcp:db", "tool:write_sql")

	// 3. Propagate capabilities transitively
	currentGraph.PropagateTransitiveCapabilities()

	// 4. Differential change detection
	diff := analysis.CompareGraphs(baseGraph, currentGraph)

	// 5. Evaluate Rego rules
	evaluator := analysis.NewRegoEvaluator("")
	violations, err := evaluator.Evaluate(ctx, currentGraph, diff)
	if err != nil {
		panic(err)
	}

	// 6. Generate PR Comment Markdown
	prReporter := reporter.NewPRCommentReporter()
	output := prReporter.GenerateMarkdown(diff, violations, currentGraph)

	fmt.Println(output)
}
