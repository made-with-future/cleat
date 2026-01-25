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

	packageJsonContent := `{
		"scripts": {
			"build": "vite build",
			"test": "vitest"
		}
	}`
	err = os.WriteFile("package.json", []byte(packageJsonContent), 0644)
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
		if mod.Npm != nil {
			foundNpm = true
			if len(mod.Npm.Scripts) != 2 {
				t.Errorf("Expected 2 Npm scripts, got %d: %v", len(mod.Npm.Scripts), mod.Npm.Scripts)
			}
			scriptsMap := make(map[string]bool)
			for _, s := range mod.Npm.Scripts {
				scriptsMap[s] = true
			}
			if !scriptsMap["build"] || !scriptsMap["test"] {
				t.Errorf("Expected scripts 'build' and 'test', got %v", mod.Npm.Scripts)
			}
			break
		}
	}
	if !foundNpm {
		t.Error("Expected Npm module to be auto-detected")
	}
}

func TestLoadConfigAutoNpm_NoFrontend(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-npm-no-frontend-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.Mkdir("frontend", 0755)
	err = os.WriteFile("frontend/package.json", []byte(`{"scripts":{"build":"echo"}}`), 0644)
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

	for _, mod := range cfg.Services[0].Modules {
		if mod.Npm != nil {
			t.Error("Expected Npm module NOT to be auto-detected in frontend/ directory")
		}
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

	// Create some files to trigger auto-detection
	dockerComposeContent := `
version: '3'
services:
  backend:
    build: .
`
	err = os.WriteFile("docker-compose.yml", []byte(dockerComposeContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadDefaultConfig()
	if err != nil {
		t.Fatalf("Expected no error when cleat.yaml is not found, got %v", err)
	}
	if cfg == nil {
		t.Fatal("Expected non-nil config when cleat.yaml is not found")
	}

	if cfg.Version != LatestVersion {
		t.Errorf("Expected version %d, got %d", LatestVersion, cfg.Version)
	}

	if !cfg.Docker {
		t.Error("Expected Docker to be true via auto-detection")
	}

	foundBackend := false
	for _, svc := range cfg.Services {
		if svc.Name == "backend" {
			foundBackend = true
			break
		}
	}
	if !foundBackend {
		t.Error("Expected backend service to be auto-detected")
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

	// 3. Setup cleat.yaml that should override Python but allow NPM auto-detect
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
			break
		}
	}

	if backendSvc == nil {
		t.Fatal("Expected to find backend service")
	}

	// Backend should have docker=false from cleat.yaml override
	if backendSvc.IsDocker() {
		t.Error("Expected backend to have docker=false from cleat.yaml override")
	}

	// Check modules
	var pythonMod *PythonConfig
	var npmMod *NpmConfig
	for _, mod := range backendSvc.Modules {
		if mod.Python != nil {
			pythonMod = mod.Python
		}
		if mod.Npm != nil {
			npmMod = mod.Npm
		}
	}

	// Python should come from explicit config
	if pythonMod == nil {
		t.Fatal("Expected Python module from explicit config")
	}
	if !pythonMod.Django {
		t.Error("Expected Python module to have django=true from explicit config")
	}

	// NPM should be auto-detected since it wasn't explicitly configured
	// (This is the NEW behavior - independent auto-detection per module type)
	if npmMod == nil {
		t.Error("Expected NPM module to be auto-detected since it wasn't explicitly configured")
	}
}

func TestLoadConfig_ServicePrecedence_DisableNpm(t *testing.T) {
	// This test verifies that you CAN disable NPM if you don't want it auto-detected
	tmpDir, err := os.MkdirTemp("", "cleat-test-precedence-disable-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Setup files for auto-detection
	os.WriteFile("manage.py", []byte(""), 0644)
	os.WriteFile("package.json", []byte("{}"), 0644)

	// cleat.yaml that explicitly disables NPM
	cleatYamlContent := `
services:
  - name: backend
    modules:
      - python:
          django: true
      - npm:
          enabled: false
`
	os.WriteFile("cleat.yaml", []byte(cleatYamlContent), 0644)

	cfg, err := LoadConfig("cleat.yaml")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	svc := cfg.Services[0]

	var pythonMod *PythonConfig
	var npmMod *NpmConfig
	for _, mod := range svc.Modules {
		if mod.Python != nil {
			pythonMod = mod.Python
		}
		if mod.Npm != nil {
			npmMod = mod.Npm
		}
	}

	if pythonMod == nil || !pythonMod.IsEnabled() {
		t.Error("Expected Python module to be enabled")
	}

	if npmMod == nil {
		t.Fatal("Expected NPM module config to exist")
	}
	if npmMod.IsEnabled() {
		t.Error("Expected NPM module to be disabled via explicit enabled: false")
	}
}

func TestPythonConfig_IsEnabled(t *testing.T) {
	t.Run("nil config returns false", func(t *testing.T) {
		var p *PythonConfig
		if p.IsEnabled() {
			t.Error("Expected nil PythonConfig to return false")
		}
	})

	t.Run("nil Enabled field defaults to true", func(t *testing.T) {
		p := &PythonConfig{Django: true}
		if !p.IsEnabled() {
			t.Error("Expected PythonConfig with nil Enabled to return true")
		}
	})

	t.Run("explicit true returns true", func(t *testing.T) {
		enabled := true
		p := &PythonConfig{Enabled: &enabled}
		if !p.IsEnabled() {
			t.Error("Expected PythonConfig with Enabled=true to return true")
		}
	})

	t.Run("explicit false returns false", func(t *testing.T) {
		enabled := false
		p := &PythonConfig{Enabled: &enabled}
		if p.IsEnabled() {
			t.Error("Expected PythonConfig with Enabled=false to return false")
		}
	})
}

func TestNpmConfig_IsEnabled(t *testing.T) {
	t.Run("nil config returns false", func(t *testing.T) {
		var n *NpmConfig
		if n.IsEnabled() {
			t.Error("Expected nil NpmConfig to return false")
		}
	})

	t.Run("nil Enabled field defaults to true", func(t *testing.T) {
		n := &NpmConfig{Scripts: []string{"build"}}
		if !n.IsEnabled() {
			t.Error("Expected NpmConfig with nil Enabled to return true")
		}
	})

	t.Run("explicit true returns true", func(t *testing.T) {
		enabled := true
		n := &NpmConfig{Enabled: &enabled}
		if !n.IsEnabled() {
			t.Error("Expected NpmConfig with Enabled=true to return true")
		}
	})

	t.Run("explicit false returns false", func(t *testing.T) {
		enabled := false
		n := &NpmConfig{Enabled: &enabled}
		if n.IsEnabled() {
			t.Error("Expected NpmConfig with Enabled=false to return false")
		}
	})
}

func TestLoadConfig_ModuleAutoDetectWithExplicitOverride(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-module-override-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Setup files that would trigger auto-detection for both Python and NPM
	os.WriteFile("manage.py", []byte(""), 0644)
	os.WriteFile("package.json", []byte("{}"), 0644)

	t.Run("explicit Python overrides auto-detect, NPM still auto-detected", func(t *testing.T) {
		cleatYaml := `
services:
  - name: backend
    modules:
      - python:
          django: true
          package_manager: poetry
`
		os.WriteFile("cleat.yaml", []byte(cleatYaml), 0644)

		cfg, err := LoadConfig("cleat.yaml")
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		svc := cfg.Services[0]

		// Should have both Python and NPM modules
		var pythonMod *PythonConfig
		var npmMod *NpmConfig
		for _, mod := range svc.Modules {
			if mod.Python != nil {
				pythonMod = mod.Python
			}
			if mod.Npm != nil {
				npmMod = mod.Npm
			}
		}

		if pythonMod == nil {
			t.Fatal("Expected Python module to exist")
		}
		if pythonMod.PackageManager != "poetry" {
			t.Errorf("Expected package_manager 'poetry' from explicit config, got '%s'", pythonMod.PackageManager)
		}

		if npmMod == nil {
			t.Error("Expected NPM module to be auto-detected since it wasn't explicitly configured")
		}
	})

	t.Run("explicit NPM overrides auto-detect, Python still auto-detected", func(t *testing.T) {
		cleatYaml := `
services:
  - name: backend
    modules:
      - npm:
          scripts: [lint, test]
`
		os.WriteFile("cleat.yaml", []byte(cleatYaml), 0644)

		cfg, err := LoadConfig("cleat.yaml")
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		svc := cfg.Services[0]

		var pythonMod *PythonConfig
		var npmMod *NpmConfig
		for _, mod := range svc.Modules {
			if mod.Python != nil {
				pythonMod = mod.Python
			}
			if mod.Npm != nil {
				npmMod = mod.Npm
			}
		}

		if pythonMod == nil {
			t.Error("Expected Python module to be auto-detected since it wasn't explicitly configured")
		}

		if npmMod == nil {
			t.Fatal("Expected NPM module to exist")
		}
		if len(npmMod.Scripts) != 2 || npmMod.Scripts[0] != "lint" || npmMod.Scripts[1] != "test" {
			t.Errorf("Expected scripts [lint, test] from explicit config, got %v", npmMod.Scripts)
		}
	})
}

func TestLoadConfig_DisableModule(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-disable-module-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Setup files that would trigger auto-detection
	os.WriteFile("manage.py", []byte(""), 0644)
	os.WriteFile("package.json", []byte("{}"), 0644)

	t.Run("disable Python auto-detection", func(t *testing.T) {
		cleatYaml := `
services:
  - name: backend
    modules:
      - python:
          enabled: false
`
		os.WriteFile("cleat.yaml", []byte(cleatYaml), 0644)

		cfg, err := LoadConfig("cleat.yaml")
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		svc := cfg.Services[0]

		var pythonMod *PythonConfig
		var npmMod *NpmConfig
		for _, mod := range svc.Modules {
			if mod.Python != nil {
				pythonMod = mod.Python
			}
			if mod.Npm != nil {
				npmMod = mod.Npm
			}
		}

		// Python should exist but be disabled
		if pythonMod == nil {
			t.Fatal("Expected Python module config to exist")
		}
		if pythonMod.IsEnabled() {
			t.Error("Expected Python module to be disabled")
		}

		// NPM should still be auto-detected
		if npmMod == nil {
			t.Error("Expected NPM module to be auto-detected")
		}
	})

	t.Run("disable NPM auto-detection", func(t *testing.T) {
		cleatYaml := `
services:
  - name: backend
    modules:
      - npm:
          enabled: false
`
		os.WriteFile("cleat.yaml", []byte(cleatYaml), 0644)

		cfg, err := LoadConfig("cleat.yaml")
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		svc := cfg.Services[0]

		var pythonMod *PythonConfig
		var npmMod *NpmConfig
		for _, mod := range svc.Modules {
			if mod.Python != nil {
				pythonMod = mod.Python
			}
			if mod.Npm != nil {
				npmMod = mod.Npm
			}
		}

		// Python should be auto-detected
		if pythonMod == nil {
			t.Error("Expected Python module to be auto-detected")
		}

		// NPM should exist but be disabled
		if npmMod == nil {
			t.Fatal("Expected NPM module config to exist")
		}
		if npmMod.IsEnabled() {
			t.Error("Expected NPM module to be disabled")
		}
	})

	t.Run("disable both modules", func(t *testing.T) {
		cleatYaml := `
services:
  - name: backend
    modules:
      - python:
          enabled: false
      - npm:
          enabled: false
`
		os.WriteFile("cleat.yaml", []byte(cleatYaml), 0644)

		cfg, err := LoadConfig("cleat.yaml")
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		svc := cfg.Services[0]

		for _, mod := range svc.Modules {
			if mod.Python != nil && mod.Python.IsEnabled() {
				t.Error("Expected Python module to be disabled")
			}
			if mod.Npm != nil && mod.Npm.IsEnabled() {
				t.Error("Expected NPM module to be disabled")
			}
		}
	})
}

func TestLoadConfig_DisabledModuleSkipsDefaults(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-disabled-defaults-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile("package.json", []byte("{}"), 0644)

	cleatYaml := `
services:
  - name: backend
    modules:
      - python:
          enabled: false
          django: true
`
	os.WriteFile("cleat.yaml", []byte(cleatYaml), 0644)

	cfg, err := LoadConfig("cleat.yaml")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	svc := cfg.Services[0]

	for _, mod := range svc.Modules {
		if mod.Python != nil {
			// Disabled modules should NOT have defaults applied
			if mod.Python.PackageManager != "" {
				t.Errorf("Expected disabled Python module to NOT have package_manager default applied, got '%s'", mod.Python.PackageManager)
			}
			if mod.Python.DjangoService != "" {
				t.Errorf("Expected disabled Python module to NOT have django_service default applied, got '%s'", mod.Python.DjangoService)
			}
		}
	}
}

func TestLoadDefaultConfig_UpwardsSearch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-upwards-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	projectRoot := tmpDir
	subDir := filepath.Join(projectRoot, "a", "b", "c")
	os.MkdirAll(subDir, 0755)

	cleatYaml := "version: 1\ndocker: false"
	os.WriteFile(filepath.Join(projectRoot, "cleat.yaml"), []byte(cleatYaml), 0644)

	oldWd, _ := os.Getwd()
	os.Chdir(subDir)
	defer os.Chdir(oldWd)

	cfg, err := LoadDefaultConfig()
	if err != nil {
		t.Fatalf("LoadDefaultConfig failed: %v", err)
	}

	if cfg.Docker != false {
		t.Error("Expected Docker to be false (from upwards found cleat.yaml)")
	}

	absRoot, _ := filepath.Abs(projectRoot)
	expectedSourcePath := filepath.Join(absRoot, "cleat.yaml")
	if cfg.SourcePath != expectedSourcePath {
		t.Errorf("Expected SourcePath '%s', got '%s'", expectedSourcePath, cfg.SourcePath)
	}
}
