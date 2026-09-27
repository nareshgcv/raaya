package subcmds

import (
	"fmt"

	"github.com/spf13/cobra"
	"raaya/pkg/discovery"
	"raaya/pkg/graph"
)

var liveFlag bool

var DoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Inspect local agent environment and running MCP endpoints",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🏥 Running Raaya Environment Doctor...")

		sg := graph.NewSecurityGraph()

		if liveFlag {
			fmt.Println("\n📡 Probing active local TCP/SSE ports...")
			results := discovery.ProbeLocalhostPorts(sg)
			for _, res := range results {
				if res.Open {
					fmt.Printf("  [OPEN] Active endpoint detected at %s\n", res.Target)
				}
			}
		} else {
			fmt.Println("✔ Static environment configuration valid. (Use --live for port scanning)")
		}
	},
}

func init() {
	DoctorCmd.Flags().BoolVarP(&liveFlag, "live", "l", false, "Probe local TCP/SSE ports for active unauthenticated MCP servers")
}
