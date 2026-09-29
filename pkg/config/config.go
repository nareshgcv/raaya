package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ProjectName string   `yaml:"project_name"`
	RulesPath   string   `yaml:"rules_path"`
	Excludes    []string `yaml:"excludes"`
	Severity    string   `yaml:"severity"`
}

// LoadConfig parses a .raaya.yaml file if present in the workspace.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		ProjectName: "raaya-project",
		Severity:    "MEDIUM",
		Excludes:    []string{"vendor/", ".git/"},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return default configuration if missing
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return cfg, nil
}
