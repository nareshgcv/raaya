// Package analysis turns a security graph into findings, policy results and
// diffs. The built-in checks live in the rules subpackage.
package analysis

import (
	"errors"
	"fmt"

	"github.com/nareshgcv/raaya/pkg/analysis/rules"
	"github.com/nareshgcv/raaya/pkg/blastradius"
	"github.com/nareshgcv/raaya/pkg/graph"
)

// Re-exported so callers only need this package.
type (
	Finding  = rules.Finding
	Severity = rules.Severity
)

const (
	SevHigh   = rules.SevHigh
	SevMedium = rules.SevMedium
	SevLow    = rules.SevLow
)

// Evaluate runs the built-in rules over g. Results are sorted.
func Evaluate(g *graph.Graph) []Finding { return rules.Run(g) }

// ParseSeverity parses HIGH, MEDIUM or LOW.
func ParseSeverity(s string) (Severity, bool) { return rules.ParseSeverity(s) }

// SortFindings orders findings by severity, then rule, agent and node.
func SortFindings(fs []Finding) { rules.Sort(fs) }

// AnyAtOrAbove reports whether any finding is at least sev.
func AnyAtOrAbove(fs []Finding, sev Severity) bool { return rules.AnyAtOrAbove(fs, sev) }

// ErrRegoUnavailable is returned when policies are requested from a binary
// built without Rego support.
var ErrRegoUnavailable = errors.New("this raaya binary was built without Rego support; rebuild with: go build -tags rego ./cmd/raaya")

// PolicyInput is the document Rego policies receive as `input`.
//
//	input.graph.nodes[id]          node (id, name, type, metadata)
//	input.graph.edges[_]           edge (source_id, target_id, relation, permission)
//	input.reach[agent_id][node_id] effective permission ("" = reachable, no limit)
//	input.findings[_]              built-in findings
type PolicyInput struct {
	Graph    *graph.Graph                                `json:"graph"`
	Reach    map[string]map[string]graph.PermissionLevel `json:"reach"`
	Findings []Finding                                   `json:"findings"`
}

// NewPolicyInput assembles the policy input for g.
func NewPolicyInput(g *graph.Graph, findings []Finding) PolicyInput {
	reach := map[string]map[string]graph.PermissionLevel{}
	for _, a := range g.NodesOfType(graph.NodeAgent) {
		reach[a.ID] = blastradius.From(g, a.ID)
	}
	if findings == nil {
		findings = []Finding{}
	}
	return PolicyInput{Graph: g, Reach: reach, Findings: findings}
}

// policyFinding converts one element of data.raaya.deny into a Finding.
// Elements may be plain strings or objects with msg/message, rule_id,
// severity, node_id and agent_id.
func policyFinding(item any) Finding {
	f := Finding{RuleID: "POLICY", Severity: SevMedium}
	switch v := item.(type) {
	case string:
		f.Message = v
	case map[string]any:
		str := func(k string) string { s, _ := v[k].(string); return s }
		f.Message = str("msg")
		if f.Message == "" {
			f.Message = str("message")
		}
		if id := str("rule_id"); id != "" {
			f.RuleID = id
		}
		if sev, ok := ParseSeverity(str("severity")); ok {
			f.Severity = sev
		}
		f.NodeID, f.AgentID = str("node_id"), str("agent_id")
	default:
		f.Message = fmt.Sprint(v)
	}
	return f
}
