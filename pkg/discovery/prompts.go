package discovery

import (
	"os"

	"github.com/raaya/pkg/graph"
)

func RegisterAgentPrompt(agentName, promptPath, model string, sg *graph.SecurityGraph) error {
	content, err := os.ReadFile(promptPath)
	if err != nil {
		return err
	}

	agentID := "agent:" + agentName
	sg.AddNode(&graph.Node{
		ID:   agentID,
		Type: graph.NodeAgent,
		Name: agentName,
		Metadata: map[string]interface{}{
			"prompt_length": len(content),
			"model":         model,
		},
		Location: &graph.SourceLocation{
			FilePath: promptPath,
		},
	})

	if model != "" {
		modelID := "model:" + model
		sg.AddNode(&graph.Node{
			ID:   modelID,
			Type: graph.NodeModel,
			Name: model,
		})

		sg.AddEdge(&graph.Edge{
			ID:       agentID + "->" + modelID,
			SourceID: agentID,
			TargetID: modelID,
			Type:     graph.EdgeUsesModel,
		})
	}

	return nil
}
