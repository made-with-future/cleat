package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/madewithfuture/cleat/internal/config/schema"
)

type DjangoDetector struct{}

func (d *DjangoDetector) Detect(baseDir string, cfg *schema.Config) error {
	rootCovered := false
	for _, svc := range cfg.Services {
		if svc.Dir == "." || svc.Dir == "" {
			rootCovered = true
			break
		}
	}

	if !rootCovered {
		if _, err := os.Stat(filepath.Join(baseDir, "manage.py")); err == nil {
			cfg.Services = append(cfg.Services, schema.ServiceConfig{
				Name: "default",
				Dir:  ".",
			})
		}
	}

	// Group services by directory for smarter auto-detection
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
		hasManagePy := false
		if _, err := os.Stat(filepath.Join(searchDir, "manage.py")); err == nil {
			hasManagePy = true
		} else if searchDir == baseDir {
			// If not found in root, check if any service matches a subdirectory containing manage.py
			for _, s := range svcs {
				if (s.Dir == "." || s.Dir == "") && s.Name != "" {
					if _, err := os.Stat(filepath.Join(baseDir, s.Name, "manage.py")); err == nil {
						hasManagePy = true
						break
					}
				}
			}
		}

		if hasManagePy {
			var matches []*schema.ServiceConfig
			var others []*schema.ServiceConfig
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

				if matchesPython(s, searchDir) {
					matches = append(matches, s)
				} else {
					others = append(others, s)
				}
			}

			if len(matches) > 0 {
				for _, s := range matches {
					s.Modules = append(s.Modules, schema.ModuleConfig{Python: &schema.PythonConfig{Django: true}})
				}
			} else if len(others) > 0 {
				for _, s := range others {
					s.Modules = append(s.Modules, schema.ModuleConfig{Python: &schema.PythonConfig{Django: true}})
				}
			}
		}
	}

	// Apply defaults
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		searchDir := baseDir
		if svc.Dir != "" {
			searchDir = filepath.Join(baseDir, svc.Dir)
		}

		for j := range svc.Modules {
			mod := &svc.Modules[j]
			if mod.Python != nil && mod.Python.IsEnabled() {
				if mod.Python.DjangoService == "" {
					mod.Python.DjangoService = svc.Name
				}
				if mod.Python.PackageManager == "" {
					mod.Python.PackageManager = detectPackageManager(searchDir, baseDir)
				}
			}
		}
	}

	return nil
}

func detectPackageManager(dir string, baseDir string) string {
	// 1. Check service root
	if pm := checkDirForPackageManager(dir); pm != "" {
		return pm
	}

	// 2. Check project root if different
	if dir != baseDir {
		if pm := checkDirForPackageManager(baseDir); pm != "" {
			return pm
		}
	}

	return "uv"
}

func checkDirForPackageManager(dir string) string {
	if _, err := os.Stat(filepath.Join(dir, "uv.lock")); err == nil {
		return "uv"
	}
	if _, err := os.Stat(filepath.Join(dir, "requirements.txt")); err == nil {
		return "pip"
	}
	if _, err := os.Stat(filepath.Join(dir, "poetry.lock")); err == nil {
		return "poetry"
	}
	return ""
}

func matchesPython(svc *schema.ServiceConfig, searchDir string) bool {
	if svc.Dockerfile != "" {
		dfPath := filepath.Join(searchDir, svc.Dockerfile)
		if data, err := os.ReadFile(dfPath); err == nil {
			content := strings.ToLower(string(data))
			if strings.Contains(content, "python") || strings.Contains(content, "requirements.txt") || strings.Contains(content, "manage.py") || strings.Contains(content, "pip ") || strings.Contains(content, "uv ") {
				return true
			}
			// If it mentions other stacks but not python, it's probably NOT python
			if strings.Contains(content, "node") || strings.Contains(content, "package.json") || strings.Contains(content, "npm") || strings.Contains(content, "go.mod") || strings.Contains(content, "go build") {
				return false
			}
		}
	}

	if svc.Command != "" {
		cmd := strings.ToLower(svc.Command)
		if strings.Contains(cmd, "python") || strings.Contains(cmd, "manage.py") || strings.Contains(cmd, "pip ") || strings.Contains(cmd, "uv ") {
			return true
		}
		if strings.Contains(cmd, "npm") || strings.Contains(cmd, "node") || strings.Contains(cmd, "go build") || strings.Contains(cmd, "go run") {
			return false
		}
	}

	if svc.Image != "" {
		img := strings.ToLower(svc.Image)
		if strings.Contains(img, "python") {
			return true
		}
		if strings.Contains(img, "node") || strings.Contains(img, "golang") || strings.Contains(img, "postgres") || strings.Contains(img, "redis") {
			return false
		}
	}

	name := strings.ToLower(svc.Name)
	if (strings.Contains(name, "node") || strings.Contains(name, "npm") || strings.Contains(name, "js") || strings.Contains(name, "frontend") || strings.Contains(name, "ui")) && !strings.Contains(name, "python") && !strings.Contains(name, "django") {
		return false
	}
	return strings.Contains(name, "python") || strings.Contains(name, "django") || strings.Contains(name, "api") || strings.Contains(name, "backend")
}
