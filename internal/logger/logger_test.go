package logger

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoggerAllLevels(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-logger-test-all-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")
	
	err = Init(logFile, "debug")
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}

	t.Run("Debug", func(t *testing.T) {
		os.Truncate(logFile, 0)
		Debug("debug msg", map[string]interface{}{"d": 1})
		verifyLog(t, logFile, "debug", "debug msg", map[string]interface{}{"d": float64(1)})
	})

	t.Run("Info", func(t *testing.T) {
		os.Truncate(logFile, 0)
		Info("info msg", map[string]interface{}{"i": 2})
		verifyLog(t, logFile, "info", "info msg", map[string]interface{}{"i": float64(2)})
	})

	t.Run("Warn", func(t *testing.T) {
		os.Truncate(logFile, 0)
		Warn("warn msg", map[string]interface{}{"w": 3})
		verifyLog(t, logFile, "warn", "warn msg", map[string]interface{}{"w": float64(3)})
	})

	t.Run("Error", func(t *testing.T) {
		os.Truncate(logFile, 0)
		err := errors.New("boom")
		Error("error msg", err, map[string]interface{}{"e": 4})
		verifyLog(t, logFile, "error", "error msg", map[string]interface{}{"e": float64(4), "error": "boom"})
	})
}

func verifyLog(t *testing.T, path, level, msg string, fields map[string]interface{}) {
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}

	var entry map[string]interface{}
	err = json.Unmarshal(content, &entry)
	if err != nil {
		t.Fatalf("failed to unmarshal log: %v", err)
	}

	if entry["level"] != level {
		t.Errorf("expected level %q, got %q", level, entry["level"])
	}
	if entry["message"] != msg {
		t.Errorf("expected message %q, got %q", msg, entry["message"])
	}
	for k, v := range fields {
		if entry[k] != v {
			t.Errorf("expected field %q to be %v, got %v", k, v, entry[k])
		}
	}
}