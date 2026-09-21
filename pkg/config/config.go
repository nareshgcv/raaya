package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	IgnorePaths []string          `yaml:"ignore_paths"`
	Thresholds  ThresholdSettings `yaml:"thresholds"`
	Rules       RuleSettings     `yaml:"rules"`
}

type ThresholdSettings struct {
	MaxBlastRadius int `yaml:"max_blast_radius"`
}

type RuleSettings struct {
	Disabled []string `yaml:"disabled"`
}

func LoadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{
			Thresholds: ThresholdSettings{MaxBlastRadius: 5},
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
