package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madewithfuture/cleat/internal/config/schema"
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

	cfg := &schema.Config{}
	d := &EnvDetector{}
	err = d.Detect(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(cfg.Envs) != 1 {
		t.Errorf("expected 1 env, got %d", len(cfg.Envs))
	}
}