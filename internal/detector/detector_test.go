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

func TestNpmDetector_InvalidJson(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-npm-invalid-*")
	defer os.RemoveAll(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{invalid json"), 0644)

	d := &NpmDetector{}
	cfg := &schema.Config{
		Services: []schema.ServiceConfig{
			{Name: "frontend", Dir: "."},
		},
	}
	err := d.Detect(tmpDir, cfg)
	if err == nil {
		t.Error("expected error for invalid package.json")
	}
}

func TestDockerDetector_InvalidDir(t *testing.T) {
	d := &DockerDetector{}
	cfg := &schema.Config{}
	err := d.Detect("/non/existent/path", cfg)
	if err != nil {
		t.Fatalf("Detect should not return error for non-existent path, got %v", err)
	}
}

func TestDockerDetector_YmlFallback(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-docker-yml-*")
	defer os.RemoveAll(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte("services: {web: {build: .}}"), 0644)

	d := &DockerDetector{}
	cfg := &schema.Config{}
	d.Detect(tmpDir, cfg)
	if !cfg.Docker {
		t.Error("expected Docker to be true for .yml file")
	}
}

func TestDockerDetector_BuildFormats(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-docker-formats-*")
	defer os.RemoveAll(tmpDir)

	dc := `
services:
  s1:
    build: 
      context: ./s1
  s2:
    build: ./s2
  s3:
    image: redis
`
	os.WriteFile(filepath.Join(tmpDir, "docker-compose.yaml"), []byte(dc), 0644)

	d := &DockerDetector{}
	cfg := &schema.Config{}
	d.Detect(tmpDir, cfg)

	expected := map[string]string{
		"s1": "./s1",
		"s2": "./s2",
		"s3": "",
	}

	for _, s := range cfg.Services {
		if dir, ok := expected[s.Name]; ok {
			if s.Dir != dir {
				t.Errorf("expected %s dir %q, got %q", s.Name, dir, s.Dir)
			}
		} else {
			t.Errorf("unexpected service %s", s.Name)
		}
	}
}

func TestDockerDetector_Malformed(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-docker-malformed-*")
	defer os.RemoveAll(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "docker-compose.yaml"), []byte("invalid: yaml: ["), 0644)

	d := &DockerDetector{}
	cfg := &schema.Config{}
	err := d.Detect(tmpDir, cfg)
	if err == nil {
		t.Error("expected error for malformed docker-compose.yaml")
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

func TestTerraformDetector_Recursive(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-tf-recursive-*")
	defer os.RemoveAll(tmpDir)

	// Case 1: .iac in a subfolder
	projectDir := filepath.Join(tmpDir, "my-project")
	iacDir := filepath.Join(projectDir, ".iac")
	os.MkdirAll(filepath.Join(iacDir, "prod"), 0755)
	os.WriteFile(filepath.Join(iacDir, "prod", "main.tf"), []byte(""), 0644)

	d := &TerraformDetector{}
	cfg := &schema.Config{}
	err := d.Detect(tmpDir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Terraform == nil {
		t.Errorf("expected Terraform config to be detected in subfolder %s", iacDir)
	} else {
		if cfg.Terraform.Dir != "my-project/.iac" {
			t.Errorf("expected Dir to be my-project/.iac, got %q", cfg.Terraform.Dir)
		}
	}

	// Case 2: Deep terraform files inside .iac
	tmpDir2, _ := os.MkdirTemp("", "cleat-tf-deep-*")
	defer os.RemoveAll(tmpDir2)
	iacDir2 := filepath.Join(tmpDir2, ".iac")
	os.MkdirAll(filepath.Join(iacDir2, "staging", "terraform"), 0755)
	os.WriteFile(filepath.Join(iacDir2, "staging", "terraform", "main.tf"), []byte(""), 0644)

	cfg2 := &schema.Config{}
	err = d.Detect(tmpDir2, cfg2)
	if err != nil {
		t.Fatal(err)
	}

	if cfg2.Terraform == nil {
		t.Fatal("expected Terraform config to be detected")
	}

	foundStaging := false
	for _, env := range cfg2.Terraform.Envs {
		if env == "staging" {
			foundStaging = true
			break
		}
	}
	if !foundStaging {
		t.Errorf("expected staging env to be detected from deep .tf file, got %v", cfg2.Terraform.Envs)
	}
}

func TestPythonPackageManagerDetection(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		svcDir   string
		expected string
	}{
		{
			name:     "default to uv",
			files:    []string{"manage.py"},
			expected: "uv",
		},
		{
			name:     "detect uv.lock",
			files:    []string{"manage.py", "uv.lock"},
			expected: "uv",
		},
		{
			name:     "detect requirements.txt",
			files:    []string{"manage.py", "requirements.txt"},
			expected: "pip",
		},
		{
			name:     "detect poetry.lock",
			files:    []string{"manage.py", "poetry.lock"},
			expected: "poetry",
		},
		{
			name:     "uv.lock takes priority over requirements.txt",
			files:    []string{"manage.py", "uv.lock", "requirements.txt"},
			expected: "uv",
		},
		{
			name:     "detect requirements.txt in project root for nested service",
			files:    []string{"requirements.txt", "services/api/manage.py"},
			svcDir:   "services/api",
			expected: "pip",
		},
		{
			name:     "service root takes priority over project root",
			files:    []string{"uv.lock", "services/api/manage.py", "services/api/requirements.txt"},
			svcDir:   "services/api",
			expected: "pip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, _ := os.MkdirTemp("", "cleat-python-pm-*")
			defer os.RemoveAll(tmpDir)

			for _, f := range tt.files {
				path := filepath.Join(tmpDir, f)
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, []byte(""), 0644)
			}

			svcDir := tt.svcDir
			if svcDir == "" {
				svcDir = "."
			}

			d := &DjangoDetector{}
			cfg := &schema.Config{
				Services: []schema.ServiceConfig{
					{Name: "backend", Dir: svcDir},
				},
			}
			err := d.Detect(tmpDir, cfg)
			if err != nil {
				t.Fatal(err)
			}

			pm := ""
			for _, mod := range cfg.Services[0].Modules {
				if mod.Python != nil {
					pm = mod.Python.PackageManager
					break
				}
			}

			if pm != tt.expected {
				t.Errorf("expected package manager %q, got %q", tt.expected, pm)
			}
		})
	}
}

func TestDualStackDetector(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-dual-stack-*")
	defer os.RemoveAll(tmpDir)

	// Create both Python and NPM markers
	os.WriteFile(filepath.Join(tmpDir, "manage.py"), []byte(""), 0644)
	packageJson := `{"scripts": {"build": "vite build"}}`
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJson), 0644)

	cfg := &schema.Config{
		Services: []schema.ServiceConfig{
			{Name: "foo", Dir: "."},
		},
	}

	// Run both detectors
	django := &DjangoDetector{}
	if err := django.Detect(tmpDir, cfg); err != nil {
		t.Fatal(err)
	}

	npm := &NpmDetector{}
	if err := npm.Detect(tmpDir, cfg); err != nil {
		t.Fatal(err)
	}

	svc := cfg.Services[0]
	hasPython := false
	hasNpm := false

	for _, mod := range svc.Modules {
		if mod.Python != nil {
			hasPython = true
		}
		if mod.Npm != nil {
			hasNpm = true
		}
	}

	if !hasPython {
		t.Error("expected Python module to be detected")
	}
	if !hasNpm {
		t.Error("expected NPM module to be detected")
	}

	if len(svc.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(svc.Modules))
	}
}

