package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/madewithfuture/cleat/internal/config/schema"
)

type GoDetector struct{}

func (d *GoDetector) Detect(baseDir string, cfg *schema.Config) error {
	if len(cfg.Services) == 0 {
		if _, err := os.Stat(filepath.Join(baseDir, "go.mod")); err == nil {
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
		hasGoMod := false
		if _, err := os.Stat(filepath.Join(searchDir, "go.mod")); err == nil {
			hasGoMod = true
		}

		if hasGoMod {
			var matches []*schema.ServiceConfig
			var others []*schema.ServiceConfig
			for _, s := range svcs {
				explicit := false
				for _, m := range s.Modules {
					if m.Go != nil {
						explicit = true
						break
					}
				}
				if explicit {
					continue
				}

				if matchesGo(s.Name) {
					matches = append(matches, s)
				} else {
					others = append(others, s)
				}
			}

			if len(matches) > 0 {
				for _, s := range matches {
					s.Modules = append(s.Modules, schema.ModuleConfig{Go: &schema.GoConfig{}})
				}
			} else if len(others) > 0 {
				for _, s := range others {
					s.Modules = append(s.Modules, schema.ModuleConfig{Go: &schema.GoConfig{}})
				}
			}
		}
	}

	// set sensible defaults
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		for j := range svc.Modules {
			mod := &svc.Modules[j]
			if mod.Go != nil && mod.Go.IsEnabled() {
				if mod.Go.Service == "" {
					if svc.Name == "default" || svc.Name == "" {
						mod.Go.Service = "backend-go"
					} else {
						mod.Go.Service = svc.Name
					}
				}
			}
		}
	}
	return nil
}

func matchesGo(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "go") || strings.Contains(name, "golang") || strings.Contains(name, "backend") || strings.Contains(name, "api") || strings.Contains(name, "server") || strings.Contains(name, "cli")
}
