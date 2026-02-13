package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/madewithfuture/cleat/internal/config/schema"
)

type GoDetector struct{}

func (d *GoDetector) Detect(baseDir string, cfg *schema.Config) error {
	rootCovered := false
	for _, svc := range cfg.Services {
		if svc.Dir == "." || svc.Dir == "" {
			rootCovered = true
			break
		}
	}

	if !rootCovered {
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
		} else if searchDir == baseDir {
			// If not found in root, check if any service matches a subdirectory containing go.mod
			for _, s := range svcs {
				if (s.Dir == "." || s.Dir == "") && s.Name != "" {
					if _, err := os.Stat(filepath.Join(baseDir, s.Name, "go.mod")); err == nil {
						hasGoMod = true
						break
					}
				}
			}
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

				if matchesGo(s, searchDir) {
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
					mod.Go.Service = svc.Name
				}
			}
		}
	}
	return nil
}

func matchesGo(svc *schema.ServiceConfig, searchDir string) bool {
	if svc.Dockerfile != "" {
		dfPath := filepath.Join(searchDir, svc.Dockerfile)
		if data, err := os.ReadFile(dfPath); err == nil {
			content := strings.ToLower(string(data))
			if strings.Contains(content, "golang") || strings.Contains(content, "go.mod") || strings.Contains(content, "go build") || strings.Contains(content, "go run") {
				return true
			}
			if strings.Contains(content, "python") || strings.Contains(content, "node") || strings.Contains(content, "package.json") {
				return false
			}
		}
	}

	if svc.Command != "" {
		cmd := strings.ToLower(svc.Command)
		if strings.Contains(cmd, "go build") || strings.Contains(cmd, "go run") || strings.Contains(cmd, "go test") {
			return true
		}
		if strings.Contains(cmd, "python") || strings.Contains(cmd, "manage.py") || strings.Contains(cmd, "npm") || strings.Contains(cmd, "node") {
			return false
		}
	}

	if svc.Image != "" {
		img := strings.ToLower(svc.Image)
		if strings.Contains(img, "golang") || strings.Contains(img, "go:") {
			return true
		}
		if strings.Contains(img, "python") || strings.Contains(img, "node") || strings.Contains(img, "postgres") || strings.Contains(img, "redis") {
			return false
		}
	}

	name := strings.ToLower(svc.Name)
	if (strings.Contains(name, "python") || strings.Contains(name, "django") || strings.Contains(name, "node") || strings.Contains(name, "npm") || strings.Contains(name, "js")) && !strings.Contains(name, "go") && !strings.Contains(name, "golang") {
		return false
	}
	return strings.Contains(name, "go") || strings.Contains(name, "golang") || strings.Contains(name, "api") || strings.Contains(name, "server") || strings.Contains(name, "cli") || strings.Contains(name, "backend")
}
