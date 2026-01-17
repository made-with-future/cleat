package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
)

func TestBuildProject(t *testing.T) {
	var executedCommands []string

	// Mock runner
	runner = func(name string, args ...string) error {
		executedCommands = append(executedCommands, name+" "+strings.Join(args, " "))
		return nil
	}
	defer func() { runner = runCommand }() // Restore original runner

	t.Run("Django and Docker", func(t *testing.T) {
		executedCommands = nil
		cfg := &config.Config{
			Docker:        true,
			Django:        true,
			DjangoService: "web",
		}
		err := buildProject(cfg)
		if err != nil {
			t.Fatalf("buildProject failed: %v", err)
		}

		expected := []string{
			"docker compose run --rm web python manage.py collectstatic --noinput",
			"docker compose build",
		}

		if len(executedCommands) != len(expected) {
			t.Fatalf("Expected %d commands, got %d", len(expected), len(executedCommands))
		}

		for i, cmd := range expected {
			if executedCommands[i] != cmd {
				t.Errorf("Expected command %d to be '%s', got '%s'", i, cmd, executedCommands[i])
			}
		}
	})

	t.Run("Django only", func(t *testing.T) {
		executedCommands = nil
		cfg := &config.Config{
			Docker: false,
			Django: true,
		}
		err := buildProject(cfg)
		if err != nil {
			t.Fatalf("buildProject failed: %v", err)
		}

		expected := []string{
			"python manage.py collectstatic --noinput",
		}

		if len(executedCommands) != len(expected) {
			t.Fatalf("Expected %d commands, got %d", len(expected), len(executedCommands))
		}

		if executedCommands[0] != expected[0] {
			t.Errorf("Expected command to be '%s', got '%s'", expected[0], executedCommands[0])
		}
	})

	t.Run("Docker only", func(t *testing.T) {
		executedCommands = nil
		cfg := &config.Config{
			Docker: true,
			Django: false,
		}
		err := buildProject(cfg)
		if err != nil {
			t.Fatalf("buildProject failed: %v", err)
		}

		expected := []string{
			"docker compose build",
		}

		if len(executedCommands) != len(expected) {
			t.Fatalf("Expected %d commands, got %d", len(expected), len(executedCommands))
		}

		if executedCommands[0] != expected[0] {
			t.Errorf("Expected command to be '%s', got '%s'", expected[0], executedCommands[0])
		}
	})

	t.Run("NPM and Docker", func(t *testing.T) {
		executedCommands = nil
		cfg := &config.Config{
			Docker: true,
			Npm: config.NpmConfig{
				Service: "node",
				Scripts: []string{"build", "css"},
			},
		}
		err := buildProject(cfg)
		if err != nil {
			t.Fatalf("buildProject failed: %v", err)
		}

		expected := []string{
			"docker compose run --rm node npm run build",
			"docker compose run --rm node npm run css",
			"docker compose build",
		}

		if len(executedCommands) != len(expected) {
			t.Fatalf("Expected %d commands, got %d", len(expected), len(executedCommands))
		}

		for i, cmd := range expected {
			if executedCommands[i] != cmd {
				t.Errorf("Expected command %d to be '%s', got '%s'", i, cmd, executedCommands[i])
			}
		}
	})

	t.Run("NPM local", func(t *testing.T) {
		executedCommands = nil
		cfg := &config.Config{
			Docker: false,
			Npm: config.NpmConfig{
				Scripts: []string{"build"},
			},
		}
		err := buildProject(cfg)
		if err != nil {
			t.Fatalf("buildProject failed: %v", err)
		}

		expected := []string{
			"npm run build",
		}

		if len(executedCommands) != len(expected) {
			t.Fatalf("Expected %d commands, got %d", len(expected), len(executedCommands))
		}

		if executedCommands[0] != expected[0] {
			t.Errorf("Expected command to be '%s', got '%s'", expected[0], executedCommands[0])
		}
	})

	t.Run("Opinionated Django and NPM local", func(t *testing.T) {
		// This test needs to create folders and files
		tmpDir, err := os.MkdirTemp("", "cleat-opinion-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		os.Mkdir("backend", 0755)
		os.WriteFile("backend/manage.py", []byte(""), 0644)
		os.Mkdir("frontend", 0755)
		os.WriteFile("frontend/package.json", []byte("{}"), 0644)

		executedCommands = nil
		cfg := &config.Config{
			Docker: false,
			Django: true,
			Npm: config.NpmConfig{
				Scripts: []string{"build"},
			},
		}
		err = buildProject(cfg)
		if err != nil {
			t.Fatalf("buildProject failed: %v", err)
		}

		expected := []string{
			"npm --prefix frontend run build",
			"python backend/manage.py collectstatic --noinput",
		}

		if len(executedCommands) != len(expected) {
			t.Fatalf("Expected %d commands, got %d. Commands: %v", len(expected), len(executedCommands), executedCommands)
		}

		for i, cmd := range expected {
			if executedCommands[i] != cmd {
				t.Errorf("Expected command %d to be '%s', got '%s'", i, cmd, executedCommands[i])
			}
		}
	})
}
