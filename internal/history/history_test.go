package history

import (
	"os"
	"path/filepath"
	"strings"
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

func TestHistoryProjectRoot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-history-root-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a project structure
	// tmpDir/ (root)
	//   cleat.yaml
	//   cmd/
	projectRoot := tmpDir
	subDir := filepath.Join(projectRoot, "cmd")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(projectRoot, "cleat.yaml"), []byte("version: 1"), 0644)

	// Mock home directory
	oldUserHomeDir := userHomeDir
	userHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { userHomeDir = oldUserHomeDir }()

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// 1. Save from root
	os.Chdir(projectRoot)
	entry1 := HistoryEntry{Command: "root command", Timestamp: time.Now()}
	Save(entry1)

	// 2. Save from subDir
	os.Chdir(subDir)
	entry2 := HistoryEntry{Command: "subdir command", Timestamp: time.Now()}
	Save(entry2)

	// Check how many history files are in tmpDir/.cleat
	files, _ := os.ReadDir(filepath.Join(tmpDir, ".cleat"))
	var historyFiles []string
	for _, f := range files {
		if !f.IsDir() {
			historyFiles = append(historyFiles, f.Name())
		}
	}

	if len(historyFiles) > 1 {
		t.Errorf("Expected 1 history file, got %d: %v", len(historyFiles), historyFiles)
	}

	// Try to load from subDir and see if it has root command
	entries, _ := Load()
	foundRootCmd := false
	for _, e := range entries {
		if e.Command == "root command" {
			foundRootCmd = true
			break
		}
	}
	if !foundRootCmd {
		t.Error("Did not find root command when loading from subdir")
	}
}

func TestHistoryGitRoot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-history-git-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a project structure
	// tmpDir/ (root)
	//   .git/
	//   cmd/
	projectRoot := tmpDir
	subDir := filepath.Join(projectRoot, "cmd")
	os.MkdirAll(subDir, 0755)
	os.MkdirAll(filepath.Join(projectRoot, ".git"), 0755)

	// Mock home directory
	oldUserHomeDir := userHomeDir
	userHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { userHomeDir = oldUserHomeDir }()

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// 1. Save from subDir
	os.Chdir(subDir)
	entry1 := HistoryEntry{Command: "subdir command", Timestamp: time.Now()}
	Save(entry1)

	// 2. Load and verify identity
	files, _ := os.ReadDir(filepath.Join(tmpDir, ".cleat"))
	if len(files) != 1 {
		t.Errorf("Expected 1 history file, got %d", len(files))
	}

	// The filename should start with the tmpDir base name, not "cmd"
	expectedPrefix := filepath.Base(projectRoot)
	if !strings.HasPrefix(files[0].Name(), expectedPrefix) {
		t.Errorf("Expected filename to start with %s, got %s", expectedPrefix, files[0].Name())
	}
}
