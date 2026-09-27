package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	IgnorePaths []string `yaml:"ignore_paths"`
	Rules       struct {
		Disabled []string `yaml:"disabled"`
	} `yaml:"rules"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return default empty config if file doesn't exist
		}
		return nil, err
	}
	err = yaml.Unmarshal(data, cfg)
	return cfg, err
}
