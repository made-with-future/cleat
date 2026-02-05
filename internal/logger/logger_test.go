package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLogger(t *testing.T) {
	// Create a temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "cleat-logger-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")
	
	// Initialize logger with the test file and a project context
	err = Init(logFile, "debug", map[string]interface{}{"project": "test-project"})
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}

	// Log a message
	Info("test message", map[string]interface{}{"key": "value"})

	// Verify file content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	var entry map[string]interface{}
	err = json.Unmarshal(content, &entry)
	if err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry["message"] != "test message" {
		t.Errorf("expected message 'test message', got %q", entry["message"])
	}
	if entry["key"] != "value" {
		t.Errorf("expected key 'value', got %v", entry["key"])
	}
	if entry["level"] != "info" {
		t.Errorf("expected level 'info', got %v", entry["level"])
	}
	if entry["project"] != "test-project" {
		t.Errorf("expected project 'test-project', got %v", entry["project"])
	}
}
