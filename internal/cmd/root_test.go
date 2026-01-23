package cmd

import (
	"errors"
	"testing"
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
	defer func() {
		UIStart = oldUIStart
		Exit = oldExit
	}()

	var exitCode int
	Exit = func(code int) {
		exitCode = code
	}

	t.Run("No args, TUI returns build", func(t *testing.T) {
		calls := 0
		UIStart = func() (string, map[string]string, error) {
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
		UIStart = func() (string, map[string]string, error) {
			return "", nil, errors.New("TUI error")
		}
		run([]string{"cleat"})
		if exitCode != 1 {
			t.Errorf("expected exit code 1, got %d", exitCode)
		}
	})

	t.Run("No args, TUI returns empty (quit)", func(t *testing.T) {
		exitCode = 0
		UIStart = func() (string, map[string]string, error) {
			return "", nil, nil
		}
		run([]string{"cleat"})
		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}
	})

	t.Run("No args, TUI returns npm run", func(t *testing.T) {
		calls := 0
		UIStart = func() (string, map[string]string, error) {
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
		UIStart = func() (string, map[string]string, error) {
			calls++
			if calls == 1 {
				return "gcp init", nil, nil
			}
			return "", nil, nil
		}
		run([]string{"cleat"})
	})

	t.Run("With args, executes subcommand", func(t *testing.T) {
		// version command is safe to run as it just prints
		run([]string{"cleat", "version"})
	})
}

func TestRunLoop(t *testing.T) {
	oldUIStart := UIStart
	oldExit := Exit
	defer func() {
		UIStart = oldUIStart
		Exit = oldExit
	}()

	Exit = func(code int) {}

	t.Run("Loop for run command", func(t *testing.T) {
		calls := 0
		UIStart = func() (string, map[string]string, error) {
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
	})

	t.Run("Loop for gcp init command", func(t *testing.T) {
		calls := 0
		UIStart = func() (string, map[string]string, error) {
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
	})

	t.Run("Loop for build command", func(t *testing.T) {
		calls := 0
		UIStart = func() (string, map[string]string, error) {
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
	})
}
