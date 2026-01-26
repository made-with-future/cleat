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
		waitCalls = 0
		// version command is safe to run as it just prints
		run([]string{"cleat", "version"})
		if waitCalls != 0 {
			t.Errorf("expected Wait to be called 0 times for CLI mode, got %d", waitCalls)
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
		if waitCalls != 1 {
			t.Errorf("expected Wait to be called 1 time, got %d", waitCalls)
		}
	})

	t.Run("Loop for gcp init command", func(t *testing.T) {
		waitCalls = 0
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
		if waitCalls != 1 {
			t.Errorf("expected Wait to be called 1 time, got %d", waitCalls)
		}
	})

	t.Run("Loop for build command", func(t *testing.T) {
		waitCalls = 0
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
		if waitCalls != 1 {
			t.Errorf("expected Wait to be called 1 time, got %d", waitCalls)
		}
	})

	t.Run("Re-run command with 'r'", func(t *testing.T) {
		waitCalls = 0
		uiCalls := 0
		UIStart = func() (string, map[string]string, error) {
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
		// Wait should be called twice: once after first run (returned true), once after second run (returned false)
		if waitCalls != 2 {
			t.Errorf("expected Wait to be called 2 times, got %d", waitCalls)
		}
	})
}
