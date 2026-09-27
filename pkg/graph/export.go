// pkg/graph/export.go
package graph

import (
	"fmt"
	"io"
	"strings"
)

type Graph struct {
	Agents []AgentNode
}

type AgentNode struct {
	Name   string
	Prompt string
	Tools  []ToolNode
}

type ToolNode struct {
	Name     string
	Server   string
	RiskTag  string // e.g., "[OK]" or "[⚠️ OVER-PERMISSIONED]"
}

func RenderASCII(w io.Writer, g Graph) {
	for _, agent := range g.Agents {
		fmt.Fprintf(w, "Agent: %s\n", agent.Name)
		fmt.Fprintf(w, "├── Prompt: %s\n", agent.Prompt)
		fmt.Fprintln(w, "└── Tools:")
		for i, tool := range agent.Tools {
			connector := "├──"
			if i == len(agent.Tools)-1 {
				connector = "└──"
			}
			fmt.Fprintf(w, "    %s %s/%s %s\n", connector, tool.Server, tool.Name, tool.RiskTag)
		}
		fmt.Fprintln(w)
	}
}

func RenderMermaid(w io.Writer, g Graph) {
	fmt.Fprintln(w, "```mermaid")
	fmt.Fprintln(w, "graph TD")
	for _, agent := range g.Agents {
		aID := strings.ReplaceAll(agent.Name, "-", "_")
		fmt.Fprintf(w, "  %s[%s] --> %s_prompt[%s]\n", aID, agent.Name, aID, agent.Prompt)
		for _, tool := range agent.Tools {
			tID := strings.ReplaceAll(tool.Name, "-", "_")
			fmt.Fprintf(w, "  %s --> %s[%s/%s]\n", aID, tID, tool.Server, tool.Name)
		}
	}
	fmt.Fprintln(w, "```")
}
