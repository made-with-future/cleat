package cmd

import (
	"testing"
)

func TestBuildCmd(t *testing.T) {
	if buildCmd.Use != "build" {
		t.Errorf("expected buildCmd.Use to be 'build', got %s", buildCmd.Use)
	}

	if buildCmd.Short == "" {
		t.Error("expected buildCmd.Short to be non-empty")
	}

	// Verify it's registered as a subcommand
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "build" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected build command to be registered with root")
	}
}
