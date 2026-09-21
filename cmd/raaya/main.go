package main

import (
	"context"
	"fmt"
	"os"

	"github.com/raaya/pkg/analysis"
	"github.com/raaya/pkg/discovery"
	"github.com/raaya/pkg/graph"
	"github.com/raaya/pkg/reporter"
	"github.com/spf13/cobra"
)

func main() {
	var configPath string
	var sarifOutput bool

	var rootCmd = &cobra.Command{
		Use:   "ape",
		Short: "ape is a local dependency graph analyzer for AI agents and MCP servers",
	}

	var scanCmd = &cobra.Command{
		Use:   "scan",
		Short: "Scan local repository for MCP configurations and agent dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			sg := graph.NewSecurityGraph()

			// 1. Scan codebase AST for @tool annotations
			if err := discovery.ScanDirectoryAnnotations(".", sg); err != nil {
				return fmt.Errorf("failed scanning code annotations: %w", err)
			}

			// 2. Parse MCP Configurations if provided
			if configPath != "" {
				if err := discovery.ParseMCPConfig(configPath, sg); err != nil {
					return fmt.Errorf("failed processing mcp config: %w", err)
				}
			}

			// 3. Run local Rego analysis
			ctx := context.Background()
			res, err := analysis.EvaluateGraph(ctx, sg)
			if err != nil {
				return fmt.Errorf("failed to analyze graph: %w", err)
			}

			// 4. Output Results
			if sarifOutput {
				sarifData, err := reporter.GenerateSARIF(res)
				if err != nil {
					return err
				}
				fmt.Println(string(sarifData))
			} else {
				reporter.RenderTerminalTree(sg)
				fmt.Printf("Scan complete. Allowed: %v, Violations: %d\n", res.Allowed, len(res.Violations))
				for _, v := range res.Violations {
					fmt.Printf(" - [VIOLATION] %s\n", v)
				}
			}

			if !res.Allowed {
				os.Exit(1)
			}

			return nil
		},
	}

	scanCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to mcp.json file")
	scanCmd.Flags().BoolVar(&sarifOutput, "sarif", false, "Output results in SARIF v2.1.0 format")

	rootCmd.AddCommand(scanCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
