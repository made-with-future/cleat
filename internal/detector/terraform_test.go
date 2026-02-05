package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
)

func TestTerraformDetector(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	iacDir := filepath.Join(tmpDir, ".iac")
	os.Mkdir(iacDir, 0755)
	devDir := filepath.Join(iacDir, "dev")
	os.Mkdir(devDir, 0755)
	os.WriteFile(filepath.Join(devDir, "main.tf"), []byte(""), 0644)

	cfg := &config.Config{}
	d := &TerraformDetector{}
	err = d.Detect(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if cfg.Terraform == nil || !cfg.Terraform.UseFolders {
		t.Error("expected Terraform with UseFolders=true")
	}

	if len(cfg.Terraform.Envs) != 1 || cfg.Terraform.Envs[0] != "dev" {
		t.Errorf("expected Terraform env 'dev', got %v", cfg.Terraform.Envs)
	}
}
