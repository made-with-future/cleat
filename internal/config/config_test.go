package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	content := `
version: 1
docker: true
services:
  - name: backend
    dir: ./backend
    modules:
      - python:
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

	// Check Services
	if len(cfg.Services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(cfg.Services))
	}
	svc := cfg.Services[0]
	if svc.Name != "backend" {
		t.Errorf("Expected service name 'backend', got '%s'", svc.Name)
	}
	if len(svc.Modules) == 0 || svc.Modules[0].Python == nil || !svc.Modules[0].Python.Django {
		t.Error("Expected Django enabled in module")
	}
	if svc.Modules[0].Python.DjangoService != "custom-backend" {
		t.Errorf("Expected DjangoService to be 'custom-backend', got '%s'", svc.Modules[0].Python.DjangoService)
	}
}

func TestLoadConfigMultiService(t *testing.T) {
	content := `
version: 1
docker: true
services:
  - name: backend
    dir: ./backend
    modules:
      - python:
          django: true
          django_service: web
  - name: frontend
    dir: ./frontend
    modules:
      - npm:
          scripts:
            - build
`
	tmpfile, err := os.CreateTemp("", "cleat_multi.yaml")
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

	if cfg.Version != 1 {
		t.Errorf("Expected version 1, got %d", cfg.Version)
	}
	if len(cfg.Services) != 2 {
		t.Fatalf("Expected 2 services, got %d", len(cfg.Services))
	}

	svc1 := cfg.Services[0]
	if svc1.Name != "backend" || svc1.Dir != "./backend" {
		t.Errorf("Unexpected svc1: %+v", svc1)
	}
	if len(svc1.Modules) == 0 || svc1.Modules[0].Python == nil || !svc1.Modules[0].Python.Django {
		t.Error("Expected Django enabled for svc1 in modules")
	}

	svc2 := cfg.Services[1]
	if svc2.Name != "frontend" || svc2.Dir != "./frontend" {
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

func TestLoadConfigEnvs(t *testing.T) {
	t.Run("Valid envs", func(t *testing.T) {
		content := `
version: 1
envs:
  - production
  - staging
`
		tmpfile, err := os.CreateTemp("", "cleat_envs.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())
		os.WriteFile(tmpfile.Name(), []byte(content), 0644)

		cfg, err := LoadConfig(tmpfile.Name())
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if len(cfg.Envs) != 2 {
			t.Errorf("Expected 2 envs, got %d", len(cfg.Envs))
		}
		if cfg.Envs[0] != "production" || cfg.Envs[1] != "staging" {
			t.Errorf("Unexpected envs: %v", cfg.Envs)
		}
	})

	t.Run("Invalid empty envs", func(t *testing.T) {
		content := `
version: 1
envs: []
`
		tmpfile, err := os.CreateTemp("", "cleat_envs_empty.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())
		os.WriteFile(tmpfile.Name(), []byte(content), 0644)

		_, err = LoadConfig(tmpfile.Name())
		if err == nil {
			t.Error("Expected error for empty envs, got nil")
		} else if err.Error() != "envs must have at least one item if provided" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("Omitted envs is valid", func(t *testing.T) {
		content := `
version: 1
`
		tmpfile, err := os.CreateTemp("", "cleat_envs_omitted.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())
		os.WriteFile(tmpfile.Name(), []byte(content), 0644)

		cfg, err := LoadConfig(tmpfile.Name())
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if cfg.Envs != nil {
			t.Errorf("Expected Envs to be nil when omitted, got %v", cfg.Envs)
		}
	})

	t.Run("Auto-detect envs", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "cleat_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		envsDir := filepath.Join(tempDir, ".envs")
		if err := os.Mkdir(envsDir, 0755); err != nil {
			t.Fatal(err)
		}

		os.WriteFile(filepath.Join(envsDir, "production.env"), []byte(""), 0644)
		os.WriteFile(filepath.Join(envsDir, "staging.env"), []byte(""), 0644)
		os.WriteFile(filepath.Join(envsDir, "README.md"), []byte(""), 0644) // Should be ignored

		configPath := filepath.Join(tempDir, "cleat.yaml")
		os.WriteFile(configPath, []byte("version: 1"), 0644)

		cfg, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if len(cfg.Envs) != 2 {
			t.Errorf("Expected 2 auto-detected envs, got %d: %v", len(cfg.Envs), cfg.Envs)
		}

		// Order should be alphabetical due to os.ReadDir
		if cfg.Envs[0] != "production" || cfg.Envs[1] != "staging" {
			t.Errorf("Unexpected auto-detected envs: %v", cfg.Envs)
		}
	})
}

func TestLoadConfigPackageManager(t *testing.T) {
	t.Run("Default to uv", func(t *testing.T) {
		content := `
services:
  - name: backend
    modules:
      - python:
          django: true
`
		tmpfile, err := os.CreateTemp("", "cleat_pm_default.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())
		os.WriteFile(tmpfile.Name(), []byte(content), 0644)

		cfg, err := LoadConfig(tmpfile.Name())
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		pm := cfg.Services[0].Modules[0].Python.PackageManager
		if pm != "uv" {
			t.Errorf("Expected default package manager 'uv', got '%s'", pm)
		}
	})

	t.Run("Explicit pip", func(t *testing.T) {
		content := `
services:
  - name: backend
    modules:
      - python:
          django: true
          package_manager: pip
