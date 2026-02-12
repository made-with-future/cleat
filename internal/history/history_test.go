package history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/madewithfuture/cleat/internal/config"
)

func TestHistory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-history-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock home directory
	oldUserHomeDir := UserHomeDir
	UserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { UserHomeDir = oldUserHomeDir }()

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

	// Test Clear
	err = Clear()
	if err != nil {
		t.Fatalf("Failed to clear history: %v", err)
	}

	entries, err = Load()
	if err != nil {
		t.Errorf("Expected no error loading cleared history, got %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", len(entries))
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
	oldUserHomeDir := UserHomeDir
	UserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { UserHomeDir = oldUserHomeDir }()

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
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".history.yaml") {
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
	oldUserHomeDir := UserHomeDir
	UserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { UserHomeDir = oldUserHomeDir }()

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// 1. Save from subDir
	os.Chdir(subDir)
	entry1 := HistoryEntry{Command: "subdir command", Timestamp: time.Now()}
	Save(entry1)

	// 2. Load and verify identity
	files, _ := os.ReadDir(filepath.Join(tmpDir, ".cleat"))
	var historyFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".history.yaml") {
			historyFiles = append(historyFiles, f.Name())
		}
	}
	if len(historyFiles) != 1 {
		t.Errorf("Expected 1 history file, got %d", len(historyFiles))
	}

	// The filename should start with the tmpDir base name, not "cmd"
	expectedPrefix := filepath.Base(projectRoot)
	if !strings.HasPrefix(historyFiles[0], expectedPrefix) {
		t.Errorf("Expected filename to start with %s, got %s", expectedPrefix, historyFiles[0])
	}
}

func TestHistoryRootEdgeCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-history-edge-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldUserHomeDir := UserHomeDir
	UserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { UserHomeDir = oldUserHomeDir }()

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// Test with root directory (edge case)
	os.Chdir("/")
	path, err := getHistoryFilePath()
	if err != nil {
		t.Fatalf("Failed to get history file path for root: %v", err)
	}

	// Should use "root" as project name for /
	if !strings.Contains(path, "root-") {
		t.Errorf("Expected 'root-' in path for root directory, got %s", path)
	}
}

func TestHistoryWithInputs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-history-inputs-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldUserHomeDir := UserHomeDir
	UserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { UserHomeDir = oldUserHomeDir }()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Test saving and loading with various input combinations
	testCases := []struct {
		name   string
		inputs map[string]string
	}{
		{"nil inputs", nil},
		{"empty inputs", map[string]string{}},
		{"single input", map[string]string{"key": "value"}},
		{"multiple inputs", map[string]string{"env": "prod", "region": "us-west"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entry := HistoryEntry{
				Command:   "test-" + tc.name,
				Timestamp: time.Now(),
				Inputs:    tc.inputs,
				Success:   true,
			}

			err := Save(entry)
			if err != nil {
				t.Fatalf("Failed to save: %v", err)
			}

			entries, err := Load()
			if err != nil {
				t.Fatalf("Failed to load: %v", err)
			}

			// Find our entry (it should be first due to prepending)
			if len(entries) == 0 {
				t.Fatal("No entries loaded")
			}

			loaded := entries[0]
			if loaded.Command != entry.Command {
				t.Errorf("Command mismatch: expected %s, got %s", entry.Command, loaded.Command)
			}

			if len(loaded.Inputs) != len(tc.inputs) {
				t.Errorf("Inputs length mismatch: expected %d, got %d", len(tc.inputs), len(loaded.Inputs))
			}
		})
	}
}

func TestStats(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-stats-test-*")
	defer os.RemoveAll(tmpDir)

	oldUserHomeDir := UserHomeDir
	UserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { UserHomeDir = oldUserHomeDir }()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Test UpdateStats
	UpdateStats("build")
	UpdateStats("build")
	UpdateStats("run")

	top, err := GetTopCommands(10)
	if err != nil {
		t.Fatalf("GetTopCommands failed: %v", err)
	}

	if len(top) != 2 {
		t.Errorf("expected 2 top commands, got %d", len(top))
	}
	if top[0].Command != "build" || top[0].Count != 2 {
		t.Errorf("expected build count 2, got %v", top[0])
	}
}

func TestWorkflows(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-wf-test-*")
	defer os.RemoveAll(tmpDir)

	oldUserHomeDir := UserHomeDir
	UserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { UserHomeDir = oldUserHomeDir }()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create a dummy cleat.yaml
	os.WriteFile("cleat.yaml", []byte("version: 1"), 0644)

	wf := config.Workflow{
		Name:     "deploy",
		Commands: []string{"build", "gcp app-engine deploy"},
	}

	t.Run("SaveToProject", func(t *testing.T) {
		err := SaveWorkflowToProject(wf)
		if err != nil {
			t.Fatalf("SaveWorkflowToProject failed: %v", err)
		}

		loaded, _ := LoadWorkflows(nil)
		found := false
		for _, w := range loaded {
			if w.Name == "deploy" {
				found = true
				break
			}
		}
		if !found {
			t.Error("workflow not found in project")
		}
	})

	t.Run("SaveToUser", func(t *testing.T) {
		wfUser := wf
		wfUser.Name = "user-wf"
		err := SaveWorkflowToUser(wfUser)
		if err != nil {
			t.Fatalf("SaveWorkflowToUser failed: %v", err)
		}

		loaded, _ := LoadWorkflows(nil)
		found := false
		for _, w := range loaded {
			if w.Name == "user-wf" {
				found = true
				break
			}
		}
		if !found {
			t.Error("workflow not found in user config")
		}
	})

	t.Run("DeleteWorkflow", func(t *testing.T) {
		err := DeleteWorkflow("deploy")
		if err != nil {
			t.Fatalf("DeleteWorkflow failed: %v", err)
		}

		loaded, _ := LoadWorkflows(nil)
		for _, w := range loaded {
			if w.Name == "deploy" {
				t.Error("workflow still exists after delete")
			}
		}
	})
}
