package cmd

import (
	"testing"
)

func TestRunCmd(t *testing.T) {
	if runCmd.Use != "run" {
		t.Errorf("expected runCmd.Use to be 'run', got %s", runCmd.Use)
	}

	if runCmd.Short == "" {
		t.Error("expected runCmd.Short to be non-empty")
	}

	// Verify it's registered as a subcommand
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "run" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected run command to be registered with root")
	}
}
