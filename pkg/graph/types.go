package graph

import "time"

type NodeType string

const (
	NodeAgent    NodeType = "AGENT"
	NodeTool     NodeType = "TOOL"
	NodeServer   NodeType = "SERVER"
	NodeResource NodeType = "RESOURCE"
	NodeModel    NodeType = "MODEL"
)

type EdgeType string

const (
	EdgeInvokes   EdgeType = "INVOKES"
	EdgeHosts     EdgeType = "HOSTS"
	EdgeAccesses  EdgeType = "ACCESSES"
	EdgeUsesModel EdgeType = "USES_MODEL"
)

type SourceLocation struct {
	FilePath  string `json:"file_path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type Node struct {
	ID         string                 `json:"id"`
	Type       NodeType               `json:"type"`
	Name       string                 `json:"name"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Location   *SourceLocation        `json:"location,omitempty"`
}

type Edge struct {
	ID       string                 `json:"id"`
	SourceID string                 `json:"source_id"`
	TargetID string                 `json:"target_id"`
	Type     EdgeType               `json:"type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type SecurityGraph struct {
	Nodes     map[string]*Node `json:"nodes"`
	Edges     []*Edge          `json:"edges"`
	CreatedAt time.Time        `json:"created_at"`
}

func NewSecurityGraph() *SecurityGraph {
	return &SecurityGraph{
		Nodes:     make(map[string]*Node),
		Edges:     make([]*Edge, 0),
		CreatedAt: time.Now(),
	}
}

func (g *SecurityGraph) AddNode(node *Node) {
	g.Nodes[node.ID] = node
}

func (g *SecurityGraph) AddEdge(edge *Edge) {
	g.Edges = append(g.Edges, edge)
}
