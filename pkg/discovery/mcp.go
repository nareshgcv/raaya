package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"raaya/pkg/graph"
)

type MCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

type MCPServerConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Env     map[string]string `json:"env"`
	Tools   []MCPToolConfig   `json:"tools,omitempty"`
}

type MCPToolConfig struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Resources   []string `json:"resources,omitempty"`
}

type MCPScanner struct{}

func NewMCPScanner() *MCPScanner {
	return &MCPScanner{}
}

// Scan looks for MCP server configuration manifests in the project root.
func (s *MCPScanner) Scan(projectRoot string) ([]*graph.Node, []graph.Edge, error) {
	configPath := filepath.Join(projectRoot, "mcp.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Fallback check for alternative desktop/project manifest names
		configPath = filepath.Join(projectRoot, "claude_desktop_config.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return nil, nil, nil // No MCP config found; non-fatal
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read MCP config: %w", err)
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, nil, fmt.Errorf("failed to parse MCP config JSON: %w", err)
	}

	var nodes []*graph.Node
	var edges []graph.Edge

	// 1. Root Agent Node
	agentNode := &graph.Node{
		ID:   "agent:primary",
		Name: "Primary AI Agent",
		Type: graph.NodeAgent,
	}
	nodes = append(nodes, agentNode)

	// 2. Walk Servers, Tools, and Resources
	for serverName, serverCfg := range config.MCPServers {
		serverID := fmt.Sprintf("server:%s", serverName)
		serverNode := &graph.Node{
			ID:   serverID,
			Name: serverName,
			Type: graph.NodeMCPServer,
			Metadata: map[string]string{
				"command": serverCfg.Command,
			},
		}
		nodes = append(nodes, serverNode)

		// Edge: Agent -> MCP Server
		edges = append(edges, graph.Edge{
			SourceID:   agentNode.ID,
			TargetID:   serverID,
			Capability: "connect",
			Permission: graph.PermExecute,
		})

		for _, tool := range serverCfg.Tools {
			toolID := fmt.Sprintf("tool:%s:%s", serverName, tool.Name)
			toolNode := &graph.Node{
				ID:          toolID,
				Name:        tool.Name,
				Type:        graph.NodeTool,
				Permissions: []graph.PermissionLevel{graph.PermExecute},
			}
			nodes = append(nodes, toolNode)

			// Edge: MCP Server -> Tool
			edges = append(edges, graph.Edge{
				SourceID:   serverID,
				TargetID:   toolID,
				Capability: "exposes",
				Permission: graph.PermExecute,
			})

			for _, res := range tool.Resources {
				resID := fmt.Sprintf("resource:%s", res)
				resNode := &graph.Node{
					ID:   resID,
					Name: res,
					Type: graph.NodeResource,
				}
				nodes = append(nodes, resNode)

				// Edge: Tool -> Resource
				edges = append(edges, graph.Edge{
					SourceID:   toolID,
					TargetID:   resID,
					Capability: "accesses",
					Permission: graph.PermRead,
				})
			}
		}
	}

	return nodes, edges, nil
}
