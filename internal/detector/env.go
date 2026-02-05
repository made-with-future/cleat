package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
)

func init() {
	Register(&EnvDetector{})
}

type EnvDetector struct{}

func (d *EnvDetector) Detect(baseDir string, cfg *config.Config) error {
	if cfg.Envs != nil {
		return nil
	}

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
	return nil
}
