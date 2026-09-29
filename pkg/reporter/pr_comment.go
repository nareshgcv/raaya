package reporter

import (
	"fmt"
	"strings"

	"raaya/pkg/analysis"
	"raaya/pkg/graph"
)

type PRReporter struct{}

func NewPRReporter() *PRReporter {
	return &PRReporter{}
}

func (r *PRReporter) RenderComment(diff *analysis.GraphDiff, violations []string, targetGraph *graph.SecurityGraph) string {
	var sb strings.Builder

	sb.WriteString("## 🛡️ Raaya Agentic Security Surface Report\n\n")

	if diff.SecuritySurface == analysis.SurfaceExpanded {
		sb.WriteString("**Security Surface Status:** ⚠️ `EXPANDED` (Capabilities Escalated)\n\n")
	} else {
		sb.WriteString("**Security Surface Status:** ✅ `STABLE`\n\n")
	}

	if len(violations) > 0 {
		sb.WriteString("### 🚨 Policy Violations\n")
		for _, v := range violations {
			sb.WriteString(fmt.Sprintf("- %s\n", v))
		}
		sb.WriteString("\n")
	}

	if len(diff.AddedPaths) > 0 {
		sb.WriteString("### 📈 New Reachable Capabilities\n")
		for _, p := range diff.AddedPaths {
			sb.WriteString(fmt.Sprintf("- Agent `%s` $\\rightarrow$ Target `%s` | Capabilities: `%v`\n",
				p.AgentID, p.TargetID, p.Capabilities))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### 🕸️ Security DAG Visualization\n")
	sb.WriteString("```mermaid\n")
	sb.WriteString("graph TD\n")
	for _, node := range targetGraph.Nodes {
		sb.WriteString(fmt.Sprintf("  %s[\"%s (%s)\"]\n", node.ID, node.Name, node.Type))
	}
	for _, edge := range targetGraph.Edges {
		sb.WriteString(fmt.Sprintf("  %s --> %s\n", edge.From, edge.To))
	}
	sb.WriteString("```\n")

	return sb.String()
}
