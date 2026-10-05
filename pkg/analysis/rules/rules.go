// Package rules holds Raaya's built-in checks. Each check reads the security
// graph and returns findings; Run executes them all.
package rules

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/nareshgcv/raaya/pkg/graph"
)

// Severity of a finding.
type Severity string

const (
	SevHigh   Severity = "HIGH"
	SevMedium Severity = "MEDIUM"
	SevLow    Severity = "LOW"
)

// Rank orders severities: LOW < MEDIUM < HIGH.
func (s Severity) Rank() int {
	switch s {
	case SevHigh:
		return 3
	case SevMedium:
		return 2
	case SevLow:
		return 1
	default:
		return 0
	}
}

// ParseSeverity parses HIGH, MEDIUM or LOW (case-insensitive).
func ParseSeverity(s string) (Severity, bool) {
	sev := Severity(strings.ToUpper(strings.TrimSpace(s)))
	return sev, sev.Rank() > 0
}

// Finding is one reported issue.
type Finding struct {
	RuleID   string   `json:"rule_id"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	AgentID  string   `json:"agent_id,omitempty"`
	NodeID   string   `json:"node_id,omitempty"`
	File     string   `json:"file,omitempty"`
	Line     int      `json:"line,omitempty"`
}

// Key identifies a finding across scans, independent of message wording.
func (f Finding) Key() string {
	return f.RuleID + "\x00" + f.AgentID + "\x00" + f.NodeID
}

// Rule describes a built-in check (used for SARIF rule metadata).
type Rule struct {
	ID          string
	Name        string
	Severity    Severity
	Description string
}

// All lists the built-in rules.
var All = []Rule{
	{"RAAYA001", "execute-reachable", SevHigh, "An agent can reach a tool or resource with EXECUTE or ADMIN capability."},
	{"RAAYA002", "write-to-resource", SevMedium, "An agent can modify an external resource."},
	{"RAAYA003", "hardcoded-secret", SevHigh, "An MCP server configuration contains a literal credential."},
	{"RAAYA004", "unpinned-server-package", SevMedium, "An MCP server is launched from a package runner without a pinned version."},
	{"RAAYA005", "broad-filesystem-access", SevMedium, "An MCP server is given the filesystem root or a home directory."},
	{"RAAYA006", "undeclared-tool-capability", SevLow, "Tools declare neither MCP annotations nor @raaya:capability, so Raaya assumes they change state."},
	{"RAAYA007", "plaintext-remote-server", SevHigh, "A remote MCP server is reached over unencrypted HTTP."},
}

// Reach-based rules (RAAYA001, 002, 005) live in unauth.go; credential rules
// (RAAYA003, 007) in secrets.go; supply-chain and hygiene rules here.
var checks = []func(*graph.Graph) []Finding{
	reachableCapabilities,
	broadFilesystemAccess,
	hardcodedSecrets,
	plaintextRemoteServers,
	unpinnedPackages,
	undeclaredCapabilities,
}

// Run executes every built-in rule over g and returns sorted findings.
func Run(g *graph.Graph) []Finding {
	var out []Finding
	for _, check := range checks {
		out = append(out, check(g)...)
	}
	Sort(out)
	return out
}

// RAAYA004
func unpinnedPackages(g *graph.Graph) []Finding {
	var out []Finding
	for _, s := range g.NodesOfType(graph.NodeMCPServer) {
		if v := s.Metadata[graph.MetaUnpinnedPackage]; v != "" {
			out = append(out, newFinding(g, "RAAYA004", SevMedium, "", s.ID,
				fmt.Sprintf("Server %q runs %q without a pinned version, so a new upstream release changes what your agent executes", s.Name, v)))
		}
	}
	return out
}

// RAAYA006: one finding per server, listing a few example tools.
func undeclaredCapabilities(g *graph.Graph) []Finding {
	assumed := map[string]map[string]bool{}
	for _, e := range g.Edges {
		t := g.Nodes[e.TargetID]
		if e.Relation != "exposes" || t == nil || t.Metadata[graph.MetaPermissionSource] != graph.PermissionAssumed {
			continue
		}
		if assumed[e.SourceID] == nil {
			assumed[e.SourceID] = map[string]bool{}
		}
		assumed[e.SourceID][t.Name] = true
	}
	var out []Finding
	for serverID, set := range assumed {
		names := make([]string, 0, len(set))
		for n := range set {
			names = append(names, n)
		}
		sort.Strings(names)
		example := strings.Join(names[:min(3, len(names))], ", ")
		serverName := serverID
		if n := g.Nodes[serverID]; n != nil {
			serverName = n.Name
		}
		out = append(out, newFinding(g, "RAAYA006", SevLow, "", serverID,
			fmt.Sprintf("%d tool(s) on %q declare no capability (%s); assumed WRITE. Add MCP tool annotations or @raaya:capability", len(names), serverName, example)))
	}
	return out
}

func newFinding(g *graph.Graph, rule string, sev Severity, agentID, nodeID, msg string) Finding {
	file, line := locate(g, nodeID)
	return Finding{RuleID: rule, Severity: sev, Message: msg, AgentID: agentID, NodeID: nodeID, File: file, Line: line}
}

// locate finds the file that defines a node: its source file for tools found
// in code, otherwise the config file of the nearest ancestor server.
func locate(g *graph.Graph, nodeID string) (string, int) {
	seen := map[string]bool{}
	for id := nodeID; id != "" && !seen[id]; id = g.Parent(id) {
		seen[id] = true
		n := g.Nodes[id]
		if n == nil {
			break
		}
		if f := n.Metadata[graph.MetaFile]; f != "" {
			line, _ := strconv.Atoi(n.Metadata[graph.MetaLine])
			return f, line
		}
		if f := n.Metadata[graph.MetaSource]; f != "" {
			return f, 0
		}
	}
	return "", 0
}

func typeLabel(t graph.NodeType) string {
	switch t {
	case graph.NodeTool:
		return "tool"
	case graph.NodeResource:
		return "resource"
	case graph.NodeMCPServer:
		return "server"
	default:
		return "node"
	}
}

// Sort orders findings by severity, then rule, agent and node.
func Sort(fs []Finding) {
	sort.SliceStable(fs, func(i, j int) bool {
		a, b := fs[i], fs[j]
		if a.Severity.Rank() != b.Severity.Rank() {
			return a.Severity.Rank() > b.Severity.Rank()
		}
		if a.RuleID != b.RuleID {
			return a.RuleID < b.RuleID
		}
		if a.AgentID != b.AgentID {
			return a.AgentID < b.AgentID
		}
		if a.NodeID != b.NodeID {
			return a.NodeID < b.NodeID
		}
		return a.Message < b.Message
	})
}

// AnyAtOrAbove reports whether any finding is at least sev.
func AnyAtOrAbove(fs []Finding, sev Severity) bool {
	for _, f := range fs {
		if f.Severity.Rank() >= sev.Rank() {
			return true
		}
	}
	return false
}
