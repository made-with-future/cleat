package history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStats(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-stats-test-*")
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

	// Update stats
	err = UpdateStats("cmd1")
	if err != nil {
		t.Fatalf("Failed to update stats: %v", err)
	}

	stats, err := LoadStats()
	if err != nil {
		t.Fatalf("Failed to load stats: %v", err)
	}

	if stats.Commands["cmd1"].Count != 1 {
		t.Errorf("Expected cmd1 count 1, got %d", stats.Commands["cmd1"].Count)
	}

	// Update again
	err = UpdateStats("cmd1")
	if err != nil {
		t.Fatalf("Failed to update stats: %v", err)
	}

	err = UpdateStats("cmd2")
	if err != nil {
		t.Fatalf("Failed to update stats: %v", err)
	}

	stats, err = LoadStats()
	if err != nil {
		t.Fatalf("Failed to load stats: %v", err)
	}

	if stats.Commands["cmd1"].Count != 2 {
		t.Errorf("Expected cmd1 count 2, got %d", stats.Commands["cmd1"].Count)
	}

	if stats.Commands["cmd2"].Count != 1 {
		t.Errorf("Expected cmd2 count 1, got %d", stats.Commands["cmd2"].Count)
	}

	// Check file existence
	files, _ := os.ReadDir(filepath.Join(tmpDir, ".cleat"))
	found := false
	var foundName string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".stats.yaml") {
			found = true
			foundName = f.Name()
		}
	}
	if !found {
		t.Error("Stats file not created")
	} else {
		t.Logf("Created stats file: %s", foundName)
	}
}

func TestStatsFileSeparation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-stats-sep-test-*")
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

	// Project 1
	p1 := filepath.Join(tmpDir, "p1")
	os.MkdirAll(p1, 0755)
	os.WriteFile(filepath.Join(p1, "cleat.yaml"), []byte(""), 0644)

	// Project 2
	p2 := filepath.Join(tmpDir, "p2")
	os.MkdirAll(p2, 0755)
	os.WriteFile(filepath.Join(p2, "cleat.yaml"), []byte(""), 0644)

	// Update stats for p1
	os.Chdir(p1)
	UpdateStats("cmd-p1")

	// Update stats for p2
	os.Chdir(p2)
	UpdateStats("cmd-p2")

	// Load stats for p1
	os.Chdir(p1)
	s1, _ := LoadStats()
	if _, ok := s1.Commands["cmd-p1"]; !ok {
		t.Error("p1 should have cmd-p1")
	}
	if _, ok := s1.Commands["cmd-p2"]; ok {
		t.Error("p1 should NOT have cmd-p2")
	}

	// Load stats for p2
	os.Chdir(p2)
	s2, _ := LoadStats()
	if _, ok := s2.Commands["cmd-p2"]; !ok {
		t.Error("p2 should have cmd-p2")
	}
	if _, ok := s2.Commands["cmd-p1"]; ok {
		t.Error("p2 should NOT have cmd-p1")
	}
}

func TestGetTopCommands(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-stats-top-test-*")
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

	// Setup stats
	cmds := []string{"a", "b", "b", "c", "c", "c", "d", "d", "d", "d"}
	for _, cmd := range cmds {
		UpdateStats(cmd)
	}

	top, err := GetTopCommands(3)
	if err != nil {
		t.Fatalf("Failed to get top commands: %v", err)
	}

	if len(top) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(top))
	}

	expected := []string{"d", "c", "b"}
	for i, cmd := range top {
		if cmd.Command != expected[i] {
			t.Errorf("Expected #%d to be %s, got %s", i+1, expected[i], cmd.Command)
		}
	}

	if top[0].Count != 4 {
		t.Errorf("Expected 'd' count 4, got %d", top[0].Count)
	}
}
