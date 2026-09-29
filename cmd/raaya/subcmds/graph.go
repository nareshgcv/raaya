package subcmds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"raaya/pkg/blastradius"
	"raaya/pkg/discovery"
)

var (
	exportMermaid bool
	showPaths     bool
)

var GraphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Export or visualize the Agentic Security DAG and reachability paths",
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("path")
		if repoPath == "" {
			repoPath = "."
		}

		// 1. Build Hybrid Security Graph
		scanner := discovery.NewHybridScanner(repoPath)
		sg, err := scanner.BuildGraphFromASTAndConfig("mcp.json")
		if err != nil {
			fmt.Printf("Error building security graph: %v\n", err)
			os.Exit(1)
		}

		// 2. Transitive Capability Propagation
		sg.PropagateTransitiveCapabilities()

		// 3. Render Reachability Paths if requested
		if showPaths {
			calc := blastradius.NewBlastRadiusCalculator(sg)
			paths := calc.ComputeAgentReachability()

			fmt.Println("=== Agent Reachability Paths ===")
			for _, p := range paths {
				fmt.Printf("Agent [%s] -> Target [%s] (%s)\n  Path: %v\n  Capabilities: %v\n\n",
					p.AgentID, p.TargetID, p.TargetType, p.Path, p.Capabilities)
			}
			return
		}

		// 4. Default / Export Mermaid DAG
		fmt.Println(sg.ExportMermaid())
	},
}

func init() {
	GraphCmd.Flags().StringP("path", "p", ".", "Path to repository root")
	GraphCmd.Flags().BoolVarP(&exportMermaid, "mermaid", "m", true, "Output graph as Mermaid JS format")
	GraphCmd.Flags().BoolVarP(&showPaths, "paths", "r", false, "Show calculated reachability paths")
}