func TestDualStackDetector_NoServices(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-dual-stack-no-svc-*")
	defer os.RemoveAll(tmpDir)

	// Create both Python and NPM markers at root
	os.WriteFile(filepath.Join(tmpDir, "manage.py"), []byte(""), 0644)
	packageJson := `{"scripts": {"build": "vite build"}}`
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJson), 0644)

	cfg := &schema.Config{}

	// Run all detectors via DetectAll
	if err := DetectAll(tmpDir, cfg); err != nil {
		t.Fatal(err)
	}

	if len(cfg.Services) == 0 {
		t.Fatal("expected at least one service to be created")
	}

	svc := cfg.Services[0]
	hasPython := false
	hasNpm := false

	for _, mod := range svc.Modules {
		if mod.Python != nil {
			hasPython = true
		}
		if mod.Npm != nil {
			hasNpm = true
		}
	}

	if !hasPython {
		t.Error("expected Python module to be detected")
	}
	if !hasNpm {
		t.Error("expected NPM module to be detected")
	}
}

func TestDetectionInDockerContext_AmbiguousNames(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-docker-ambiguous-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Mkdir(filepath.Join(tmpDir, "app"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "app", "manage.py"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "app", "package.json"), []byte(`{"scripts":{"test":"echo"}}`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "app", "Dockerfile.py"), []byte("FROM python:3.9\nCOPY requirements.txt .\nRUN pip install -r requirements.txt"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "app", "Dockerfile.js"), []byte("FROM node:16\nCOPY package.json .\nRUN npm install"), 0644)

	dc := `
version: '3'
services:
  s1:
    build:
      context: ./app
      dockerfile: Dockerfile.py
  s2:
    build:
      context: ./app
      dockerfile: Dockerfile.js
  s3:
    build:
      context: ./app
    command: python manage.py runserver
  s4:
    build:
      context: ./app
    command: ["npm", "start"]
`
	os.WriteFile(filepath.Join(tmpDir, "docker-compose.yaml"), []byte(dc), 0644)

	cfg := &schema.Config{}
	err = DetectAll(tmpDir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Services) != 4 {
		t.Errorf("expected 4 services, got %d", len(cfg.Services))
	}

	for _, svc := range cfg.Services {
		hasPython := false
		hasNpm := false
		for _, mod := range svc.Modules {
			if mod.Python != nil {
				hasPython = true
			}
			if mod.Npm != nil {
				hasNpm = true
			}
		}

		switch svc.Name {
		case "s1", "s3":
			if !hasPython {
				t.Errorf("expected Python module for %s", svc.Name)
			}
			if hasNpm {
				t.Errorf("did NOT expect NPM module for %s", svc.Name)
			}
		case "s2", "s4":
			if !hasNpm {
				t.Errorf("expected NPM module for %s", svc.Name)
			}
			if hasPython {
				t.Errorf("did NOT expect Python module for %s", svc.Name)
			}
		}
	}
}

