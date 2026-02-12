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

				if matchesNpm(s.Name) {
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
					if svc.Name == "default" || svc.Name == "" {
						mod.Npm.Service = "backend-node"
					} else {
						mod.Npm.Service = svc.Name
					}
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

func matchesNpm(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "npm") || strings.Contains(name, "node") || strings.Contains(name, "frontend") || strings.Contains(name, "ui") || strings.Contains(name, "vite") || strings.Contains(name, "assets")
}
