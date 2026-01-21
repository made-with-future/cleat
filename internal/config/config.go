package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const LatestVersion = 1

type NpmConfig struct {
	Service string   `yaml:"service"`
	Scripts []string `yaml:"scripts"`
}

type PythonConfig struct {
	Django         bool   `yaml:"django"`
	DjangoService  string `yaml:"django_service"`
	PackageManager string `yaml:"package_manager"`
}

type GCPConfig struct {
	ProjectName string `yaml:"project_name"`
	Account     string `yaml:"account,omitempty"`
}

type TerraformConfig struct {
}

type ModuleConfig struct {
	Python *PythonConfig `yaml:"python,omitempty"`
	Npm    *NpmConfig    `yaml:"npm,omitempty"`
}

type ServiceConfig struct {
	Name    string         `yaml:"name"`
	Dir     string         `yaml:"dir"`
	Docker  bool           `yaml:"docker"`
	Modules []ModuleConfig `yaml:"modules"`
}

type Config struct {
	Version             int              `yaml:"version"`
	Docker              bool             `yaml:"docker"`
	GoogleCloudPlatform *GCPConfig       `yaml:"google_cloud_platform,omitempty"`
	Terraform           *TerraformConfig `yaml:"terraform,omitempty"`
	Envs                []string         `yaml:"envs,omitempty"`
	Services            []ServiceConfig  `yaml:"services"`

	// Inputs stores transient values collected during execution
	Inputs map[string]string `yaml:"-"`
}

var transientInputs = make(map[string]string)

// SetTransientInputs sets inputs that will be merged into all future loaded configs
func SetTransientInputs(inputs map[string]string) {
	for k, v := range inputs {
		transientInputs[k] = v
	}
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

	if cfg.Version == 0 {
		cfg.Version = LatestVersion
	}

	if cfg.Version > LatestVersion || cfg.Version < 1 {
		return nil, fmt.Errorf("unrecognized configuration version: %d", cfg.Version)
	}

	if cfg.Envs == nil {
		// Auto-detect envs from .envs/*.env
		envsDir := filepath.Join(baseDir, ".envs")
		if entries, err := os.ReadDir(envsDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".env") {
					envName := strings.TrimSuffix(entry.Name(), ".env")
					cfg.Envs = append(cfg.Envs, envName)
				}
			}
		}
	}

	if cfg.Envs != nil && len(cfg.Envs) == 0 {
		return nil, fmt.Errorf("envs must have at least one item if provided")
	}

	cfg.Inputs = make(map[string]string)
	for k, v := range transientInputs {
		cfg.Inputs[k] = v
	}

	// Apply defaults and auto-detection for each service and its modules
	for i := range cfg.Services {
		svc := &cfg.Services[i]

		// Auto-detect modules
		hasPython := false
		hasNpm := false
		for _, m := range svc.Modules {
			if m.Python != nil {
				hasPython = true
			}
			if m.Npm != nil {
				hasNpm = true
			}
		}

		searchDir := baseDir
		if svc.Dir != "" {
			searchDir = filepath.Join(baseDir, svc.Dir)
		}

		if !hasPython {
			// Check for Django
			if _, err := os.Stat(filepath.Join(searchDir, "manage.py")); err == nil {
				svc.Modules = append(svc.Modules, ModuleConfig{Python: &PythonConfig{Django: true}})
			} else if _, err := os.Stat(filepath.Join(searchDir, "backend/manage.py")); err == nil {
				svc.Modules = append(svc.Modules, ModuleConfig{Python: &PythonConfig{Django: true}})
			}
		}

		if !hasNpm {
			// Check for NPM
			if _, err := os.Stat(filepath.Join(searchDir, "package.json")); err == nil {
				svc.Modules = append(svc.Modules, ModuleConfig{Npm: &NpmConfig{}})
			} else if _, err := os.Stat(filepath.Join(searchDir, "frontend/package.json")); err == nil {
				svc.Modules = append(svc.Modules, ModuleConfig{Npm: &NpmConfig{}})
			}
		}

		// Auto-detect Docker for service
		if !svc.Docker {
			if _, err := os.Stat(filepath.Join(searchDir, "docker-compose.yaml")); err == nil {
				svc.Docker = true
			}
		}

		for j := range svc.Modules {
			mod := &svc.Modules[j]

			if mod.Python != nil {
				if mod.Python.DjangoService == "" {
					mod.Python.DjangoService = "backend"
				}
				if mod.Python.PackageManager == "" {
					mod.Python.PackageManager = "uv"
				}
			}

			if mod.Npm != nil {
				if len(mod.Npm.Scripts) == 0 {
					searchDir := baseDir
					if svc.Dir != "" {
						searchDir = filepath.Join(baseDir, svc.Dir)
					}
					if _, err := os.Stat(filepath.Join(searchDir, "frontend/package.json")); err == nil {
						mod.Npm.Scripts = []string{"build"}
					} else if _, err := os.Stat(filepath.Join(searchDir, "package.json")); err == nil {
						mod.Npm.Scripts = []string{"build"}
					}
				}

				if mod.Npm.Service == "" {
					if cfg.Docker {
						mod.Npm.Service = "backend-node"
					}
				}
			}
		}
	}

	// Global auto-detection
	if !cfg.Docker {
		if _, err := os.Stat(filepath.Join(baseDir, "docker-compose.yaml")); err == nil {
			cfg.Docker = true
		}
	}

	return &cfg, nil
}
