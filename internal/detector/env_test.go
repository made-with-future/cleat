package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
)

func TestEnvDetector(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	envsDir := filepath.Join(tmpDir, ".envs")
	os.Mkdir(envsDir, 0755)
	os.WriteFile(filepath.Join(envsDir, "dev.env"), []byte(""), 0644)
	os.WriteFile(filepath.Join(envsDir, "prod.env"), []byte(""), 0644)

	cfg := &config.Config{}
	d := &EnvDetector{}
	err = d.Detect(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(cfg.Envs) != 2 {
		t.Errorf("expected 2 envs, got %d", len(cfg.Envs))
	}
}
