package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/madewithfuture/cleat/internal/config/schema"
)

type RubyDetector struct{}

func (d *RubyDetector) Detect(baseDir string, cfg *schema.Config) error {
	rootCovered := false
	for _, svc := range cfg.Services {
		if svc.Dir == "." || svc.Dir == "" {
			rootCovered = true
			break
		}
	}

	if !rootCovered {
		if _, err := os.Stat(filepath.Join(baseDir, "Gemfile")); err == nil {
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
		hasGemfile := false
		if _, err := os.Stat(filepath.Join(searchDir, "Gemfile")); err == nil {
			hasGemfile = true
		} else if searchDir == baseDir {
			// If not found in root, check if any service matches a subdirectory containing Gemfile
			for _, s := range svcs {
				if (s.Dir == "." || s.Dir == "") && s.Name != "" {
					if _, err := os.Stat(filepath.Join(baseDir, s.Name, "Gemfile")); err == nil {
						hasGemfile = true
						break
					}
				}
			}
		}

		if hasGemfile {
			var matches []*schema.ServiceConfig
			var others []*schema.ServiceConfig
			for _, s := range svcs {
				explicit := false
				for _, m := range s.Modules {
					if m.Ruby != nil {
						explicit = true
						break
					}
				}
				if explicit {
					continue
				}

				if matchesRuby(s, searchDir) {
					matches = append(matches, s)
				} else {
					others = append(others, s)
				}
			}

			if len(matches) > 0 {
				for _, s := range matches {
					s.Modules = append(s.Modules, schema.ModuleConfig{Ruby: d.detectRubyConfig(s, searchDir)})
				}
			} else if len(others) > 0 {
				for _, s := range others {
					s.Modules = append(s.Modules, schema.ModuleConfig{Ruby: d.detectRubyConfig(s, searchDir)})
				}
			}
		}
	}

	// Apply defaults
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		for j := range svc.Modules {
			mod := &svc.Modules[j]
			if mod.Ruby != nil && mod.Ruby.IsEnabled() {
				if mod.Ruby.RailsService == "" {
					mod.Ruby.RailsService = svc.Name
				}
			}
		}
	}

	return nil
}

func (d *RubyDetector) detectRubyConfig(svc *schema.ServiceConfig, dir string) *schema.RubyConfig {
	isRails := false
	if _, err := os.Stat(filepath.Join(dir, "bin", "rails")); err == nil {
		isRails = true
	} else if _, err := os.Stat(filepath.Join(dir, "config", "application.rb")); err == nil {
		isRails = true
	}

	return &schema.RubyConfig{
		Rails: isRails,
	}
}

func matchesRuby(svc *schema.ServiceConfig, searchDir string) bool {
	if svc.Dockerfile != "" {
		dfPath := filepath.Join(searchDir, svc.Dockerfile)
		if data, err := os.ReadFile(dfPath); err == nil {
			content := strings.ToLower(string(data))
			if strings.Contains(content, "ruby") || strings.Contains(content, "gemfile") || strings.Contains(content, "bundle ") || strings.Contains(content, "rails ") {
				return true
			}
			if strings.Contains(content, "node") || strings.Contains(content, "package.json") || strings.Contains(content, "python") || strings.Contains(content, "go.mod") {
				return false
			}
		}
	}

	if svc.Command != "" {
		cmd := strings.ToLower(svc.Command)
		if strings.Contains(cmd, "ruby") || strings.Contains(cmd, "bundle ") || strings.Contains(cmd, "rails ") || strings.Contains(cmd, "rake ") {
			return true
		}
		if strings.Contains(cmd, "npm") || strings.Contains(cmd, "node") || strings.Contains(cmd, "python") || strings.Contains(cmd, "go build") {
			return false
		}
	}

	if svc.Image != "" {
		img := strings.ToLower(svc.Image)
		if strings.Contains(img, "ruby") {
			return true
		}
		if strings.Contains(img, "node") || strings.Contains(img, "golang") || strings.Contains(img, "python") {
			return false
		}
	}

	name := strings.ToLower(svc.Name)
	if (strings.Contains(name, "node") || strings.Contains(name, "npm") || strings.Contains(name, "python") || strings.Contains(name, "frontend")) && !strings.Contains(name, "ruby") && !strings.Contains(name, "rails") {
		return false
	}
	return strings.Contains(name, "ruby") || strings.Contains(name, "rails") || strings.Contains(name, "api") || strings.Contains(name, "backend")
}
