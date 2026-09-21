package graph

type Builder struct {
	graph *SecurityGraph
}

func NewBuilder() *Builder {
	return &Builder{
		graph: NewSecurityGraph(),
	}
}

func (b *Builder) Graph() *SecurityGraph {
	return b.graph
}

func (b *Builder) LinkServerToTool(serverID string, toolID string) {
	edge := &Edge{
		ID:       serverID + "->" + toolID,
		SourceID: serverID,
		TargetID: toolID,
		Type:     EdgeHosts,
	}
	b.graph.AddEdge(edge)
}
