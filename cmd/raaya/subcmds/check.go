package subcmds

import (
	"os"

	"github.com/spf13/cobra"
	"raaya/pkg/analysis"
	"raaya/pkg/discovery"
	"raaya/pkg/graph"
	"raaya/pkg/reporter"
)

var CheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Scan repository for AI agent and MCP security violations",
	Run: func(cmd *cobra.Command, args []string) {
		sg := graph.NewSecurityGraph()

		// 1. Discover MCP configuration files
		_ = discovery.DiscoverMCPConfigs("mcp.json", sg)
		_ = discovery.DiscoverMCPConfigs(".cursor/mcp.json", sg)

		// 2. Evaluate Policy Rules
		engine := analysis.NewEngine()
		findings := engine.Run(sg)

		// 3. Render Output
		reporter.PrintTerminalReport(findings)

		// Exit code 1 if critical/high vulnerabilities exist
		if len(findings) > 0 {
			os.Exit(1)
		}
	},
}
