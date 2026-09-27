package fixer

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/raaya/pkg/graph"
)

// FixType defines the category of auto-remediation to execute.
type FixType string

const (
	FixExtractSecret FixType = "EXTRACT_SECRET"
	FixDisableTool   FixType = "DISABLE_TOOL"
)

// RemediationResult tracks the status of an applied fix.
type RemediationResult struct {
	FindingID string   `json:"finding_id"`
	FixType   FixType  `json:"fix_type"`
	FilePath  string   `json:"file_path"`
	Applied   bool     `json:"applied"`
	Message   string   `json:"message"`
	Diff      []string `json:"diff,omitempty"`
}

// Fixer manages workspace remediation actions.
type Fixer struct {
	WorkspaceRoot string
}

// NewFixer initializes a Fixer instance for a given workspace root.
func NewFixer(workspaceRoot string) *Fixer {
	return &Fixer{
		WorkspaceRoot: workspaceRoot,
	}
}

// ExtractSecretToEnv removes hardcoded secret values from a file and appends them to a .env file.
func (f *Fixer) ExtractSecretToEnv(findingID string, targetNode *graph.Node, secretKey string, secretVal string) (*RemediationResult, error) {
	if targetNode.Location == nil || targetNode.Location.FilePath == "" {
		return nil, fmt.Errorf("node %s lacks valid file location metadata", targetNode.ID)
	}

	targetPath := targetNode.Location.FilePath

	// 1. Append secret key and value to .env
	envPath := fmt.Sprintf("%s/.env", strings.TrimRight(f.WorkspaceRoot, "/"))
	envEntry := fmt.Sprintf("%s=%s\n", secretKey, secretVal)

	envFile, err := os.OpenFile(envPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open .env for writing: %w", err)
	}
	defer envFile.Close()

	if _, err := envFile.WriteString(envEntry); err != nil {
		return nil, fmt.Errorf("failed to append key to .env: %w", err)
	}

	// 2. Replace hardcoded secret in target file with environment variable reference
	if err := f.replaceInFile(targetPath, secretVal, fmt.Sprintf("${%s}", secretKey)); err != nil {
		return nil, fmt.Errorf("failed to sanitize source file %s: %w", targetPath, err)
	}

	return &RemediationResult{
		FindingID: findingID,
		FixType:   FixExtractSecret,
		FilePath:  targetPath,
		Applied:   true,
		Message:   fmt.Sprintf("Successfully extracted secret '%s' to .env and sanitized source code.", secretKey),
		Diff: []string{
			fmt.Sprintf("- %s", secretVal),
			fmt.Sprintf("+ ${%s}", secretKey),
		},
	}, nil
}

// replaceInFile helper replaces targeted content line-by-line.
func (f *Fixer) replaceInFile(filePath string, targetStr string, replacementStr string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, targetStr) {
			line = strings.ReplaceAll(line, targetStr, replacementStr)
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	output := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(filePath, []byte(output), 0644)
}
