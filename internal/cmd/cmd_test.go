package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/spf13/cobra"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()
	return buf.String(), err
}

func TestSubcommands(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-cmd-test-*")
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile("cleat.yaml", []byte(`
version: 1
docker: true
google_cloud_platform:
  project_name: test-proj
services:
  - name: web
    dir: .
    modules:
      - python:
          django: true
      - npm:
          scripts: ["build"]
`), 0644)

	tests := []struct {
		name string
		args []string
	}{
		{"build", []string{"build"}},
		{"run", []string{"run"}},
		{"version", []string{"version"}},
		{"docker down", []string{"docker", "down"}},
		{"docker rebuild", []string{"docker", "rebuild"}},
		{"docker remove-orphans", []string{"docker", "remove-orphans"}},
		{"docker down svc", []string{"docker", "down", "web"}},
		{"docker rebuild svc", []string{"docker", "rebuild", "web"}},
		{"docker remove-orphans svc", []string{"docker", "remove-orphans", "web"}},
		{"django migrate", []string{"django", "migrate"}},
		{"django makemigrations", []string{"django", "makemigrations"}},
		{"django collectstatic", []string{"django", "collectstatic"}},
		{"django create-user-dev", []string{"django", "create-user-dev"}},
		{"django gen-random-secret-key", []string{"django", "gen-random-secret-key"}},
		{"django migrate svc", []string{"django", "migrate", "web"}},
		{"django makemigrations svc", []string{"django", "makemigrations", "web"}},
		{"django collectstatic svc", []string{"django", "collectstatic", "web"}},
		{"django create-user-dev svc", []string{"django", "create-user-dev", "web"}},
		{"django gen-random-secret-key svc", []string{"django", "gen-random-secret-key", "web"}},
		{"gcp activate", []string{"gcp", "activate"}},
		{"gcp init", []string{"gcp", "init"}},
		{"gcp set-config", []string{"gcp", "set-config"}},
		{"gcp adc-login", []string{"gcp", "adc-login"}},
		{"gcp adc-impersonate-login", []string{"gcp", "adc-impersonate-login"}},
		{"gcp console", []string{"gcp", "console"}},
		{"npm install", []string{"npm", "install"}},
		{"npm build svc", []string{"npm", "build", "web"}},
		{"terraform init", []string{"terraform", "init"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _ = executeCommand(rootCmd, tt.args...)
		})
	}
}

func TestTerraformSubcommands(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-tf-test-*")
	defer os.RemoveAll(tmpDir)
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cfgContent := `
version: 1
terraform:
  envs: ["prod"]
`
	os.WriteFile("cleat.yaml", []byte(cfgContent), 0644)

	tfTests := []struct {
		name string
		args []string
	}{
		{"init", []string{"terraform", "init", "prod"}},
		{"init-upgrade", []string{"terraform", "init-upgrade", "prod"}},
		{"plan", []string{"terraform", "plan", "prod"}},
		{"apply", []string{"terraform", "apply", "prod"}},
		{"apply-refresh", []string{"terraform", "apply-refresh", "prod"}},
	}

	for _, tt := range tfTests {
		t.Run(tt.name, func(t *testing.T) {
			_, _ = executeCommand(rootCmd, tt.args...)
		})
	}
}

func TestGCPAppEngineSubcommands(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-gcp-ae-test-*")
	defer os.RemoveAll(tmpDir)
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cfgContent := `
version: 1
google_cloud_platform:
  project_name: test-proj
app_yaml: app.yaml
services:
  - name: backend
    app_yaml: backend/app.yaml
`
	os.WriteFile("cleat.yaml", []byte(cfgContent), 0644)
	os.WriteFile("app.yaml", []byte("runtime: python39"), 0644)
	os.Mkdir("backend", 0755)
	os.WriteFile("backend/app.yaml", []byte("runtime: python39"), 0644)

	gcpTests := []struct {
		name string
		args []string
	}{
		{"deploy", []string{"gcp", "app-engine", "deploy"}},
		{"deploy svc", []string{"gcp", "app-engine", "deploy", "backend"}},
		{"promote", []string{"gcp", "app-engine", "promote"}},
		{"promote svc", []string{"gcp", "app-engine", "promote", "backend"}},
	}

	for _, tt := range gcpTests {
		t.Run(tt.name, func(t *testing.T) {
			_, _ = executeCommand(rootCmd, tt.args...)
		})
	}
}

func TestCreateSessionAndMerge(t *testing.T) {
	cfg := &config.Config{}
	preCollectedInputs = map[string]string{"foo": "bar"}
	sess := createSessionAndMerge(cfg)
	if sess.Inputs["foo"] != "bar" {
		t.Errorf("expected input foo=bar, got %v", sess.Inputs["foo"])
	}

	// Test with nil preCollectedInputs
	preCollectedInputs = nil
	sess = createSessionAndMerge(cfg)
	if sess == nil {
		t.Fatal("expected non-nil session")
	}
}
