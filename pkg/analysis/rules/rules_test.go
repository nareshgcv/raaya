package rules

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

var connect = graph.Edge{SourceID: "agent:a", TargetID: "server:s", Relation: "connects"}

func TestRunFlagsExecuteAndSecrets(t *testing.T) {
	shell := graph.Edge{SourceID: "server:s", TargetID: "tool:s:shell", Relation: "exposes", Permission: graph.PermExecute}
	g := graphWith(connect, shell)
	g.Nodes["server:s"].Metadata[graph.MetaSource] = ".mcp.json"
	g.Nodes["server:s"].Metadata[graph.MetaHardcodedSecrets] = "env API_TOKEN"

	fs := Run(g)
	if len(fs) != 2 {
		t.Fatalf("want 2 findings, got %+v", fs)
	}
	if fs[0].RuleID != "RAAYA001" || fs[0].NodeID != "tool:s:shell" || fs[0].File != ".mcp.json" {
		t.Errorf("first finding: %+v", fs[0])
	}
	if fs[1].RuleID != "RAAYA003" {
		t.Errorf("second finding: %+v", fs[1])
	}
}

func TestRunAggregatesAssumedTools(t *testing.T) {
	g := graphWith(connect, graph.Edge{SourceID: "server:s", TargetID: "tool:s:deploy", Relation: "exposes", Permission: graph.PermWrite})
	g.Nodes["tool:s:deploy"].Metadata[graph.MetaPermissionSource] = graph.PermissionAssumed
	fs := Run(g)
	if len(fs) != 1 || fs[0].RuleID != "RAAYA006" || fs[0].NodeID != "server:s" {
		t.Fatalf("got %+v", fs)
	}
}

func TestPlaintextRemoteServer(t *testing.T) {
	g := graphWith(connect)
	g.Nodes["server:s"].Metadata[graph.MetaURL] = "http://mcp.example.com/sse"
	if fs := Run(g); len(fs) != 1 || fs[0].RuleID != "RAAYA007" {
		t.Fatalf("got %+v", fs)
	}
	g.Nodes["server:s"].Metadata[graph.MetaURL] = "http://localhost:8080/mcp"
	if fs := Run(g); len(fs) != 0 {
		t.Fatalf("loopback should be exempt, got %+v", fs)
	}
}

func TestSubagentDoesNotRepeatParentFindings(t *testing.T) {
	shell := graph.Edge{SourceID: "server:s", TargetID: "tool:s:shell", Relation: "exposes", Permission: graph.PermExecute}
	sub := graph.Edge{SourceID: "agent:a/sub", TargetID: "tool:s:shell", Relation: "allowed-tool", Permission: graph.PermExecute}
	g := graphWith(connect, shell, sub)
	g.Nodes["agent:a/sub"].Metadata[graph.MetaParentAgent] = "agent:a"
	fs := Run(g)
	if len(fs) != 1 || fs[0].AgentID != "agent:a" {
		t.Fatalf("got %+v", fs)
	}
}
