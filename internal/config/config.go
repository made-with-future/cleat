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
	Enabled *bool    `yaml:"enabled,omitempty"`
	Service string   `yaml:"service"`
	Scripts []string `yaml:"scripts"`
}

type PythonConfig struct {
	Enabled        *bool  `yaml:"enabled,omitempty"`
	Django         bool   `yaml:"django"`
	DjangoService  string `yaml:"django_service"`
	PackageManager string `yaml:"package_manager"`
}

type GCPConfig struct {
	ProjectName string `yaml:"project_name"`
	Account     string `yaml:"account,omitempty"`
}

type TerraformConfig struct {
	UseFolders bool     `yaml:"use_folders"`
	Envs       []string `yaml:"envs,omitempty"`
}

type ModuleConfig struct {
	Python *PythonConfig `yaml:"python,omitempty"`
	Npm    *NpmConfig    `yaml:"npm,omitempty"`
}

type ServiceConfig struct {
	Name    string         `yaml:"name"`
	Dir     string         `yaml:"dir"`
	Docker  *bool          `yaml:"docker,omitempty"`
	Modules []ModuleConfig `yaml:"modules"`
}

func (s *ServiceConfig) IsDocker() bool {
	if s == nil {
		return false
	}
	return s.Docker != nil && *s.Docker
}

func (p *PythonConfig) IsEnabled() bool {
	if p == nil {
		return false
	}
	return p.Enabled == nil || *p.Enabled
}

func (n *NpmConfig) IsEnabled() bool {
	if n == nil {
		return false
	}
	return n.Enabled == nil || *n.Enabled
}

