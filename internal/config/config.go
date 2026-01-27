package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	AppYaml string         `yaml:"app_yaml,omitempty"`
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
	AppYaml             string           `yaml:"app_yaml,omitempty"`

	// Inputs stores transient values collected during execution
	Inputs map[string]string `yaml:"-"`

	// SourcePath is the absolute path to the loaded config file
	SourcePath string `yaml:"-"`
}

var transientInputs = make(map[string]string)

// SetTransientInputs sets inputs that will be merged into all future loaded configs
func SetTransientInputs(inputs map[string]string) {
	for k, v := range inputs {
		transientInputs[k] = v
	}
}

// FindProjectRoot searches upwards from the current directory for a cleat.yaml file or a .git directory.
func FindProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	curr := cwd
	for {
		// Check for cleat.yaml
		if _, err := os.Stat(filepath.Join(curr, "cleat.yaml")); err == nil {
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

// LoadDefaultConfig searches upwards for cleat.yaml and loads it.
// If the file is not found, it returns a default config with auto-detection enabled.
func LoadDefaultConfig() (*Config, error) {
	cwd, _ := os.Getwd()
	curr := cwd
	for {
		path := filepath.Join(curr, "cleat.yaml")
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

	baseDir := filepath.Dir(path)

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
	dcServices := make(map[string]bool)
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
					dcServices[name] = true
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

	// Group services by directory for smarter auto-detection
	servicesByDir := make(map[string][]*ServiceConfig)
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		searchDir := baseDir
		if svc.Dir != "" {
			searchDir = filepath.Join(baseDir, svc.Dir)
		} else if imageOnlyServices[svc.Name] {
			searchDir = ""
		}
		if searchDir != "" {
			servicesByDir[searchDir] = append(servicesByDir[searchDir], svc)
		}
	}

	// Auto-detect modules for each directory
	for searchDir, svcs := range servicesByDir {
		hasManagePy := false
		if _, err := os.Stat(filepath.Join(searchDir, "manage.py")); err == nil {
			hasManagePy = true
		} else if _, err := os.Stat(filepath.Join(searchDir, "backend/manage.py")); err == nil {
			hasManagePy = true
		}

		hasPackageJson := false
		if _, err := os.Stat(filepath.Join(searchDir, "package.json")); err == nil {
			hasPackageJson = true
		}

		if hasManagePy {
			// Find service(s) for Python
			var matches []*ServiceConfig
			var others []*ServiceConfig
			for _, s := range svcs {
				explicit := false
				for _, m := range s.Modules {
					if m.Python != nil {
						explicit = true
						break
					}
				}
				if explicit {
					continue
				}

				if matchesPython(s.Name) {
					matches = append(matches, s)
				} else {
					others = append(others, s)
				}
			}

			if len(matches) > 0 {
				for _, s := range matches {
					s.Modules = append(s.Modules, ModuleConfig{Python: &PythonConfig{Django: true}})
				}
			} else if len(others) > 0 {
				for _, s := range others {
					s.Modules = append(s.Modules, ModuleConfig{Python: &PythonConfig{Django: true}})
				}
			}
		}

		if hasPackageJson {
			// Find service(s) for NPM
			var matches []*ServiceConfig
			var others []*ServiceConfig
			for _, s := range svcs {
				explicit := false
				for _, m := range s.Modules {
					if m.Npm != nil {
						explicit = true
						break
					}
				}
				if explicit {
					continue
				}

				if matchesNpm(s.Name) {
					matches = append(matches, s)
				} else {
					others = append(others, s)
				}
			}

			if len(matches) > 0 {
				for _, s := range matches {
					s.Modules = append(s.Modules, ModuleConfig{Npm: &NpmConfig{}})
				}
			} else if len(others) > 0 {
				for _, s := range others {
					s.Modules = append(s.Modules, ModuleConfig{Npm: &NpmConfig{}})
				}
			}
		}
	}

	// Apply defaults and other auto-detections (Docker)
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		searchDir := baseDir
		if svc.Dir != "" {
			searchDir = filepath.Join(baseDir, svc.Dir)
		} else if imageOnlyServices[svc.Name] {
			searchDir = ""
		}

		// Auto-detect Docker for service
		if searchDir != "" && svc.Docker == nil {
			if _, err := os.Stat(filepath.Join(searchDir, "docker-compose.yaml")); err == nil {
				svc.Docker = ptrBool(true)
			} else if _, err := os.Stat(filepath.Join(searchDir, "docker-compose.yml")); err == nil {
				svc.Docker = ptrBool(true)
			}
		}

		// Apply defaults to modules (skip disabled ones)
		for j := range svc.Modules {
			mod := &svc.Modules[j]

			if mod.Python != nil && mod.Python.IsEnabled() {
				if mod.Python.DjangoService == "" {
					if svc.Name == "default" || svc.Name == "" {
						if cfg.Docker && len(dcServices) > 0 {
							if _, ok := dcServices["backend"]; ok {
								mod.Python.DjangoService = "backend"
							} else {
								mod.Python.DjangoService = svc.Name
							}
						} else {
							mod.Python.DjangoService = "backend"
						}
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
					packageJsonPath := filepath.Join(searchDir, "package.json")
					if _, err := os.Stat(packageJsonPath); err == nil {
						mod.Npm.Scripts = readNpmScripts(packageJsonPath)
					}
				}

				if mod.Npm.Service == "" {
					if cfg.Docker {
						if svc.Name == "default" || svc.Name == "" {
							if _, ok := dcServices["backend-node"]; ok {
								mod.Npm.Service = "backend-node"
							} else if _, ok := dcServices["frontend"]; ok {
								mod.Npm.Service = "frontend"
							} else {
								mod.Npm.Service = "backend-node" // Fallback to legacy default if no matches found
							}
						} else {
							mod.Npm.Service = svc.Name
						}
					}
				}
			}
		}
	}

	// Auto-detect GCP app.yaml
	if cfg.GoogleCloudPlatform != nil {
		if _, err := os.Stat(filepath.Join(baseDir, "app.yaml")); err == nil {
			if cfg.AppYaml == "" {
				cfg.AppYaml = "app.yaml"
			}
		}

		entries, err := os.ReadDir(baseDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					appYamlPath := filepath.Join(entry.Name(), "app.yaml")
					if _, err := os.Stat(filepath.Join(baseDir, appYamlPath)); err == nil {
						found := false
						for i := range cfg.Services {
							if cfg.Services[i].Dir == entry.Name() || cfg.Services[i].Name == entry.Name() {
								if cfg.Services[i].AppYaml == "" {
									cfg.Services[i].AppYaml = appYamlPath
								}
								found = true
								break
							}
						}
						if !found && entry.Name() != ".git" && entry.Name() != ".envs" && entry.Name() != ".iac" && entry.Name() != "terraform" {
							cfg.Services = append(cfg.Services, ServiceConfig{
								Name:    entry.Name(),
								Dir:     entry.Name(),
								AppYaml: appYamlPath,
							})
						}
					}
				}
			}
		}
	}

	return &cfg, nil
}

type packageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

func readNpmScripts(packageJsonPath string) []string {
	data, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return nil
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}

	scripts := make([]string, 0, len(pkg.Scripts))
	for s := range pkg.Scripts {
		scripts = append(scripts, s)
	}
	sort.Strings(scripts)
	return scripts
}

func matchesPython(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "python") || strings.Contains(name, "django") || strings.Contains(name, "backend") || strings.Contains(name, "api")
}

func matchesNpm(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "npm") || strings.Contains(name, "node") || strings.Contains(name, "frontend") || strings.Contains(name, "ui") || strings.Contains(name, "vite") || strings.Contains(name, "assets")
}
