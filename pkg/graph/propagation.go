package graph

type NodeType string

const (
	NodeAgent     NodeType = "AGENT"
	NodeMCPServer NodeType = "MCP_SERVER"
	NodeTool      NodeType = "TOOL"
	NodeResource  NodeType = "RESOURCE"
)

type Capability string

const (
	CapReadDatabase  Capability = "READ_DATABASE"
	CapWriteDatabase Capability = "WRITE_DATABASE"
	CapExecCommand   Capability = "EXEC_COMMAND"
	CapFileSystem    Capability = "FILE_SYSTEM"
	CapNetworkAccess Capability = "NETWORK_ACCESS"
)

type Node struct {
	ID           string       `json:"id"`
	Type         NodeType     `json:"type"`
	Name         string       `json:"name"`
	Capabilities []Capability `json:"capabilities"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type SecurityGraph struct {
	Nodes map[string]Node `json:"nodes"`
	Edges []Edge          `json:"edges"`
}

func NewSecurityGraph() *SecurityGraph {
	return &SecurityGraph{
		Nodes: make(map[string]Node),
		Edges: make([]Edge, 0),
	}
}

func (g *SecurityGraph) AddNode(n Node) { g.Nodes[n.ID] = n }

func (g *SecurityGraph) AddEdge(from, to string) {
	for _, e := range g.Edges {
		if e.From == from && e.To == to {
			return
		}
	}
	g.Edges = append(g.Edges, Edge{From: from, To: to})
}

// PropagateTransitiveCapabilities propagates downstream capabilities to root Agents
func (g *SecurityGraph) PropagateTransitiveCapabilities() {
	adj := make(map[string][]string)
	for _, e := range g.Edges {
		adj[e.From] = append(adj[e.From], e.To)
	}

	for id, node := range g.Nodes {
		if node.Type == NodeAgent {
			visited := make(map[string]bool)
			inherited := g.collectCaps(id, adj, visited)
			node.Capabilities = mergeUniqueCaps(node.Capabilities, inherited)
			g.Nodes[id] = node
		}
	}
}

func (g *SecurityGraph) collectCaps(curr string, adj map[string][]string, visited map[string]bool) []Capability {
	if visited[curr] {
		return nil
	}
	visited[curr] = true

	var caps []Capability
	currNode := g.Nodes[curr]
	caps = append(caps, currNode.Capabilities...)

	for _, nxt := range adj[curr] {
		caps = append(caps, g.collectCaps(nxt, adj, visited)...)
	}
	return caps
}

func mergeUniqueCaps(a, b []Capability) []Capability {
	set := make(map[Capability]bool)
	for _, c := range a { set[c] = true }
	for _, c := range b { set[c] = true }
	out := make([]Capability, 0, len(set))
	for c := range set { out = append(out, c) }
	return out
}
