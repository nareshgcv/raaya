package fixer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"raaya/pkg/analysis/static"
)

func ApplyFixes(workspace string, findings []static.Finding) (int, error) {
	fixedCount := 0
	secretsExtracted := make(map[string]string)

	for _, f := range findings {
		if !f.IsFixable {
			continue
		}

		switch f.RuleID {
		case "RAA004": // Plaintext Secrets
			count, err := fixPlaintextSecret(f, secretsExtracted)
			if err == nil && count > 0 {
				fixedCount += count
			}
		}
	}

	if len(secretsExtracted) > 0 {
		if err := appendToEnvFiles(workspace, secretsExtracted); err != nil {
			return fixedCount, fmt.Errorf("failed updating .env files: %w", err)
		}
	}

	return fixedCount, nil
}

func fixPlaintextSecret(f static.Finding, secretMap map[string]string) (int, error) {
	content, err := os.ReadFile(f.FilePath)
	if err != nil {
		return 0, err
	}

	secretRegex := regexp.MustCompile(`(["'])(sk-[a-zA-Z0-9]{20,}|[a-zA-Z0-9_-]{32,})(["'])`)
	loc := secretRegex.FindStringSubmatchIndex(string(content))
	if len(loc) < 6 {
		return 0, nil
	}

	rawSecret := string(content[loc[4]:loc[5]])
	envKey := "RAAYA_EXTRACTED_SECRET"

	updatedContent := secretRegex.ReplaceAllString(string(content), fmt.Sprintf(`"${%s}"`, envKey))
	if err := os.WriteFile(f.FilePath, []byte(updatedContent), 0644); err != nil {
		return 0, err
	}

	secretMap[envKey] = rawSecret
	return 1, nil
}

func appendToEnvFiles(workspace string, secrets map[string]string) error {
	envPath := filepath.Join(workspace, ".env")
	envExamplePath := filepath.Join(workspace, ".env.example")

	fEnv, err := os.OpenFile(envPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer fEnv.Close()

	fExample, err := os.OpenFile(envExamplePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fExample.Close()

	for key, val := range secrets {
		fmt.Fprintf(fEnv, "%s=%s\n", key, val)
		fmt.Fprintf(fExample, "%s=your_%s_here\n", key, strings.ToLower(key))
	}

	return nil
}
