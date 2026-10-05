// Package graph is Raaya's security graph: agents, MCP servers, tools and
// resources, joined by edges that may carry a permission level.
package graph

import (
	"sort"
	"strings"
)

// NodeType is the kind of entity a node represents.
type NodeType string

const (
	NodeAgent     NodeType = "AGENT"
	NodeMCPServer NodeType = "MCP_SERVER"
	NodeTool      NodeType = "TOOL"
	NodeResource  NodeType = "RESOURCE"
)

// PermissionLevel is an access level. The zero value (PermNone) marks a
// structural edge: it connects two nodes without limiting what flows along it.
type PermissionLevel string

const (
	PermNone    PermissionLevel = ""
	PermRead    PermissionLevel = "READ"
	PermWrite   PermissionLevel = "WRITE"
	PermExecute PermissionLevel = "EXECUTE"
	PermAdmin   PermissionLevel = "ADMIN"
)

// Rank orders permissions: READ < WRITE < EXECUTE < ADMIN. PermNone is 0.
func (p PermissionLevel) Rank() int {
	switch p {
	case PermRead:
		return 1
	case PermWrite:
		return 2
	case PermExecute:
		return 3
	case PermAdmin:
		return 4
	default:
		return 0
	}
}

// ParsePermission parses READ, WRITE, EXECUTE or ADMIN (case-insensitive).
func ParsePermission(s string) (PermissionLevel, bool) {
	p := PermissionLevel(strings.ToUpper(strings.TrimSpace(s)))
	if p.Rank() == 0 {
		return PermNone, false
	}
	return p, true
}

// MaxPermission returns the strongest of ps, or PermNone.
func MaxPermission(ps ...PermissionLevel) PermissionLevel {
	best := PermNone
	for _, p := range ps {
			best = p
		}
	}
	return best
}

// Metadata keys and values shared by discovery and analysis.
const (
	MetaSource           = "source"            // config file a node came from
	MetaFile             = "file"              // source file (tools found in code, subagents)
	MetaLine             = "line"              // line in MetaFile
	MetaTransport        = "transport"         // stdio, http, source
	MetaURL              = "url"               // redacted server URL
	MetaPermissionSource = "permission_source" // how a tool's permission was decided
	MetaHardcodedSecrets = "hardcoded_secrets" // env/header names holding literal secrets
	MetaUnpinnedPackage  = "unpinned_package"  // package launched without a pinned version
	MetaBroadPath        = "broad"             // "true" on filesystem-root/home resources
	MetaParentAgent      = "parent_agent"      // subagents: the agent they belong to

	PermissionAssumed = "assumed" // no annotations; permission is a conservative guess
)

// Node is one entity in the graph.
type Node struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     NodeType          `json:"type"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Edge is a directed relationship. Permission caps what the source can do to
// the target; PermNone means the edge is structural.
type Edge struct {
	SourceID   string          `json:"source_id"`
	TargetID   string          `json:"target_id"`
	Relation   string          `json:"relation"`
	Permission PermissionLevel `json:"permission,omitempty"`
	Inferred   bool            `json:"inferred,omitempty"`
}

// Key identifies an edge. Two edges between the same nodes with different
// permissions are distinct edges.
func (e Edge) Key() string {
	return e.SourceID + "\x00" + e.TargetID + "\x00" + e.Relation + "\x00" + string(e.Permission)
}

// Graph is the directed security graph. Build it with the methods in builder.go.
type Graph struct {
	Nodes map[string]*Node `json:"nodes"`
	Edges []Edge           `json:"edges"`

	index map[string]struct{}
}

// SortedNodes returns all nodes ordered by ID.
func (g *Graph) SortedNodes() []*Node {
	out := make([]*Node, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// NodesOfType returns the nodes of type t ordered by ID.
func (g *Graph) NodesOfType(t NodeType) []*Node {
	var out []*Node
	for _, n := range g.SortedNodes() {
		if n.Type == t {
			out = append(out, n)
		}
	}
	return out
}

// SortedEdges returns a sorted copy of the edges.
func (g *Graph) SortedEdges() []Edge {
	out := append([]Edge(nil), g.Edges...)
	sort.Slice(out, func(i, j int) bool { return edgeLess(out[i], out[j]) })
	return out
}

// Parent returns the lowest-sorting source of an edge into id, or "".
func (g *Graph) Parent(id string) string {
	best := ""
	for _, e := range g.Edges {
		if e.TargetID == id && (best == "" || e.SourceID < best) {
			best = e.SourceID
		}
	}
	return best
}

func edgeLess(a, b Edge) bool {
	if a.SourceID != b.SourceID {
		return a.SourceID < b.SourceID
	}
	if a.TargetID != b.TargetID {
		return a.TargetID < b.TargetID
	}
	if a.Relation != b.Relation {
		return a.Relation < b.Relation
	}
	return a.Permission.Rank() < b.Permission.Rank()
}
