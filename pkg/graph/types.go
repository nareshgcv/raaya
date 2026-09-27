package graph

import "fmt"

type NodeKind string

const (
	KindAgentPrompt NodeKind = "agent_prompt"
	KindMCPServer   NodeKind = "mcp_server"
	KindToolDef     NodeKind = "tool_definition"
	KindSecret      NodeKind = "secret"
)

type AssetNode struct {
	ID         string            `json:"id"`
	Kind       NodeKind          `json:"kind"`
	Name       string            `json:"name"`
	SourceFile string            `json:"source_file"`
	LineNumber int               `json:"line_number,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type EdgeRelation string

const (
	RelLoadsPrompt EdgeRelation = "LOADS_PROMPT"
	RelUsesServer  EdgeRelation = "USES_SERVER"
	RelExposesTool EdgeRelation = "EXPOSES_TOOL"
	RelUsesSecret  EdgeRelation = "USES_SECRET"
)

type DirectEdge struct {
	FromID   string       `json:"from_id"`
	ToID     string       `json:"to_id"`
	Relation EdgeRelation `json:"relation"`
}

type SecurityGraph struct {
	Nodes map[string]AssetNode `json:"nodes"`
	Edges []DirectEdge         `json:"edges"`
}

func NewSecurityGraph() *SecurityGraph {
	return &SecurityGraph{
		Nodes: make(map[string]AssetNode),
		Edges: make([]DirectEdge, 0),
	}
}

func (g *SecurityGraph) AddNode(node AssetNode) {
	g.Nodes[node.ID] = node
}

func (g *SecurityGraph) AddEdge(fromID, toID string, rel EdgeRelation) {
	g.Edges = append(g.Edges, DirectEdge{
		FromID:   fromID,
		ToID:     toID,
		Relation: rel,
	})
}
