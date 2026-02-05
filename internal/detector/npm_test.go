package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
)

func TestNpmDetector(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	packageJson := `{"scripts": {"start": "node index.js", "test": "jest"}}`
	err = os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJson), 0644)
	if err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	cfg := &config.Config{
		Services: []config.ServiceConfig{
			{Name: "frontend"},
		},
	}
	d := &NpmDetector{}
	err = d.Detect(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(cfg.Services[0].Modules) != 1 || cfg.Services[0].Modules[0].Npm == nil {
		t.Errorf("expected Npm module to be detected, got: %+v", cfg.Services[0].Modules)
	}

	scripts := cfg.Services[0].Modules[0].Npm.Scripts
	if len(scripts) != 2 || scripts[0] != "start" || scripts[1] != "test" {
		t.Errorf("expected scripts [start, test], got %v", scripts)
	}
}
