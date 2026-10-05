package graph

import (
	"bytes"
	"testing"
)

func TestAddEdgeDeduplicates(t *testing.T) {
	g := NewGraph()
	e := Edge{SourceID: "a", TargetID: "b", Relation: "x", Permission: PermRead}
	if !g.AddEdge(e) || g.AddEdge(e) {
		t.Fatal("second identical edge should be rejected")
	}
	e.Permission = PermWrite
	if !g.AddEdge(e) {
		t.Fatal("same endpoints with a different permission is a distinct edge")
	}
}

func TestSetPermissionMergesDuplicates(t *testing.T) {
	g := NewGraph()
	g.AddEdge(Edge{SourceID: "s", TargetID: "t", Relation: "exposes", Permission: PermRead})
	g.AddEdge(Edge{SourceID: "s", TargetID: "t", Relation: "exposes", Permission: PermWrite})
	g.SetPermission(func(e Edge) bool { return e.TargetID == "t" }, PermExecute)
	if len(g.Edges) != 1 || g.Edges[0].Permission != PermExecute {
		t.Fatalf("got %+v", g.Edges)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	g := NewGraph()
	g.AddNode(&Node{ID: "agent:a", Name: "A", Type: NodeAgent})
	g.AddNode(&Node{ID: "server:s", Name: "s", Type: NodeMCPServer})
	g.AddEdge(Edge{SourceID: "agent:a", TargetID: "server:s", Relation: "connects"})

	var buf bytes.Buffer
	if err := g.WriteJSON(&buf); err != nil {
		t.Fatal(err)
	}
	back, err := ReadJSON(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(back.Nodes) != 2 || len(back.Edges) != 1 || back.AddEdge(g.Edges[0]) {
		t.Fatalf("round trip lost data or the duplicate index: %+v", back)
	}
}