func ptrBool(b bool) *bool {
	return &b
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

// LoadDefaultConfig loads cleat.yaml from the current directory.
// If the file is not found, it returns a default config with auto-detection enabled.
func LoadDefaultConfig() (*Config, error) {
	cfg, err := LoadConfig("cleat.yaml")
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}
	return cfg, nil
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

	// Auto-detect Terraform
	iacDir := filepath.Join(baseDir, ".iac")
	if info, err := os.Stat(iacDir); err == nil && info.IsDir() {
		if cfg.Terraform == nil {
			cfg.Terraform = &TerraformConfig{}
		}

		// Check for subdirectories (multiple envs) or .tf files (single env)
		entries, err := os.ReadDir(iacDir)
		if err == nil {
			useFolders := false
			detectedEnvs := []string{}
			hasTfFiles := false

			for _, entry := range entries {
				if entry.IsDir() {
					// Check if subdirectory contains .tf files
					subDir := filepath.Join(iacDir, entry.Name())
					subEntries, _ := os.ReadDir(subDir)
					for _, subEntry := range subEntries {
						if !subEntry.IsDir() && strings.HasSuffix(subEntry.Name(), ".tf") {
							useFolders = true
							detectedEnvs = append(detectedEnvs, entry.Name())
							break
						}
					}
				} else if strings.HasSuffix(entry.Name(), ".tf") {
					hasTfFiles = true
				}
			}

			if useFolders {
				cfg.Terraform.UseFolders = true
				if cfg.Terraform.Envs == nil {
					cfg.Terraform.Envs = detectedEnvs
				}
				for _, env := range detectedEnvs {
					found := false
					for _, existing := range cfg.Envs {
						if existing == env {
							found = true
							break
						}
					}
					if !found {
						cfg.Envs = append(cfg.Envs, env)
					}
				}
			} else if hasTfFiles {
				cfg.Terraform.UseFolders = false
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

	// Auto-detect Docker and Services from docker-compose
	dockerComposeFile := ""
	if _, err := os.Stat(filepath.Join(baseDir, "docker-compose.yaml")); err == nil {
		dockerComposeFile = "docker-compose.yaml"
	} else if _, err := os.Stat(filepath.Join(baseDir, "docker-compose.yml")); err == nil {
		dockerComposeFile = "docker-compose.yml"
	}

	imageOnlyServices := make(map[string]bool)
	if dockerComposeFile != "" {
		cfg.Docker = true
		dcPath := filepath.Join(baseDir, dockerComposeFile)
		if dcData, err := os.ReadFile(dcPath); err == nil {
			type dcService struct {
				Build interface{} `yaml:"build"`
			}
			var dc struct {
				Services map[string]dcService `yaml:"services"`
			}
			if err := yaml.Unmarshal(dcData, &dc); err == nil {
				for name, s := range dc.Services {
					buildContext := ""
					if s.Build != nil {
						if b, ok := s.Build.(string); ok {
							buildContext = b
						} else if b, ok := s.Build.(map[string]interface{}); ok {
							if context, ok := b["context"].(string); ok {
								buildContext = context
							}
						}
					}

					if buildContext == "" {
						imageOnlyServices[name] = true
					}

					found := false
					for i := range cfg.Services {
						if cfg.Services[i].Name == name {
							if cfg.Services[i].Docker == nil {
								cfg.Services[i].Docker = ptrBool(true)
							}
							if cfg.Services[i].Dir == "" && buildContext != "" {
								cfg.Services[i].Dir = buildContext
							}
							found = true
							break
						}
					}
					if !found {
						cfg.Services = append(cfg.Services, ServiceConfig{
							Name:   name,
							Docker: ptrBool(true),
							Dir:    buildContext,
						})
					}
				}
			}
		}
	}

	// Apply defaults and auto-detection for each service and its modules
	for i := range cfg.Services {
		svc := &cfg.Services[i]

		// Track which modules are explicitly configured
		var explicitPython *PythonConfig
		var explicitNpm *NpmConfig
		for _, m := range svc.Modules {
			if m.Python != nil {
				explicitPython = m.Python
			}
			if m.Npm != nil {
				explicitNpm = m.Npm
			}
		}

		searchDir := baseDir
		if svc.Dir != "" {
			searchDir = filepath.Join(baseDir, svc.Dir)
		} else if imageOnlyServices[svc.Name] {
			searchDir = ""
		}

		// Auto-detect modules only if not explicitly configured
		if searchDir != "" {
			// Auto-detect Python/Django if not explicitly configured
			if explicitPython == nil {
				if _, err := os.Stat(filepath.Join(searchDir, "manage.py")); err == nil {
					svc.Modules = append(svc.Modules, ModuleConfig{Python: &PythonConfig{Django: true}})
				} else if _, err := os.Stat(filepath.Join(searchDir, "backend/manage.py")); err == nil {
					svc.Modules = append(svc.Modules, ModuleConfig{Python: &PythonConfig{Django: true}})
				}
			}

			// Auto-detect NPM if not explicitly configured
			if explicitNpm == nil {
				if _, err := os.Stat(filepath.Join(searchDir, "package.json")); err == nil {
					svc.Modules = append(svc.Modules, ModuleConfig{Npm: &NpmConfig{}})
				} else if _, err := os.Stat(filepath.Join(searchDir, "frontend/package.json")); err == nil {
					svc.Modules = append(svc.Modules, ModuleConfig{Npm: &NpmConfig{}})
				}
			}

			// Auto-detect Docker for service
			if svc.Docker == nil {
				if _, err := os.Stat(filepath.Join(searchDir, "docker-compose.yaml")); err == nil {
					svc.Docker = ptrBool(true)
				} else if _, err := os.Stat(filepath.Join(searchDir, "docker-compose.yml")); err == nil {
					svc.Docker = ptrBool(true)
				}
			}
		}

		// Apply defaults to modules (skip disabled ones)
		for j := range svc.Modules {
			mod := &svc.Modules[j]

			if mod.Python != nil && mod.Python.IsEnabled() {
				if mod.Python.DjangoService == "" {
					if svc.Name == "default" || svc.Name == "" {
						mod.Python.DjangoService = "backend"
					} else {
						mod.Python.DjangoService = svc.Name
					}
				}
				if mod.Python.PackageManager == "" {
					mod.Python.PackageManager = "uv"
				}
			}

			if mod.Npm != nil && mod.Npm.IsEnabled() {
				if len(mod.Npm.Scripts) == 0 && searchDir != "" {
					if _, err := os.Stat(filepath.Join(searchDir, "frontend/package.json")); err == nil {
						mod.Npm.Scripts = []string{"build"}
					} else if _, err := os.Stat(filepath.Join(searchDir, "package.json")); err == nil {
						mod.Npm.Scripts = []string{"build"}
					}
				}

				if mod.Npm.Service == "" {
					if cfg.Docker {
						if svc.Name == "default" || svc.Name == "" {
							mod.Npm.Service = "backend-node"
						} else {
							mod.Npm.Service = svc.Name
						}
					}
				}
			}
		}
	}

	return &cfg, nil
}
