package discovery

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"raaya/pkg/graph"
)

var (
	pyToolPattern = regexp.MustCompile(`@(?:mcp\.tool|tool)\((?:name=["']([^"']+)["'])?\)`)
	pyDefPattern  = regexp.MustCompile(`def\s+([a-zA-Z0-9_]+)\s*\(`)
	tsToolPattern = regexp.MustCompile(`(?:server|mcp)\.tool\s*\(\s*["']([^"']+)["']`)
)

// ScanAnnotations traverses a directory for Python/TypeScript files and extracts tool declarations
func ScanAnnotations(rootDir string, sg *graph.SecurityGraph) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && (info.Name() == "node_modules" || info.Name() == ".venv" || info.Name() == ".git") {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".py" || ext == ".ts" || ext == ".js" {
			return parseFileTools(path, ext, sg)
		}
		return nil
	})
}

func parseFileTools(filePath, ext string, sg *graph.SecurityGraph) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var lastDecoratorName string

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if ext == ".py" {
			if matches := pyToolPattern.FindStringSubmatch(line); len(matches) > 0 {
				if len(matches) > 1 && matches[1] != "" {
					lastDecoratorName = matches[1]
				} else {
					lastDecoratorName = "pending_function"
				}
				continue
			}

			if lastDecoratorName != "" {
				if defMatches := pyDefPattern.FindStringSubmatch(line); len(defMatches) > 1 {
					toolName := lastDecoratorName
					if toolName == "pending_function" {
						toolName = defMatches[1]
					}
					toolID := fmt.Sprintf("tool:%s:%s", filePath, toolName)
					sg.AddNode(graph.AssetNode{
						ID:         toolID,
						Kind:       graph.KindToolDef,
						Name:       toolName,
						SourceFile: filePath,
						LineNumber: lineNum,
					})
					lastDecoratorName = ""
				}
			}
		} else if ext == ".ts" || ext == ".js" {
			if matches := tsToolPattern.FindStringSubmatch(line); len(matches) > 1 {
				toolName := matches[1]
				toolID := fmt.Sprintf("tool:%s:%s", filePath, toolName)
				sg.AddNode(graph.AssetNode{
					ID:         toolID,
					Kind:       graph.KindToolDef,
					Name:       toolName,
					SourceFile: filePath,
					LineNumber: lineNum,
				})
			}
		}
	}
	return scanner.Err()
}
