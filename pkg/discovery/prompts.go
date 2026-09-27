package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"raaya/pkg/graph"
)

var modelPattern = regexp.MustCompile(`(?i)(gpt-4o|gpt-3\.5-turbo|claude-3-5-sonnet|claude-3-haiku|gemini-1\.5-pro)`)

func ScanPrompts(rootDir string, sg *graph.SecurityGraph) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		name := filepath.Base(path)
		if name == "system_prompt.txt" || name == "prompt.md" || filepath.Ext(path) == ".prompt" {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			promptID := fmt.Sprintf("prompt:%s", path)
			meta := make(map[string]string)

			if match := modelPattern.FindString(string(data)); match != "" {
				meta["detected_model"] = match
			}

			sg.AddNode(graph.AssetNode{
				ID:         promptID,
				Kind:       graph.KindAgentPrompt,
				Name:       name,
				SourceFile: path,
				Metadata:   meta,
			})
		}
		return nil
	})
}
