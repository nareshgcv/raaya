package subcmds

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"raaya/pkg/analysis"
	"raaya/pkg/discovery"
	"raaya/pkg/reporter"
)

var CheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Scan repository for AI agent and MCP security violations",
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("path")
		if repoPath == "" {
			repoPath = "."
		}

		// 1. Hybrid AST & Config Discovery
		scanner := discovery.NewHybridScanner(repoPath)
		sg, err := scanner.BuildGraphFromASTAndConfig("mcp.json")
		if err != nil {
			fmt.Printf("Warning: Failed to parse MCP configs: %v\n", err)
		}

		// Also scan .cursor/mcp.json if present
		if cursorGraph, err := scanner.BuildGraphFromASTAndConfig(".cursor/mcp.json"); err == nil {
			for id, node := range cursorGraph.Nodes {
				sg.AddNode(node)
			}
			for _, edge := range cursorGraph.Edges {
				sg.AddEdge(edge.From, edge.To)
			}
		}

		// 2. Propagate Transitive Capabilities (Tools/Resources -> MCPServers -> Agents)
		sg.PropagateTransitiveCapabilities()

		// 3. Evaluate Policy Rules via Dual Engine & Rego Evaluator
		evaluator := analysis.NewRegoEvaluator("")
		violations, err := evaluator.Evaluate(context.Background(), sg, nil)
		if err != nil {
			fmt.Printf("Error during policy evaluation: %v\n", err)
			os.Exit(1)
		}

		// 4. Render Terminal Report
		reporter.PrintTerminalReport(violations)

		// Exit code 1 if critical policy violations exist
		if len(violations) > 0 {
			os.Exit(1)
		}
	},
}

func init() {
	CheckCmd.Flags().StringP("path", "p", ".", "Path to the repository root")
}
