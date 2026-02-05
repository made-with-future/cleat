package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
)

func init() {
	Register(&DjangoDetector{})
}

type DjangoDetector struct{}

func (d *DjangoDetector) Detect(baseDir string, cfg *config.Config) error {
	// Group services by directory for smarter auto-detection
	servicesByDir := make(map[string][]*config.ServiceConfig)
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
		hasManagePy := false
		if _, err := os.Stat(filepath.Join(searchDir, "manage.py")); err == nil {
			hasManagePy = true
		} else if _, err := os.Stat(filepath.Join(searchDir, "backend/manage.py")); err == nil {
			hasManagePy = true
		}

		if hasManagePy {
			var matches []*config.ServiceConfig
			var others []*config.ServiceConfig
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
					s.Modules = append(s.Modules, config.ModuleConfig{Python: &config.PythonConfig{Django: true}})
				}
			} else if len(others) > 0 {
				for _, s := range others {
					s.Modules = append(s.Modules, config.ModuleConfig{Python: &config.PythonConfig{Django: true}})
				}
			}
		}
	}

	// Apply defaults
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		for j := range svc.Modules {
			mod := &svc.Modules[j]
			if mod.Python != nil && mod.Python.IsEnabled() {
				if mod.Python.DjangoService == "" {
					if svc.Name == "default" || svc.Name == "" {
						// This part depends on knowing if it's a docker project and other service names.
						// For now, let's just use "backend" or svc.Name
						mod.Python.DjangoService = "backend"
					} else {
						mod.Python.DjangoService = svc.Name
					}
				}
				if mod.Python.PackageManager == "" {
					mod.Python.PackageManager = "uv"
				}
			}
		}
	}

	return nil
}

func matchesPython(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "python") || strings.Contains(name, "django") || strings.Contains(name, "backend") || strings.Contains(name, "api")
}
