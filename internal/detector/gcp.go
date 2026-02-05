package detector

import (
	"os"
	"path/filepath"

	"github.com/madewithfuture/cleat/internal/config/schema"
)

type GcpDetector struct{}

func (d *GcpDetector) Detect(baseDir string, cfg *schema.Config) error {
	if cfg.GoogleCloudPlatform == nil {
		return nil
	}

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
						cfg.Services = append(cfg.Services, schema.ServiceConfig{
							Name:    entry.Name(),
							Dir:     entry.Name(),
							AppYaml: appYamlPath,
						})
					}
				}
			}
		}
	}

	return nil
}