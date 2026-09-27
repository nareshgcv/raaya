package rules

import (
	"fmt"
	"strings"

	"raaya/pkg/graph"
)

type UnauthenticatedEndpointRule struct{}

func (r *UnauthenticatedEndpointRule) ID() string { return "RAAYA-002-UNAUTH-ENDPOINT" }

func (r *UnauthenticatedEndpointRule) Evaluate(sg *graph.SecurityGraph) []Finding {
	var findings []Finding

	for _, node := range sg.Nodes {
		if node.Kind != graph.KindMCPServer {
			continue
		}

		url := node.Metadata["url"]
		command := node.Metadata["command"]
		transport := node.Metadata["transport"]

		// Check HTTP/SSE transport endpoints or live local listeners
		isNetworkEndpoint := strings.HasPrefix(url, "http://") || 
			strings.HasPrefix(url, "https://") || 
			transport == "tcp/sse" || 
			strings.Contains(command, "sse")

		if !isNetworkEndpoint {
			continue // Stdio-based local pipes do not expose HTTP endpoints
		}

		// Verify if the server node connects to any Secret or Auth Header edge
		hasAuthSecret := false
		for _, edge := range sg.Edges {
			if edge.FromID == node.ID && edge.Relation == graph.RelUsesSecret {
				hasAuthSecret = true
				break
			}
		}

		// Flag HTTP/SSE endpoints lacking connected credentials or headers
		if !hasAuthSecret {
			endpoint := url
			if endpoint == "" {
				endpoint = node.Metadata["endpoint"]
			}

			findings = append(findings, Finding{
				RuleID:           r.ID(),
				Severity:         SeverityHigh,
				Message:          fmt.Sprintf("MCP Server '%s' exposes a network endpoint (%s) without an explicit auth header or secret", node.Name, endpoint),
				AssetID:          node.ID,
				FilePath:         node.SourceFile,
				AutofixAvailable: false,
			})
		}
	}

	return findings
}
