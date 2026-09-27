package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	IgnoredRules map[string]bool
	IgnoredPaths []string
}

func LoadConfig(workspace string) (*Config, error) {
	cfg := &Config{
		IgnoredRules: make(map[string]bool),
		IgnoredPaths: []string{},
	}

	// Parse .raayaignore if present
	ignorePath := filepath.Join(workspace, ".raayaignore")
	if file, err := os.Open(ignorePath); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if strings.HasPrefix(line, "RAA") {
				cfg.IgnoredRules[line] = true
			} else {
				cfg.IgnoredPaths = append(cfg.IgnoredPaths, line)
			}
		}
	}

	return cfg, nil
}
