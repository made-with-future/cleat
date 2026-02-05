package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/madewithfuture/cleat/internal/config/schema"
	"github.com/madewithfuture/cleat/internal/detector"
	"gopkg.in/yaml.v3"
)

const LatestVersion = 1

// Re-export types from schema for backward compatibility if possible, 
// or update all usages in the project. Since I'm refactoring, 
// I'll use type aliases or just update the usages.
// For now, let's use type aliases for the most common ones.

type Config = schema.Config
type ServiceConfig = schema.ServiceConfig
type ModuleConfig = schema.ModuleConfig
type PythonConfig = schema.PythonConfig
type NpmConfig = schema.NpmConfig
type GCPConfig = schema.GCPConfig
type TerraformConfig = schema.TerraformConfig
type Workflow = schema.Workflow

var transientInputs = make(map[string]string)

// SetTransientInputs sets inputs that will be merged into all future loaded configs
func SetTransientInputs(inputs map[string]string) {
	for k, v := range inputs {
		transientInputs[k] = v
	}
}

// FindProjectRoot searches upwards from the current directory for a cleat.yaml/cleat.yml file or a .git directory.
func FindProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	curr := cwd
	for {
		// Check for cleat.yaml or cleat.yml
		if _, err := os.Stat(filepath.Join(curr, "cleat.yaml")); err == nil {
			return curr
		}
		if _, err := os.Stat(filepath.Join(curr, "cleat.yml")); err == nil {
			return curr
		}
		// Check for .git
		if _, err := os.Stat(filepath.Join(curr, ".git")); err == nil {
			return curr
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	return cwd
}

// LoadDefaultConfig searches upwards for cleat.yaml/cleat.yml and loads it.
// If the file is not found, it returns a default config with auto-detection enabled.
func LoadDefaultConfig() (*Config, error) {
	cwd, _ := os.Getwd()
	curr := cwd
	for {
		path := filepath.Join(curr, "cleat.yaml")
		if _, err := os.Stat(path); err == nil {
			return LoadConfig(path)
		}
		path = filepath.Join(curr, "cleat.yml")
		if _, err := os.Stat(path); err == nil {
			return LoadConfig(path)
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	return LoadConfig("cleat.yaml")
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			data = []byte{}
		} else {
			return nil, err
		}
	}

	var cfg Config
	cfg.SourcePath, _ = filepath.Abs(path)
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Version == 0 {
		cfg.Version = LatestVersion
	}

	if cfg.Version > LatestVersion || cfg.Version < 1 {
		return nil, fmt.Errorf("unrecognized configuration version: %d", cfg.Version)
	}

	if cfg.Envs != nil && len(cfg.Envs) == 0 {
		return nil, fmt.Errorf("envs must have at least one item if provided")
	}

	cfg.Inputs = make(map[string]string)
	for k, v := range transientInputs {
		cfg.Inputs[k] = v
	}

	baseDir := filepath.Dir(path)
	if err := detector.DetectAll(baseDir, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}