`
		tmpfile, err := os.CreateTemp("", "cleat_pm_pip.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())
		os.WriteFile(tmpfile.Name(), []byte(content), 0644)

		cfg, err := LoadConfig(tmpfile.Name())
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		pm := cfg.Services[0].Modules[0].Python.PackageManager
		if pm != "pip" {
			t.Errorf("Expected package manager 'pip', got '%s'", pm)
		}
	})
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
services:
  - name: default
    modules:
      - python:
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

	err = os.WriteFile("cleat.yaml", []byte("services:\n  - name: default"), 0644)
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

	err = os.WriteFile("cleat.yaml", []byte("services:\n  - name: default"), 0644)
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

func TestLoadConfigGCP(t *testing.T) {
	content := `
version: 1
google_cloud_platform:
  project_name: test-project
`
	tmpfile, err := os.CreateTemp("", "cleat_gcp.yaml")
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

	if cfg.GoogleCloudPlatform == nil {
		t.Fatal("Expected GoogleCloudPlatform to be not nil")
	}
	if cfg.GoogleCloudPlatform.ProjectName != "test-project" {
		t.Errorf("Expected project_name to be 'test-project', got '%s'", cfg.GoogleCloudPlatform.ProjectName)
	}
}

func TestLoadDefaultConfig_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-no-config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	cfg, err := LoadDefaultConfig()
	if cfg != nil {
		t.Error("Expected nil config when cleat.yaml is not found")
	}
	if err == nil {
		t.Fatal("Expected error when cleat.yaml is not found")
	}

	expectedErr := "no cleat.yaml found in current directory"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestLoadConfig_DockerComposeServices(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-docker-compose-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	dockerComposeContent := `
version: '3'
services:
  backend:
    build: .
  worker:
    image: redis
`
	err = os.WriteFile("docker-compose.yml", []byte(dockerComposeContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cleatYamlContent := `
version: 1
services:
  - name: frontend
    dir: frontend
`
	err = os.WriteFile("cleat.yaml", []byte(cleatYamlContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile("manage.py", []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig("cleat.yaml")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if !cfg.Docker {
		t.Error("Expected Docker to be true")
	}

	expectedServices := map[string]bool{
		"frontend": true,
		"backend":  true,
		"worker":   true,
	}

	if len(cfg.Services) != 3 {
		t.Errorf("Expected 3 services, got %d", len(cfg.Services))
	}

	for _, svc := range cfg.Services {
		if !expectedServices[svc.Name] {
			t.Errorf("Unexpected service: %s", svc.Name)
		}
		if svc.Name == "backend" {
			foundPython := false
			for _, mod := range svc.Modules {
				if mod.Python != nil && mod.Python.Django {
					foundPython = true
					break
				}
			}
			if !foundPython {
				t.Errorf("Expected service %s to have Python module (Django)", svc.Name)
			}
		}
	}
}

func TestLoadConfig_MultiServiceDockerCompose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-multi-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	dockerComposeContent := `
version: '3'
services:
  backend:
    build: ./backend
  backend-wo-db:
    build: ./backend
  frontend:
    build: ./frontend
`
	err = os.WriteFile("docker-compose.yml", []byte(dockerComposeContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	os.Mkdir("backend", 0755)
	os.WriteFile("backend/manage.py", []byte(""), 0644)
	os.Mkdir("frontend", 0755)
	os.WriteFile("frontend/package.json", []byte("{}"), 0644)

	cleatYamlContent := "version: 1\n"
	err = os.WriteFile("cleat.yaml", []byte(cleatYamlContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig("cleat.yaml")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Services) != 3 {
		t.Fatalf("Expected 3 services, got %d", len(cfg.Services))
	}

	for _, svc := range cfg.Services {
		if svc.Name == "backend" || svc.Name == "backend-wo-db" {
			found := false
			for _, mod := range svc.Modules {
				if mod.Python != nil && mod.Python.Django {
					found = true
					if mod.Python.DjangoService != svc.Name {
						t.Errorf("Expected DjangoService for %s to be %s, got %s", svc.Name, svc.Name, mod.Python.DjangoService)
					}
				}
			}
			if !found {
				t.Errorf("Expected service %s to have Django module", svc.Name)
			}
		}
	}
}

func TestLoadConfig_ServicePrecedence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-precedence-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// 1. Setup docker-compose.yml
	dockerComposeContent := `
services:
  backend:
    build: .
  worker:
    build: ./worker
`
	os.WriteFile("docker-compose.yml", []byte(dockerComposeContent), 0644)

	// 2. Setup files for auto-detection
	os.WriteFile("manage.py", []byte(""), 0644)
	os.WriteFile("package.json", []byte("{}"), 0644)

	// 3. Setup cleat.yaml that should override
	cleatYamlContent := `
services:
  - name: backend
    docker: false
    modules:
      - python:
          django: true
`
	os.WriteFile("cleat.yaml", []byte(cleatYamlContent), 0644)

	cfg, err := LoadConfig("cleat.yaml")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	var backendSvc *ServiceConfig
	for i := range cfg.Services {
		if cfg.Services[i].Name == "backend" {
			backendSvc = &cfg.Services[i]
		}
	}

	if backendSvc == nil {
		t.Fatal("backend service not found")
	}

	// TEST 1: docker flag should be false because it was explicitly set in cleat.yaml
	if backendSvc.IsDocker() {
		t.Errorf("Expected backend.docker to be false (from cleat.yaml), but it was overridden to true by docker-compose.yml")
	}

	// TEST 2: Modules should ONLY contain Python, NOT NPM, because modules were explicitly defined in cleat.yaml
	foundNpm := false
	for _, mod := range backendSvc.Modules {
		if mod.Npm != nil {
			foundNpm = true
		}
	}
	if foundNpm {
		t.Errorf("Expected backend NOT to have NPM module because modules were explicitly defined in cleat.yaml")
	}
}
