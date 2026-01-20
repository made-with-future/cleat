package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const LatestVersion = 2

type NpmConfig struct {
	Service string   `yaml:"service"`
	Scripts []string `yaml:"scripts"`
}

type PythonConfig struct {
	Django        bool   `yaml:"django"`
	DjangoService string `yaml:"django_service"`
}

type GCPConfig struct {
	ProjectName string `yaml:"project_name"`
	Account     string `yaml:"account,omitempty"`
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

	// Legacy fields for migration
	Python *PythonConfig `yaml:"python,omitempty"`
	Npm    *NpmConfig    `yaml:"npm,omitempty"`
}

type Config struct {
	Version             int             `yaml:"version"`
	Docker              bool            `yaml:"docker"`
	GoogleCloudPlatform *GCPConfig      `yaml:"google_cloud_platform,omitempty"`
	Services            []ServiceConfig `yaml:"services"`

	// Inputs stores transient values collected during execution
	Inputs map[string]string `yaml:"-"`

	// Legacy fields for V1
	Python *PythonConfig `yaml:"python,omitempty"`
	Npm    *NpmConfig    `yaml:"npm,omitempty"`
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

	cfg.Inputs = make(map[string]string)
	for k, v := range transientInputs {
		cfg.Inputs[k] = v
	}

	// Migrate root-level Python/Npm to Services structure
	if cfg.Python != nil || cfg.Npm != nil {
		svc := ServiceConfig{
			Name: "default",
		}
		if cfg.Python != nil {
			svc.Python = cfg.Python
			cfg.Python = nil
		}
		if cfg.Npm != nil {
			svc.Npm = cfg.Npm
			cfg.Npm = nil
		}
		cfg.Services = append(cfg.Services, svc)
	}

	// Migrate Service-level Python/Npm to Modules
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		if svc.Python != nil {
			svc.Modules = append(svc.Modules, ModuleConfig{Python: svc.Python})
			svc.Python = nil
		}
		if svc.Npm != nil {
			svc.Modules = append(svc.Modules, ModuleConfig{Npm: svc.Npm})
			svc.Npm = nil
		}
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
