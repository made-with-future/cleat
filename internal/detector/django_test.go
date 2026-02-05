package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madewithfuture/cleat/internal/config/schema"
)

func TestDjangoDetector(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = os.WriteFile(filepath.Join(tmpDir, "manage.py"), []byte(""), 0644)
	if err != nil {
		t.Fatalf("failed to write manage.py: %v", err)
	}

	cfg := &schema.Config{
		Services: []schema.ServiceConfig{
			{Name: "api"},
		},
	}
	d := &DjangoDetector{}
	err = d.Detect(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(cfg.Services[0].Modules) != 1 || cfg.Services[0].Modules[0].Python == nil {
		t.Errorf("expected Python module to be detected, got: %+v", cfg.Services[0].Modules)
	}

	if cfg.Services[0].Modules[0].Python.DjangoService != "api" {
		t.Errorf("expected DjangoService to be 'api', got %s", cfg.Services[0].Modules[0].Python.DjangoService)
	}
}