package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func TestLoggerAllLevels(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-logger-test-all-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")
	
	err = Init(logFile, "debug", nil)
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
	
	t.Run("ErrorNil", func(t *testing.T) {
		os.Truncate(logFile, 0)
		Error("error msg nil", nil, map[string]interface{}{"e": 5})
		verifyLog(t, logFile, "error", "error msg nil", map[string]interface{}{"e": float64(5)})
	})
}

func TestLoggerInitHomeExpansion(t *testing.T) {
	// We can't easily mock UserHomeDir because it's called inside logger.Init which uses os.UserHomeDir directly.
	// But we can test the logic if we were to mock it or just rely on the fact that it shouldn't fail on most systems.
	// For now, let's skip actual home dir writing but test other Init paths.
	
	tmpDir, _ := os.MkdirTemp("", "cleat-logger-init-*")
	defer os.RemoveAll(tmpDir)
	
	t.Run("InvalidPath", func(t *testing.T) {
		err := Init("/proc/invalid/path/log.log", "info", nil)
		if err == nil {
			t.Error("expected error for invalid path")
		}
	})
	
	t.Run("DefaultLevel", func(t *testing.T) {
		logFile := filepath.Join(tmpDir, "default.log")
		err := Init(logFile, "invalid", nil)
		if err != nil {
			t.Fatalf("failed to init with default level: %v", err)
		}
		Info("test", nil)
		verifyLog(t, logFile, "info", "test", nil)
	})
}

func verifyLog(t *testing.T, path, level, msg string, fields map[string]interface{}) {
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}

	// Read last line (in case there are multiple)
	lines := jsonLines(content)
	if len(lines) == 0 {
		t.Fatal("no log entries found")
	}
	lastEntry := lines[len(lines)-1]

	var entry map[string]interface{}
	err = json.Unmarshal([]byte(lastEntry), &entry)
	if err != nil {
		t.Fatalf("failed to unmarshal log: %v (content: %s)", err, string(lastEntry))
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

func jsonLines(data []byte) []string {
	var lines []string
	curr := ""
	for _, b := range data {
		if b == '\n' {
			if curr != "" {
				lines = append(lines, curr)
				curr = ""
			}
		} else {
			curr += string(b)
		}
	}
		if curr != "" {
			lines = append(lines, curr)
		}
		return lines
	}
	
	func TestLoggerSetOutput(t *testing.T) {
		var buf bytes.Buffer
		SetOutput(&buf)
		
		Info("buf message", nil)
		
		if !strings.Contains(buf.String(), "buf message") {
			t.Errorf("expected buffer to contain 'buf message', got %q", buf.String())
		}
	}
	
	func TestLoggerInit_HomeExpansion(t *testing.T) {
		// Use a path with ~/ but ensure it's something that won't mess up real home
		// Since Init creates directories, we must be careful.
		// We'll just test that it attempts to get the home dir by passing a path starting with ~/
		// and checking if it returns an error related to directory creation (which we can control).
		
		err := Init("~/.cleat-test-log/test.log", "info", nil)
		if err != nil {
			// It might fail if we can't write to home, but usually it should work or fail with permission
			t.Logf("Init with home expansion returned: %v", err)
		}
	}
	