package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/task"
)

func TestRunCommand(t *testing.T) {
	var executedCommands []string

	// Mock runner
	oldRunner := task.CommandRunner
	task.CommandRunner = func(name string, args ...string) error {
		executedCommands = append(executedCommands, name+" "+strings.Join(args, " "))
		return nil
	}
	defer func() { task.CommandRunner = oldRunner }()

	t.Run("Docker run", func(t *testing.T) {
		executedCommands = nil
		cfg := &config.Config{
			Docker: true,
		}

		err := task.Run(cfg)
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}

		expected := "docker compose up --remove-orphans"
		if len(executedCommands) != 1 || executedCommands[0] != expected {
			t.Errorf("Expected '%s', got %v", expected, executedCommands)
		}
	})

	t.Run("Docker run with op", func(t *testing.T) {
		executedCommands = nil
		tmpDir, err := os.MkdirTemp("", "cleat-run-op-test-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		os.Mkdir(".env", 0755)
		err = os.WriteFile(".env/dev.env", []byte("FOO=BAR"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		cfg := &config.Config{
			Docker: true,
		}
		err = task.Run(cfg)
		if err != nil {
			t.Fatalf("task.Run failed: %v", err)
		}

		expected := "op run --env-file ./.env/dev.env -- docker compose up --remove-orphans"
		if len(executedCommands) != 1 || executedCommands[0] != expected {
			t.Errorf("Expected '%s', got %v", expected, executedCommands)
		}
	})

	t.Run("Django local run", func(t *testing.T) {
		executedCommands = nil
		cfg := &config.Config{
			Django: true,
			Docker: false,
		}
		err := task.Run(cfg)
		if err != nil {
			t.Fatalf("task.Run failed: %v", err)
		}

		expected := "python manage.py runserver"
		if len(executedCommands) != 1 || executedCommands[0] != expected {
			t.Errorf("Expected '%s', got %v", expected, executedCommands)
		}
	})

	t.Run("NPM local run", func(t *testing.T) {
		executedCommands = nil
		cfg := &config.Config{
			Docker: false,
			Npm: config.NpmConfig{
				Scripts: []string{"build"},
			},
		}
		err := task.Run(cfg)
		if err != nil {
			t.Fatalf("task.Run failed: %v", err)
		}

		expected := "npm start"
		if len(executedCommands) != 1 || executedCommands[0] != expected {
			t.Errorf("Expected '%s', got %v", expected, executedCommands)
		}
	})
}
