package discovery

import (
	"fmt"
	"net"
	"time"

	"raaya/pkg/graph"
)

var commonMCPPorts = []int{8000, 8080, 3000, 5000, 9000, 9090}

type PortProbeResult struct {
	Port   int
	Open   bool
	Target string
}

func ProbeLocalhostPorts(sg *graph.SecurityGraph) []PortProbeResult {
	var results []PortProbeResult

	for _, port := range commonMCPPorts {
		target := fmt.Sprintf("127.0.0.1:%d", port)
		conn, err := net.DialTimeout("tcp", target, 200*time.Millisecond)

		if err == nil {
			conn.Close()
			results = append(results, PortProbeResult{Port: port, Open: true, Target: target})

			serverID := fmt.Sprintf("live_mcp_server:%d", port)
			sg.AddNode(graph.AssetNode{
				ID:         serverID,
				Kind:       graph.KindMCPServer,
				Name:       fmt.Sprintf("Live MCP Server (Port %d)", port),
				SourceFile: "localhost",
				Metadata: map[string]string{
					"transport": "tcp/sse",
					"endpoint":  target,
				},
			})
		} else {
			results = append(results, PortProbeResult{Port: port, Open: false, Target: target})
		}
	}
	return results
}
