package rules

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/nareshgcv/raaya/pkg/graph"
)

// RAAYA003: literal credentials in MCP configs. Discovery records only the
// names of the env vars/headers holding them, never the values.
func hardcodedSecrets(g *graph.Graph) []Finding {
	var out []Finding
	for _, s := range g.NodesOfType(graph.NodeMCPServer) {
		if v := s.Metadata[graph.MetaHardcodedSecrets]; v != "" {
			out = append(out, newFinding(g, "RAAYA003", SevHigh, "", s.ID,
				fmt.Sprintf("Server %q has a literal credential in its config (%s); reference an environment variable or input instead, and rotate it because it is in git history", s.Name, v)))
		}
	}
	return out
}

// RAAYA007: a remote server over plain http sends tool calls, results and any
// Authorization header in cleartext. Loopback addresses are exempt.
func plaintextRemoteServers(g *graph.Graph) []Finding {
	var out []Finding
	for _, s := range g.NodesOfType(graph.NodeMCPServer) {
		raw := s.Metadata[graph.MetaURL]
		if raw == "" {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil || !strings.EqualFold(u.Scheme, "http") || isLoopback(u.Hostname()) {
			continue
		}
		out = append(out, newFinding(g, "RAAYA007", SevHigh, "", s.ID,
			fmt.Sprintf("Server %q is reached over plain http (%s); use https", s.Name, raw)))
	}
	return out
}

func isLoopback(host string) bool {
	switch strings.ToLower(host) {
	case "localhost", "127.0.0.1", "::1", "[::1]":
		return true
	}
	return strings.HasPrefix(host, "127.")
}
