package history

import (
	"os"
	"testing"
	"time"
)

func TestHistory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-history-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock home directory
	oldUserHomeDir := userHomeDir
	userHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { userHomeDir = oldUserHomeDir }()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	entry1 := HistoryEntry{
		Timestamp: time.Now(),
		Command:   "test command 1",
		Inputs:    map[string]string{"key": "val"},
	}

	err = Save(entry1)
	if err != nil {
		t.Fatalf("Failed to save history: %v", err)
	}

	entries, err := Load()
	if err != nil {
		t.Fatalf("Failed to load history: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Command != "test command 1" {
		t.Errorf("Expected command 'test command 1', got '%s'", entries[0].Command)
	}

	// Test max size
	for i := 0; i < 60; i++ {
		Save(HistoryEntry{Command: "cmd", Timestamp: time.Now()})
	}

	entries, _ = Load()
	if len(entries) > maxHistorySize {
		t.Errorf("Expected max %d entries, got %d", maxHistorySize, len(entries))
	}
}
