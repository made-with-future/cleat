package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	content := `
docker: true
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

	if !cfg.Docker {
		t.Error("Expected Docker to be true")
	}
	if !cfg.Django {
		t.Error("Expected Django to be true")
	}
	if cfg.DjangoService != "custom-backend" {
		t.Errorf("Expected DjangoService to be 'custom-backend', got '%s'", cfg.DjangoService)
	}
}

func TestLoadConfigDefaultService(t *testing.T) {
	content := `
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

	if cfg.DjangoService != "backend" {
		t.Errorf("Expected default DjangoService to be 'backend', got '%s'", cfg.DjangoService)
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

	err = os.WriteFile("cleat.yaml", []byte("django: true"), 0644)
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

	err = os.WriteFile("cleat.yaml", []byte("django: true"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig("cleat.yaml")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Npm.Scripts) != 1 || cfg.Npm.Scripts[0] != "build" {
		t.Errorf("Expected Npm scripts to be ['build'], got %v", cfg.Npm.Scripts)
	}
}
