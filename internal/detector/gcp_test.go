package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madewithfuture/cleat/internal/config/schema"
)

func TestGcpDetector(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "app.yaml"), []byte(""), 0644)

	cfg := &schema.Config{
		GoogleCloudPlatform: &schema.GCPConfig{ProjectName: "test"},
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

	

	func TestGcpDetector_NilConfig(t *testing.T) {

		d := &GcpDetector{}

		cfg := &schema.Config{}

		err := d.Detect(".", cfg)

		if err != nil {

			t.Fatalf("Detect failed: %v", err)

		}

		if cfg.AppYaml != "" {

			t.Error("expected AppYaml to be empty")

		}

	}

	

	func TestGcpDetector_NestedServices(t *testing.T) {

		tmpDir, _ := os.MkdirTemp("", "cleat-gcp-nested-*")

		defer os.RemoveAll(tmpDir)

	

		// Create a service directory

		svcDir := filepath.Join(tmpDir, "web-service")

		os.Mkdir(svcDir, 0755)

		os.WriteFile(filepath.Join(svcDir, "app.yaml"), []byte(""), 0644)

	

		// Create a directory that should be ignored

		ignoredDir := filepath.Join(tmpDir, ".git")

		os.Mkdir(ignoredDir, 0755)

		os.WriteFile(filepath.Join(ignoredDir, "app.yaml"), []byte(""), 0644)

	

		cfg := &schema.Config{

			GoogleCloudPlatform: &schema.GCPConfig{ProjectName: "test"},

			Services: []schema.ServiceConfig{

				{Name: "web-service", Dir: "web-service"},

			},

		}

		d := &GcpDetector{}

		err := d.Detect(tmpDir, cfg)

		if err != nil {

			t.Fatal(err)

		}

	

		found := false

		for _, svc := range cfg.Services {

			if svc.Name == "web-service" {

				if svc.AppYaml != "web-service/app.yaml" {

					t.Errorf("expected AppYaml 'web-service/app.yaml', got %s", svc.AppYaml)

				}

				found = true

			}

			if svc.Name == ".git" {

				t.Error(".git directory should have been ignored")

			}

		}

		if !found {

			t.Error("expected web-service to be updated")

		}

	}

	

	func TestGcpDetector_NewServiceDiscovery(t *testing.T) {

		tmpDir, _ := os.MkdirTemp("", "cleat-gcp-discovery-*")

		defer os.RemoveAll(tmpDir)

	

		svcDir := filepath.Join(tmpDir, "new-service")

		os.Mkdir(svcDir, 0755)

		os.WriteFile(filepath.Join(svcDir, "app.yaml"), []byte(""), 0644)

	

		cfg := &schema.Config{

			GoogleCloudPlatform: &schema.GCPConfig{ProjectName: "test"},

		}

		d := &GcpDetector{}

		err := d.Detect(tmpDir, cfg)

		if err != nil {

			t.Fatal(err)

		}

	

		if len(cfg.Services) != 1 {

			t.Errorf("expected 1 service, got %d", len(cfg.Services))

		} else if cfg.Services[0].Name != "new-service" {

			t.Errorf("expected new-service, got %s", cfg.Services[0].Name)

		}

	}

	