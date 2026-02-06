package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/madewithfuture/cleat/internal/history"
)

func TestRootCmd(t *testing.T) {
	if rootCmd.Use != "cleat" {
		t.Errorf("expected rootCmd.Use to be 'cleat', got %s", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("expected rootCmd.Short to be non-empty")
	}
}

func TestRun(t *testing.T) {
	// Save original values
	oldUIStart := UIStart
	oldExit := Exit
	oldWait := Wait
	defer func() {
		UIStart = oldUIStart
		Exit = oldExit
		Wait = oldWait
	}()

	var exitCode int
	Exit = func(code int) {
		exitCode = code
	}

	var waitCalls int
	Wait = func() bool {
		waitCalls++
		return false
	}

	t.Run("No args, TUI returns build", func(t *testing.T) {
		calls := 0
		UIStart = func(string) (string, map[string]string, error) {
			calls++
			if calls == 1 {
				return "build", nil, nil
			}
			return "", nil, nil
		}
		// We need to prevent actual task execution if possible, or just check if buildCmd was triggered
		// Since we don't have a clean way to mock the command implementation without more refactoring,
		// we can at least check if it tries to run.

		// Actually, buildCmd will try to load cleat.yaml and fail if it doesn't exist.
		// Let's just verify it doesn't crash and we can maybe see if it was called.

		run([]string{"cleat"})
	})

	t.Run("No args, TUI returns error", func(t *testing.T) {
		exitCode = 0
		UIStart = func(string) (string, map[string]string, error) {
			return "", nil, errors.New("TUI error")
		}
		run([]string{"cleat"})
		if exitCode != 1 {
			t.Errorf("expected exit code 1, got %d", exitCode)
		}
	})

	t.Run("No args, TUI returns empty (quit)", func(t *testing.T) {
		exitCode = 0
		UIStart = func(string) (string, map[string]string, error) {
			return "", nil, nil
		}
		run([]string{"cleat"})
		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}
	})

	t.Run("No args, TUI returns npm run", func(t *testing.T) {
		calls := 0
		UIStart = func(string) (string, map[string]string, error) {
			calls++
			if calls == 1 {
				return "npm run test", nil, nil
			}
			return "", nil, nil
		}
		// Should set args to [npm-run test]
		run([]string{"cleat"})
	})

	t.Run("No args, TUI returns gcp init", func(t *testing.T) {
		calls := 0
		UIStart = func(string) (string, map[string]string, error) {
			calls++
			if calls == 1 {
				return "gcp init", nil, nil
			}
			return "", nil, nil
		}
		run([]string{"cleat"})
	})

	t.Run("No args, TUI returns gcp console", func(t *testing.T) {
		calls := 0
		UIStart = func(string) (string, map[string]string, error) {
			calls++
			if calls == 1 {
				return "gcp console", nil, nil
			}
			return "", nil, nil
		}
		run([]string{"cleat"})
	})

	t.Run("No args, TUI returns gcp promote", func(t *testing.T) {
		calls := 0
		UIStart = func(string) (string, map[string]string, error) {
			calls++
			if calls == 1 {
				return "gcp app-engine promote:svc", nil, nil
			}
			return "", nil, nil
		}
		run([]string{"cleat"})
	})

	t.Run("With args, executes subcommand", func(t *testing.T) {
		waitCalls = 0
		// version command is safe to run as it just prints
		run([]string{"cleat", "version"})
		if waitCalls != 0 {
			t.Errorf("expected Wait to be called 0 times for CLI mode, got %d", waitCalls)
		}
	})

	t.Run("TUI returns workflow, tracks stats", func(t *testing.T) {
		// Setup temp dirs
		tmpDir, err := os.MkdirTemp("", "cleat-stats-workflow-test-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		// Mock home directory for history package
		oldUserHomeDir := history.UserHomeDir
		history.UserHomeDir = func() (string, error) {
			return tmpDir, nil
		}
		defer func() { history.UserHomeDir = oldUserHomeDir }()

		// setup project dir with cleat.yaml and a workflow
		projectDir := filepath.Join(tmpDir, "project")
		os.Mkdir(projectDir, 0755)
		workflowContent := `
workflows:
- name: my-workflow
  commands:
    - build
`
		os.WriteFile(filepath.Join(projectDir, "cleat.yaml"), []byte(workflowContent), 0644)
		oldWd, _ := os.Getwd()
		os.Chdir(projectDir)
		defer os.Chdir(oldWd)

		uiCalls := 0
		UIStart = func(string) (string, map[string]string, error) {
			uiCalls++
			if uiCalls == 1 {
				return "workflow:my-workflow", nil, nil
			}
			return "", nil, nil // Quit on second call
		}

		// This will just run the first command of the workflow and then loop
		// It's tricky to test the full workflow run without more refactoring
		// We just want to check if stats are updated.
		run([]string{"cleat"})

		// check stats file
		stats, err := history.LoadStats()
		if err != nil {
			t.Fatalf("Failed to load stats: %v", err)
		}

		if stats.Commands["workflow:my-workflow"].Count != 1 {
			t.Errorf("Expected workflow:my-workflow count 1, got %d", stats.Commands["workflow:my-workflow"].Count)
		}

		// The inner command should NOT be tracked
		if _, ok := stats.Commands["build"]; ok {
			t.Errorf("Did not expect 'build' command to be tracked in stats for a workflow run")
		}
	})
}

func TestRunLoop(t *testing.T) {
	oldUIStart := UIStart
	oldExit := Exit
	oldWait := Wait
	defer func() {
		UIStart = oldUIStart
		Exit = oldExit
		Wait = oldWait
	}()

	Exit = func(code int) {}

	var waitCalls int
	Wait = func() bool {
		waitCalls++
		return false
	}

	t.Run("Loop for run command", func(t *testing.T) {
		waitCalls = 0
		calls := 0
		UIStart = func(string) (string, map[string]string, error) {
			calls++
			if calls == 1 {
				return "run", nil, nil
			}
			return "", nil, nil // Quit on second call
		}

		// Mocking execute is hard, but we can verify calls to UIStart
		// We expect UIStart to be called twice because we always loop back in TUI mode
		run([]string{"cleat"})

		if calls != 2 {
			t.Errorf("expected UIStart to be called 2 times, got %d", calls)
		}
		if waitCalls != 1 {
			t.Errorf("expected Wait to be called 1 time, got %d", waitCalls)
		}
	})

	t.Run("Loop for gcp init command", func(t *testing.T) {
		waitCalls = 0
		calls := 0
		UIStart = func(string) (string, map[string]string, error) {
			calls++
			if calls == 1 {
				return "gcp init", nil, nil
			}
			return "", nil, nil // Quit on second call
		}

		run([]string{"cleat"})

		if calls != 2 {
			t.Errorf("expected UIStart to be called 2 times for gcp init, got %d", calls)
		}
		if waitCalls != 1 {
			t.Errorf("expected Wait to be called 1 time, got %d", waitCalls)
		}
	})

	t.Run("Loop for build command", func(t *testing.T) {
		waitCalls = 0
		calls := 0
		UIStart = func(string) (string, map[string]string, error) {
			calls++
			if calls == 1 {
				return "build", nil, nil
			}
			return "", nil, nil
		}

		run([]string{"cleat"})

		if calls != 2 {
			t.Errorf("expected UIStart to be called 2 times, got %d", calls)
		}
		if waitCalls != 1 {
			t.Errorf("expected Wait to be called 1 time, got %d", waitCalls)
		}
	})

	t.Run("Re-run command with 'r'", func(t *testing.T) {
		waitCalls = 0
		uiCalls := 0
		UIStart = func(string) (string, map[string]string, error) {
			uiCalls++
			if uiCalls == 1 {
				return "build", nil, nil
			}
			return "", nil, nil
		}

		Wait = func() bool {
			waitCalls++
			if waitCalls == 1 {
				return true // Re-run
			}
			return false // Back to TUI
		}

		run([]string{"cleat"})

		// UIStart should be called twice: once for "build", once for "" (to quit)
		if uiCalls != 2 {
			t.Errorf("expected UIStart to be called 2 times, got %d", uiCalls)
		}
		if waitCalls != 2 {
			t.Errorf("expected Wait to be called 2 times, got %d", waitCalls)
		}
	})
}

func TestMapSelectedToArgs(t *testing.T) {
	tests := []struct {
		selected string
		want     []string
	}{
		{"build", []string{"build"}},
		{"run", []string{"run"}},
		{"workflow:test", []string{"workflow", "test"}},
		{"docker up", []string{"docker", "up"}},
		{"docker up:svc", []string{"docker", "up", "svc"}},
		{"django migrate", []string{"django", "migrate"}},
		{"django migrate:svc", []string{"django", "migrate", "svc"}},
		{"npm run dev", []string{"npm", "dev"}},
		{"npm run test:svc", []string{"npm", "svc", "test"}},
		{"npm install:svc", []string{"npm", "install", "svc"}},
		{"gcp init", []string{"gcp", "init"}},
		{"terraform plan", []string{"terraform", "plan"}},
		{"terraform plan:prod", []string{"terraform", "plan", "prod"}},
		{"unknown", nil},
	}

	for _, tt := range tests {
		got := mapSelectedToArgs(tt.selected)
		if len(got) != len(tt.want) {
			t.Errorf("mapSelectedToArgs(%q) len = %d, want %d", tt.selected, len(got), len(tt.want))
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("mapSelectedToArgs(%q) [%d] = %q, want %q", tt.selected, i, got[i], tt.want[i])
			}
		}
	}
}

func TestWaitForAnyKey(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	t.Run("Return 'r'", func(t *testing.T) {
		r, w, _ := os.Pipe()
		os.Stdin = r
		w.Write([]byte("r"))
		w.Close()
		if !waitForAnyKey() {
			t.Error("expected true for 'r'")
		}
	})

	t.Run("Return 'R'", func(t *testing.T) {
		r, w, _ := os.Pipe()
		os.Stdin = r
		w.Write([]byte("R"))
		w.Close()
		if !waitForAnyKey() {
			t.Error("expected true for 'R'")
		}
	})

	t.Run("Return other", func(t *testing.T) {
		r, w, _ := os.Pipe()
		os.Stdin = r
		w.Write([]byte("x"))
		w.Close()
		if waitForAnyKey() {
			t.Error("expected false for 'x'")
		}
	})
}
