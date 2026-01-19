package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type NpmConfig struct {
	Service string   `yaml:"service"`
	Scripts []string `yaml:"scripts"`
}

type PythonConfig struct {
	Django        bool   `yaml:"django"`
	DjangoService string `yaml:"django_service"`
}

type Config struct {
	Docker bool         `yaml:"docker"`
	Python PythonConfig `yaml:"python"`
	Npm    NpmConfig    `yaml:"npm"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(path)

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Python.DjangoService == "" {
		cfg.Python.DjangoService = "backend"
	}

	if !cfg.Docker {
		if _, err := os.Stat(filepath.Join(baseDir, "docker-compose.yaml")); err == nil {
			cfg.Docker = true
		}
	}

	if len(cfg.Npm.Scripts) == 0 {
		if _, err := os.Stat(filepath.Join(baseDir, "frontend/package.json")); err == nil {
			cfg.Npm.Scripts = []string{"build"}
		}
	}

	if cfg.Npm.Service == "" {
		if cfg.Docker {
			cfg.Npm.Service = "backend-node"
		}
	}

	return &cfg, nil
}
