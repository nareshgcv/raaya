package subcmds

import (
	"fmt"

	"github.com/spf13/cobra"
	"raaya/pkg/discovery"
	"raaya/pkg/graph"
)

var GraphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Export in-memory Security Graph as Mermaid or DOT",
	Run: func(cmd *cobra.Command, args []string) {
		sg := graph.NewSecurityGraph()

		_ = discovery.DiscoverMCPConfigs("mcp.json", sg)
		_ = discovery.DiscoverMCPConfigs(".cursor/mcp.json", sg)
		_ = discovery.ScanAnnotations(".", sg)

		fmt.Println(sg.ToMermaid())
	},
}
