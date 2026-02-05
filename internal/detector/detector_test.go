package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madewithfuture/cleat/internal/config/schema"
)

func TestDetectAll(t *testing.T) {
	cfg := &schema.Config{}
	err := DetectAll(".", cfg)
	if err != nil {
		t.Fatalf("DetectAll failed: %v", err)
	}
}

func TestDockerDetector(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-docker-detect-*")
	defer os.RemoveAll(tmpDir)

	dockerCompose := `
version: '3'
services:
  web:
    build: .
  db:
    image: postgres
`
	os.WriteFile(filepath.Join(tmpDir, "docker-compose.yaml"), []byte(dockerCompose), 0644)

	d := &DockerDetector{}
	cfg := &schema.Config{}
	err := d.Detect(tmpDir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.Docker {
		t.Error("expected Docker to be true")
	}
	if len(cfg.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(cfg.Services))
	}
}

func TestDjangoDetector(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-django-detect-*")
	defer os.RemoveAll(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "manage.py"), []byte(""), 0644)

	d := &DjangoDetector{}
	cfg := &schema.Config{
		Services: []schema.ServiceConfig{
			{Name: "backend", Dir: "."},
		},
	}
	err := d.Detect(tmpDir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	foundDjango := false
	for _, mod := range cfg.Services[0].Modules {
		if mod.Python != nil && mod.Python.Django {
			foundDjango = true
			break
		}
	}
	if !foundDjango {
		t.Error("expected Django module to be detected")
	}
}

func TestNpmDetector(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-npm-detect-*")
	defer os.RemoveAll(tmpDir)

	packageJson := `{"scripts": {"build": "vite build"}}`
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJson), 0644)

	d := &NpmDetector{}
	cfg := &schema.Config{
		Services: []schema.ServiceConfig{
			{Name: "frontend", Dir: "."},
		},
	}
	err := d.Detect(tmpDir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	foundNpm := false
	for _, mod := range cfg.Services[0].Modules {
		if mod.Npm != nil {
			foundNpm = true
			if len(mod.Npm.Scripts) != 1 || mod.Npm.Scripts[0] != "build" {
				t.Errorf("expected scripts [build], got %v", mod.Npm.Scripts)
			}
			break
		}
	}
	if !foundNpm {
		t.Error("expected NPM module to be detected")
	}
}

func TestTerraformDetector(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-tf-detect-*")
	defer os.RemoveAll(tmpDir)

	iacDir := filepath.Join(tmpDir, ".iac")
	os.Mkdir(iacDir, 0755)
	
	prodDir := filepath.Join(iacDir, "production")
	os.Mkdir(prodDir, 0755)
	os.WriteFile(filepath.Join(prodDir, "main.tf"), []byte(""), 0644)

	d := &TerraformDetector{}
	cfg := &schema.Config{}
	err := d.Detect(tmpDir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Terraform == nil {
		t.Fatal("expected Terraform config to be created")
	}
	if !cfg.Terraform.UseFolders {
		t.Error("expected UseFolders to be true")
	}
	if len(cfg.Terraform.Envs) != 1 || cfg.Terraform.Envs[0] != "production" {
		t.Errorf("expected production env, got %v", cfg.Terraform.Envs)
	}
}
