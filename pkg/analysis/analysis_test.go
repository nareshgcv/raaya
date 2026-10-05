package analysis

import (
	"strings"
	"testing"

	"github.com/nareshgcv/raaya/pkg/graph"
)

func graphWith(edges ...graph.Edge) *graph.Graph {
	g := graph.NewGraph()
	types := map[string]graph.NodeType{"agent": graph.NodeAgent, "server": graph.NodeMCPServer, "tool": graph.NodeTool, "resource": graph.NodeResource}
	for _, e := range edges {
		for _, id := range []string{e.SourceID, e.TargetID} {
			if _, ok := g.Nodes[id]; !ok {
				prefix, _, _ := strings.Cut(id, ":")
				g.AddNode(&graph.Node{ID: id, Name: id, Type: types[prefix]})
			}
		}
		g.AddEdge(e)
	}
	return g
}

var (
	connect   = graph.Edge{SourceID: "agent:a", TargetID: "server:s", Relation: "connects"}
	readTool  = graph.Edge{SourceID: "server:s", TargetID: "tool:s:read", Relation: "exposes", Permission: graph.PermRead}
	writeTool = graph.Edge{SourceID: "server:s", TargetID: "tool:s:deploy", Relation: "exposes", Permission: graph.PermWrite}
)

func TestDiffFlagsNewWriteReachAsRegression(t *testing.T) {
	d := ComputeDiff(graphWith(connect, readTool), graphWith(connect, readTool, writeTool), graph.PermWrite)
	if d.Regressions != 1 || len(d.NewlyReachable) != 1 || d.NewlyReachable[0].NodeID != "tool:s:deploy" {
		t.Fatalf("got %+v", d)
	}
}

func TestDiffDoesNotFailOnNewReadOnlyReach(t *testing.T) {
	extraRead := graph.Edge{SourceID: "server:s", TargetID: "tool:s:list", Relation: "exposes", Permission: graph.PermRead}
	d := ComputeDiff(graphWith(connect, readTool), graphWith(connect, readTool, extraRead), graph.PermWrite)
	if d.Regressions != 0 || len(d.NewlyReachable) != 1 {
		t.Fatalf("got %+v", d)
	}
}

func TestDiffDetectsEscalation(t *testing.T) {
	escalated := readTool
	escalated.Permission = graph.PermWrite
	d := ComputeDiff(graphWith(connect, readTool), graphWith(connect, escalated), graph.PermWrite)
	if len(d.Escalations) != 1 || d.Escalations[0].Before != graph.PermRead || d.Escalations[0].After != graph.PermWrite {
		t.Fatalf("escalations: %+v", d.Escalations)
	}
	if d.Regressions != 1 || len(d.AddedEdges) != 1 || len(d.RemovedEdges) != 1 {
		t.Fatalf("got %+v", d)
	}
}

func TestDiffReportsNewHighFinding(t *testing.T) {
	base := graphWith(connect, readTool)
	head := graphWith(connect, readTool)
	head.Nodes["server:s"].Metadata[graph.MetaHardcodedSecrets] = "env API_TOKEN"
	d := ComputeDiff(base, head, graph.PermWrite)
	if len(d.NewFindings) != 1 || d.NewFindings[0].RuleID != "RAAYA003" || d.Regressions != 1 {
		t.Fatalf("got %+v", d)
	}
}

func TestPolicyFindingParsesObjectsAndStrings(t *testing.T) {
	f := policyFinding(map[string]any{"msg": "no", "severity": "high", "rule_id": "ORG-1", "node_id": "tool:x"})
	if f.RuleID != "ORG-1" || f.Severity != SevHigh || f.Message != "no" || f.NodeID != "tool:x" {
		t.Fatalf("object: %+v", f)
	}
	if f := policyFinding("plain"); f.RuleID != "POLICY" || f.Message != "plain" {
		t.Fatalf("string: %+v", f)
	}
}
