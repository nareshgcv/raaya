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
	pendingTool := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if matches := toolDecoratorRegex.FindStringSubmatch(line); len(matches) > 0 {
			if len(matches) > 2 && matches[2] != "" {
				addToolNode(matches[2], filePath, lineNum, sg)
			} else {
				pendingTool = true // Look for function definition on next line(s)
			}
			continue
		}

		if pendingTool {
			if funcMatches := funcDefRegex.FindStringSubmatch(line); len(funcMatches) > 1 {
				addToolNode(funcMatches[1], filePath, lineNum, sg)
				pendingTool = false
			}
		}
	}
	return scanner.Err()
}

func addToolNode(name, path string, line int, sg *graph.SecurityGraph) {
	nodeID := "tool:" + name
	sg.AddNode(&graph.Node{
		ID:   nodeID,
		Type: graph.NodeTool,
		Name: name,
		Location: &graph.SourceLocation{
			FilePath:  path,
			StartLine: line,
		},
	})
}