func TestRootDetectionWithDockerServices(t *testing.T) {
	// Root of the test project
	baseDir := "../../testdata/fixtures/docker-compose-with-multiple-services-in-same-dir"

	cfg := &schema.Config{}
	err := DetectAll(baseDir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Should have 2 services: 'backend' and 'backend-node' (from docker)
	if len(cfg.Services) != 2 {
		t.Errorf("expected 2 services, got %d: %+v", len(cfg.Services), cfg.Services)
	}

	foundBackend := false
	foundBackendNode := false
	for _, svc := range cfg.Services {
		if svc.Name == "backend" {
			foundBackend = true
			if svc.Dir != "./backend" && svc.Dir != "backend" {
				t.Errorf("expected backend service dir './backend' or 'backend', got %q", svc.Dir)
			}
			// Should have Python module but NOT NPM (due to Dockerfile hints)
			hasPython := false
			hasNpm := false
			for _, mod := range svc.Modules {
				if mod.Python != nil {
					hasPython = true
				}
				if mod.Npm != nil {
					hasNpm = true
				}
			}
			if !hasPython {
				t.Error("expected Python module on backend service")
			}
			if hasNpm {
				t.Error("did NOT expect NPM module on backend service")
			}
		} else if svc.Name == "backend-node" {
			foundBackendNode = true
			// Should have NPM module but NOT Python
			hasPython := false
			hasNpm := false
			for _, mod := range svc.Modules {
				if mod.Python != nil {
					hasPython = true
				}
				if mod.Npm != nil {
					hasNpm = true
				}
			}
			if !hasNpm {
				t.Error("expected NPM module on backend-node service")
			}
			if hasPython {
				t.Error("did NOT expect Python module on backend-node service")
			}
		}
	}

	if !foundBackend {
		t.Error("service 'backend' not found")
	}
	if !foundBackendNode {
		t.Error("service 'backend-node' not found")
	}
}
