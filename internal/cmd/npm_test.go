package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/madewithfuture/cleat/internal/task"
)

func TestNpmRunCommand(t *testing.T) {
	// Mock runner
	oldRunner := task.CommandRunner
	task.CommandRunner = func(name string, args ...string) error {
		return nil
	}
	defer func() { task.CommandRunner = oldRunner }()

	tmpDir, err := os.MkdirTemp("", "cleat-npm-run-cmd-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	t.Run("No config", func(t *testing.T) {
		rootCmd.SetArgs([]string{"npm-run", "test"})
		err := rootCmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "no cleat.yaml found") {
			t.Errorf("expected error about missing cleat.yaml, got %v", err)
		}
	})

	t.Run("With config", func(t *testing.T) {
		os.WriteFile("cleat.yaml", []byte("npm:\n  scripts:\n    - test"), 0644)
		rootCmd.SetArgs([]string{"npm-run", "test"})
		err := rootCmd.Execute()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("Missing script arg", func(t *testing.T) {
		rootCmd.SetArgs([]string{"npm-run"})
		err := rootCmd.Execute()
		if err == nil {
			t.Error("expected error for missing argument")
		}
	})
}
