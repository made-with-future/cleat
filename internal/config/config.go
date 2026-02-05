package config

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/madewithfuture/cleat/internal/config/schema"
	"github.com/madewithfuture/cleat/internal/detector"
	"github.com/madewithfuture/cleat/internal/logger"
	"gopkg.in/yaml.v3"
)

const LatestVersion = 1

// Re-export types from schema
type Config = schema.Config
type ServiceConfig = schema.ServiceConfig
type ModuleConfig = schema.ModuleConfig
type PythonConfig = schema.PythonConfig
type NpmConfig = schema.NpmConfig
type GCPConfig = schema.GCPConfig
type TerraformConfig = schema.TerraformConfig
type Workflow = schema.Workflow

// FindProjectRoot searches upwards from the current directory for a cleat.yaml/cleat.yml file or a .git directory.
func FindProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("failed to get current working directory", err, nil)
		return "."
	}

	curr := cwd
	for {
		if _, err := os.Stat(filepath.Join(curr, "cleat.yaml")); err == nil {
			return curr
		}
		if _, err := os.Stat(filepath.Join(curr, "cleat.yml")); err == nil {
			return curr
		}
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

// GetProjectID returns a unique identifier for the current project based on its absolute path.
func GetProjectID() string {
	root := FindProjectRoot()
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "unknown"
	}

	hash := sha256.Sum256([]byte(absRoot))
	projectDirName := filepath.Base(absRoot)
	if projectDirName == "/" || projectDirName == "." || projectDirName == "" {
		projectDirName = "root"
	}

	// Use project directory name + 8 bytes of hash
	return fmt.Sprintf("%s-%x", projectDirName, hash[:8])
}

// LoadDefaultConfig searches upwards for cleat.yaml/cleat.yml and loads it.
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
			return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
		}
	}

	var cfg Config
	var absErr error
	cfg.SourcePath, absErr = filepath.Abs(path)
	if absErr != nil {
		logger.Warn("failed to get absolute path for config", map[string]interface{}{"path": path, "error": absErr.Error()})
		cfg.SourcePath = path
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config from %s: %w", path, err)
	}

	if cfg.Version == 0 {
		cfg.Version = LatestVersion
	}

	if cfg.Version > LatestVersion || cfg.Version < 1 {
		return nil, fmt.Errorf("unrecognized configuration version in %s: %d", path, cfg.Version)
	}

	if cfg.Envs != nil && len(cfg.Envs) == 0 {
		return nil, fmt.Errorf("envs in %s must have at least one item if provided", path)
	}

	cfg.Inputs = make(map[string]string)

	baseDir := filepath.Dir(path)
	if err := detector.DetectAll(baseDir, &cfg); err != nil {
		return nil, fmt.Errorf("auto-detection failed during config load of %s: %w", path, err)
	}

	return &cfg, nil
}
