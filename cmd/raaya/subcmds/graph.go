package subcmds

import (
	"fmt"

	"raaya/pkg/discovery"
	"raaya/pkg/graph"

	"github.com/spf13/cobra"
)

var format string

var GraphCmd = &cobra.Command{
	Use:   "graph [path]",
	Short: "Export the topology graph in Mermaid or DOT format",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		analyzer := discovery.NewHybridAnalyzer()
		g, err := analyzer.DiscoverGraph(cmd.Context(), path)
		if err != nil {
			return err
		}

		if format == "dot" {
			fmt.Println(graph.RenderDOT(g))
		} else {
			fmt.Println(graph.RenderMermaid(g))
		}
		return nil
	},
}

func init() {
	GraphCmd.Flags().StringVarF(&format, "format", "f", "mermaid", "Output format: mermaid | dot")
}
