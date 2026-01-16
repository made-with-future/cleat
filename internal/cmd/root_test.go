package cmd

import (
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
