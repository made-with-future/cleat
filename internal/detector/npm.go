package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/madewithfuture/cleat/internal/config/schema"
)

type NpmDetector struct{}

func (d *NpmDetector) Detect(baseDir string, cfg *schema.Config) error {
	rootCovered := false
	for _, svc := range cfg.Services {
		if svc.Dir == "." || svc.Dir == "" {
			rootCovered = true
			break
		}
	}

	if !rootCovered {
		if _, err := os.Stat(filepath.Join(baseDir, "package.json")); err == nil {
			cfg.Services = append(cfg.Services, schema.ServiceConfig{
				Name: "default",
				Dir:  ".",
			})
		}
	}

	servicesByDir := make(map[string][]*schema.ServiceConfig)
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		searchDir := baseDir
		if svc.Dir != "" {
			searchDir = filepath.Join(baseDir, svc.Dir)
		}
		if searchDir != "" {
			servicesByDir[searchDir] = append(servicesByDir[searchDir], svc)
		}
	}

	for searchDir, svcs := range servicesByDir {
		hasPackageJson := false
		if _, err := os.Stat(filepath.Join(searchDir, "package.json")); err == nil {
			hasPackageJson = true
		} else if searchDir == baseDir {
			// If not found in root, check if any service matches a subdirectory containing package.json
			for _, s := range svcs {
				if (s.Dir == "." || s.Dir == "") && s.Name != "" {
					if _, err := os.Stat(filepath.Join(baseDir, s.Name, "package.json")); err == nil {
						hasPackageJson = true
						break
					}
				}
			}
		}

		if hasPackageJson {
			var matches []*schema.ServiceConfig
			var others []*schema.ServiceConfig
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

				if matchesNpm(s, searchDir) {
					matches = append(matches, s)
				} else {
					others = append(others, s)
				}
			}

			if len(matches) > 0 {
				for _, s := range matches {
					s.Modules = append(s.Modules, schema.ModuleConfig{Npm: &schema.NpmConfig{}})
				}
			} else if len(others) > 0 {
				for _, s := range others {
					s.Modules = append(s.Modules, schema.ModuleConfig{Npm: &schema.NpmConfig{}})
				}
			}
		}
	}

	for i := range cfg.Services {
		svc := &cfg.Services[i]
		searchDir := baseDir
		if svc.Dir != "" {
			searchDir = filepath.Join(baseDir, svc.Dir)
		}

		for j := range svc.Modules {
			mod := &svc.Modules[j]
			if mod.Npm != nil && mod.Npm.IsEnabled() {
				if len(mod.Npm.Scripts) == 0 && searchDir != "" {
					packageJsonPath := filepath.Join(searchDir, "package.json")
					if _, err := os.Stat(packageJsonPath); err == nil {
						scripts, err := readNpmScripts(packageJsonPath)
						if err != nil {
							return err
						}
						mod.Npm.Scripts = scripts
					}
				}

				if mod.Npm.Service == "" {
					mod.Npm.Service = svc.Name
				}
			}
		}
	}

	return nil
}

type packageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

func readNpmScripts(packageJsonPath string) ([]string, error) {
	data, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return nil, err
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	scripts := make([]string, 0, len(pkg.Scripts))
	for s := range pkg.Scripts {
		scripts = append(scripts, s)
	}
	sort.Strings(scripts)
	return scripts, nil
}

func matchesNpm(svc *schema.ServiceConfig, searchDir string) bool {
	if svc.Dockerfile != "" {
		dfPath := filepath.Join(searchDir, svc.Dockerfile)
		if data, err := os.ReadFile(dfPath); err == nil {
			content := strings.ToLower(string(data))
			if strings.Contains(content, "node") || strings.Contains(content, "package.json") || strings.Contains(content, "npm") || strings.Contains(content, "yarn") || strings.Contains(content, "pnpm") || strings.Contains(content, "bun") {
				return true
			}
			// If it mentions other stacks but not node, it's probably NOT node
			if strings.Contains(content, "python") || strings.Contains(content, "manage.py") || strings.Contains(content, "go.mod") {
				return false
			}
		}
	}

	if svc.Command != "" {
		cmd := strings.ToLower(svc.Command)
		if strings.Contains(cmd, "npm") || strings.Contains(cmd, "node") || strings.Contains(cmd, "yarn") || strings.Contains(cmd, "pnpm") || strings.Contains(cmd, "bun") {
			return true
		}
		if strings.Contains(cmd, "python") || strings.Contains(cmd, "manage.py") || strings.Contains(cmd, "go build") || strings.Contains(cmd, "go run") {
			return false
		}
	}

	if svc.Image != "" {
		img := strings.ToLower(svc.Image)
		if strings.Contains(img, "node") {
			return true
		}
		if strings.Contains(img, "python") || strings.Contains(img, "golang") || strings.Contains(img, "postgres") || strings.Contains(img, "redis") {
			return false
		}
	}

	name := strings.ToLower(svc.Name)
	if (strings.Contains(name, "python") || strings.Contains(name, "django") || strings.Contains(name, "go") || strings.Contains(name, "golang")) && !strings.Contains(name, "node") && !strings.Contains(name, "npm") && !strings.Contains(name, "js") {
		return false
	}
	return strings.Contains(name, "npm") || strings.Contains(name, "node") || strings.Contains(name, "frontend") || strings.Contains(name, "ui") || strings.Contains(name, "vite") || strings.Contains(name, "assets") || strings.Contains(name, "backend")
}
