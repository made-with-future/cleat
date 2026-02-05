package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
)

func TestGcpDetector(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "app.yaml"), []byte(""), 0644)

	cfg := &config.Config{
		GoogleCloudPlatform: &config.GCPConfig{ProjectName: "test"},
	}
	d := &GcpDetector{}
	err = d.Detect(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if cfg.AppYaml != "app.yaml" {
		t.Errorf("expected AppYaml to be 'app.yaml', got %s", cfg.AppYaml)
	}
}
