package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madewithfuture/cleat/internal/config/schema"
)

func TestDockerDetector(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dcContent := `
services:
  web:
    build: .
  db:
    image: postgres
`
	err = os.WriteFile(filepath.Join(tmpDir, "docker-compose.yaml"), []byte(dcContent), 0644)
	if err != nil {
		t.Fatalf("failed to write docker-compose: %v", err)
	}

	cfg := &schema.Config{}
	d := &DockerDetector{}
	err = d.Detect(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if !cfg.Docker {
		t.Error("expected cfg.Docker to be true")
	}

	if len(cfg.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(cfg.Services))
	}
}