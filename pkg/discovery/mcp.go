package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"raaya/pkg/graph"
)

type MCPConfig struct {
	MCPServers map[string]struct {
		Command string            `json:"command"`
		Args    []string          `json:"args"`
		Env     map[string]string `json:"env"`
		URL     string            `json:"url"`
	} `json:"mcpServers"`
}

var secretPattern = regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password)\s*=\s*['"]?([a-zA-Z0-9_\-]{16,})['"]?`)

func DiscoverMCPConfigs(filePath string, sg *graph.SecurityGraph) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	for serverName, serverInfo := range config.MCPServers {
		serverID := fmt.Sprintf("mcp_server:%s", serverName)
		sg.AddNode(graph.AssetNode{
			ID:         serverID,
			Kind:       graph.KindMCPServer,
			Name:       serverName,
			SourceFile: filePath,
			Metadata: map[string]string{
				"command": serverInfo.Command,
				"url":     serverInfo.URL,
			},
		})

		// Check environment variables for plain-text secrets
		for envKey, envVal := range serverInfo.Env {
			if secretPattern.MatchString(envKey) || secretPattern.MatchString(envVal) {
				secretID := fmt.Sprintf("secret:%s_%s", serverName, envKey)
				sg.AddNode(graph.AssetNode{
					ID:         secretID,
					Kind:       graph.KindSecret,
					Name:       envKey,
					SourceFile: filePath,
					Metadata: map[string]string{
						"value": envVal,
					},
				})
				sg.AddEdge(serverID, secretID, graph.RelUsesSecret)
			}
		}
	}
	return nil
}
