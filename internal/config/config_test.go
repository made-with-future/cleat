package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	content := `
version: 1
docker: true
python:
  django: true
  django_service: custom-backend
`
	tmpfile, err := os.CreateTemp("", "cleat.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("Expected Version to be 1, got %d", cfg.Version)
	}
	if !cfg.Docker {
		t.Error("Expected Docker to be true")
	}

	// Root level fields should be migrated and cleared
	if cfg.Python != nil {
		t.Error("Expected root Python to be migrated and cleared")
	}

	// Check if it was migrated to Services
	if len(cfg.Services) != 1 {
		t.Fatalf("Expected 1 migrated service, got %d", len(cfg.Services))
	}
	svc := cfg.Services[0]
	if svc.Name != "default" {
		t.Errorf("Expected migrated service name 'default', got '%s'", svc.Name)
	}
	if len(svc.Modules) == 0 || svc.Modules[0].Python == nil || !svc.Modules[0].Python.Django {
		t.Error("Expected Django enabled in migrated module")
	}
	if svc.Modules[0].Python.DjangoService != "custom-backend" {
		t.Errorf("Expected DjangoService to be 'custom-backend', got '%s'", svc.Modules[0].Python.DjangoService)
	}
}

func TestLoadConfigV2(t *testing.T) {
	content := `
version: 2
docker: true
services:
  - name: backend
    location: ./backend
    python:
      django: true
      django_service: web
  - name: frontend
    location: ./frontend
    npm:
      scripts:
        - build
`
	tmpfile, err := os.CreateTemp("", "cleat_v2.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Version != 2 {
		t.Errorf("Expected version 2, got %d", cfg.Version)
	}
	if len(cfg.Services) != 2 {
		t.Fatalf("Expected 2 services, got %d", len(cfg.Services))
	}

	svc1 := cfg.Services[0]
	if svc1.Name != "backend" || svc1.Location != "./backend" {
		t.Errorf("Unexpected svc1: %+v", svc1)
	}
	if len(svc1.Modules) == 0 || svc1.Modules[0].Python == nil || !svc1.Modules[0].Python.Django {
		t.Error("Expected Django enabled for svc1 in modules")
	}

	svc2 := cfg.Services[1]
	if svc2.Name != "frontend" || svc2.Location != "./frontend" {
		t.Errorf("Unexpected svc2: %+v", svc2)
	}
	if len(svc2.Modules) == 0 || svc2.Modules[0].Npm == nil || len(svc2.Modules[0].Npm.Scripts) != 1 {
		t.Error("Expected NPM scripts for svc2 in modules")
	}
}

func TestLoadConfigDefaultVersion(t *testing.T) {
	content := `
docker: true
`
	tmpfile, err := os.CreateTemp("", "cleat.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Version != LatestVersion {
		t.Errorf("Expected default Version to be %d, got %d", LatestVersion, cfg.Version)
	}
}

func TestLoadConfigInvalidVersion(t *testing.T) {
	content := `
version: 99
docker: true
`
	tmpfile, err := os.CreateTemp("", "cleat_invalid_version.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = LoadConfig(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for unrecognized version, got nil")
	}
}

func TestLoadConfigDefaultService(t *testing.T) {
	content := `
python:
  django: true
`
	tmpfile, err := os.CreateTemp("", "cleat.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Services) == 0 || len(cfg.Services[0].Modules) == 0 || cfg.Services[0].Modules[0].Python == nil || cfg.Services[0].Modules[0].Python.DjangoService != "backend" {
		t.Errorf("Expected default DjangoService to be 'backend', got '%v'", cfg.Services[0].Modules[0].Python)
	}
}

func TestLoadConfigAutoDocker(t *testing.T) {
	// Create a dummy docker-compose.yaml in the current directory or a temp one
	// Since LoadConfig takes a path to cleat.yaml, we should probably run it in a temp dir

	tmpDir, err := os.MkdirTemp("", "cleat-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	err = os.WriteFile("docker-compose.yaml", []byte("version: '3'"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile("cleat.yaml", []byte("python:\n  django: true"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig("cleat.yaml")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if !cfg.Docker {
		t.Error("Expected Docker to be auto-detected as true because docker-compose.yaml exists")
	}
}

func TestLoadConfigAutoNpm(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-npm-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.Mkdir("frontend", 0755)
	err = os.WriteFile("frontend/package.json", []byte("{}"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile("cleat.yaml", []byte("python:\n  django: true"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig("cleat.yaml")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	foundNpm := false
	for _, mod := range cfg.Services[0].Modules {
		if mod.Npm != nil && len(mod.Npm.Scripts) == 1 && mod.Npm.Scripts[0] == "build" {
			foundNpm = true
			break
		}
	}
	if !foundNpm {
		t.Errorf("Expected Npm scripts to be ['build'], got %v", cfg.Services[0].Modules)
	}
}
