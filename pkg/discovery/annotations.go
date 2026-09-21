package discovery

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/raaya/pkg/graph"
)

var (
	toolDecoratorRegex = regexp.MustCompile(`@(tool|mcp_tool)\((?:name=["']([^"']+)["'])?`)
	funcDefRegex       = regexp.MustCompile(`def\s+([a-zA-Z0-9_]+)\(`)
)

func ScanDirectoryAnnotations(root string, sg *graph.SecurityGraph) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		if strings.HasSuffix(path, ".py") || strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".js") {
			return parseFileForTools(path, sg)
		}
		return nil
	})
}

func parseFileForTools(filePath string, sg *graph.SecurityGraph) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if matches := toolDecoratorRegex.FindStringSubmatch(line); len(matches) > 0 {
			toolName := matches[2]
			if toolName == "" && scanner.Scan() {
				lineNum++
				funcLine := scanner.Text()
				if funcMatches := funcDefRegex.FindStringSubmatch(funcLine); len(funcMatches) > 1 {
					toolName = funcMatches[1]
				}
			}

			if toolName != "" {
				nodeID := "tool:" + toolName
				sg.AddNode(&graph.Node{
					ID:   nodeID,
					Type: graph.NodeTool,
					Name: toolName,
					Location: &graph.SourceLocation{
						FilePath:  filePath,
						StartLine: lineNum,
					},
				})
			}
		}
	}
	return scanner.Err()
}
