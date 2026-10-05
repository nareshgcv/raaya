package blastradius

import (
	"strings"
	"testing"

	"github.com/nareshgcv/raaya/pkg/graph"
)

func build(edges ...graph.Edge) *graph.Graph {
	g := graph.NewGraph()
	for _, e := range edges {
		for _, id := range []string{e.SourceID, e.TargetID} {
			if _, ok := g.Nodes[id]; !ok {
				g.AddNode(&graph.Node{ID: id, Name: id, Type: typeFor(id)})
			}
		}
		g.AddEdge(e)
	}
	return g
}

func typeFor(id string) graph.NodeType {
	switch {
	case strings.HasPrefix(id, "agent"):
		return graph.NodeAgent
	case strings.HasPrefix(id, "server"):
		return graph.NodeMCPServer
	case strings.HasPrefix(id, "tool"):
		return graph.NodeTool
	default:
		return graph.NodeResource
	}
}

func TestReachUsesWeakestEdgeOnPath(t *testing.T) {
	g := build(
		graph.Edge{SourceID: "agent", TargetID: "server", Relation: "connects"},
		graph.Edge{SourceID: "server", TargetID: "tool-reader", Relation: "exposes", Permission: graph.PermRead},
		graph.Edge{SourceID: "tool-reader", TargetID: "db", Relation: "uses", Permission: graph.PermAdmin},
	)
	r := From(g, "agent")
	if got := r["db"]; got != graph.PermRead {
		t.Fatalf("db: got %q, want READ (capped by the read-only tool)", got)
	}
	if p, ok := r["server"]; !ok || p != graph.PermNone {
		t.Fatalf("server: got %q/%v, want structural reach", p, ok)
	}
}

func TestReachTakesStrongestPath(t *testing.T) {
	g := build(
		graph.Edge{SourceID: "agent", TargetID: "server", Relation: "connects"},
		graph.Edge{SourceID: "server", TargetID: "tool-reader", Relation: "exposes", Permission: graph.PermRead},
		graph.Edge{SourceID: "server", TargetID: "tool-writer", Relation: "exposes", Permission: graph.PermWrite},
		graph.Edge{SourceID: "tool-reader", TargetID: "db", Relation: "uses", Permission: graph.PermAdmin},
		graph.Edge{SourceID: "tool-writer", TargetID: "db", Relation: "uses", Permission: graph.PermAdmin},
	)
	if got := From(g, "agent")["db"]; got != graph.PermWrite {
		t.Fatalf("db: got %q, want WRITE", got)
	}
}

func TestReachHandlesCycles(t *testing.T) {
	g := build(
		graph.Edge{SourceID: "a", TargetID: "b", Relation: "x", Permission: graph.PermRead},
		graph.Edge{SourceID: "b", TargetID: "a", Relation: "x", Permission: graph.PermWrite},
		graph.Edge{SourceID: "b", TargetID: "c", Relation: "x", Permission: graph.PermExecute},
	)
	r := From(g, "a")
	if _, self := r["a"]; self {
		t.Fatal("start node must not appear in its own reach")
	}
	if r["b"] != graph.PermRead || r["c"] != graph.PermRead {
		t.Fatalf("got %v", r)
	}
}

func TestCalculateSplitsDirectAndTransitive(t *testing.T) {
	g := build(
		graph.Edge{SourceID: "server", TargetID: "tool", Relation: "exposes", Permission: graph.PermWrite},
		graph.Edge{SourceID: "tool", TargetID: "res", Relation: "uses", Permission: graph.PermWrite},
	)
	r := Calculate(g, "server")
	if len(r.Direct) != 1 || r.Direct[0].NodeID != "tool" {
		t.Fatalf("direct: %+v", r.Direct)
	}
	if len(r.Transitive) != 1 || r.Transitive[0].NodeID != "res" {
		t.Fatalf("transitive: %+v", r.Transitive)
	}
	if r.Score != 3+6 {
		t.Fatalf("score: got %d, want 9", r.Score)
	}
}
