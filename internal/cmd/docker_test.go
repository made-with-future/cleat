package cmd

import (
	"testing"
)

func TestDockerUpCmd(t *testing.T) {
	// This test verifies the docker up command is registered and can be accessed
	rootCmd.SetArgs([]string{"docker", "up", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("docker up --help failed: %v", err)
	}
}

func TestDockerDownCmd(t *testing.T) {
	// This test verifies the docker down command is registered
	rootCmd.SetArgs([]string{"docker", "down", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("docker down --help failed: %v", err)
	}
}

func TestDockerRebuildCmd(t *testing.T) {
	// This test verifies the docker rebuild command is registered
	rootCmd.SetArgs([]string{"docker", "rebuild", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("docker rebuild --help failed: %v", err)
	}
}

func TestDockerRemoveOrphansCmd(t *testing.T) {
	// This test verifies the docker remove-orphans command is registered
	rootCmd.SetArgs([]string{"docker", "remove-orphans", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("docker remove-orphans --help failed: %v", err)
	}
}
