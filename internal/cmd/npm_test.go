package cmd

import (
	"testing"
)

func TestNpmCmd(t *testing.T) {
	if npmCmd.Use != "npm [script] [service]" {
		t.Errorf("expected npmCmd.Use to be 'npm [script] [service]', got %s", npmCmd.Use)
	}

	if npmCmd.Short == "" {
		t.Error("expected npmCmd.Short to be non-empty")
	}

	// Verify it's registered as a subcommand
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "npm [script] [service]" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected npm command to be registered with root")
	}
}
