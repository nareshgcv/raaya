package discovery

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/raaya/pkg/graph"
)

type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

type MCPConfigFile struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

func ParseMCPConfig(filePath string, sg *graph.SecurityGraph) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read mcp config: %w", err)
	}

	var config MCPConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse mcp config json: %w", err)
	}

	for serverName, serverCfg := range config.MCPServers {
		serverNodeID := fmt.Sprintf("server:%s", serverName)
		
		sg.AddNode(&graph.Node{
			ID:   serverNodeID,
			Type: graph.NodeServer,
			Name: serverName,
			Metadata: map[string]interface{}{
				"command": serverCfg.Command,
				"args":    serverCfg.Args,
			},
			Location: &graph.SourceLocation{
				FilePath: filePath,
			},
		})
	}

	return nil
}
