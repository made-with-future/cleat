package executor

import (
	"os"
	"testing"
)

func TestShellExecutor_Run(t *testing.T) {
	e := &ShellExecutor{}
	
	// Test basic command (using 'true' as a reliable cross-platform-ish NOOP command on Unix)
	// On Windows this might fail, but we're on Linux as per prompt.
	err := e.Run("true")
	if err != nil {
		t.Errorf("expected no error running 'true', got %v", err)
	}

	// Test failing command
	err = e.Run("false")
	if err == nil {
		t.Error("expected error running 'false', got nil")
	}
}

func TestShellExecutor_RunWithDir(t *testing.T) {
	e := &ShellExecutor{}
	tmpDir, _ := os.MkdirTemp("", "cleat-exec-test-*")
	defer os.RemoveAll(tmpDir)

	// Test command in specific dir
	// We'll use 'ls' and check output if we could capture it, but for now we just check error.
	err := e.RunWithDir(tmpDir, "ls")
	if err != nil {
		t.Errorf("expected no error running 'ls' in tmpDir, got %v", err)
	}
}

func TestShellExecutor_Prompt(t *testing.T) {
	e := &ShellExecutor{}

	t.Run("DefaultValue", func(t *testing.T) {
		input := "\n"
		oldStdin := os.Stdin
		r, w, _ := os.Pipe()
		os.Stdin = r
		w.Write([]byte(input))
		w.Close()
		defer func() { os.Stdin = oldStdin }()

		val, err := e.Prompt("message", "default")
		if err != nil {
			t.Fatalf("Prompt failed: %v", err)
		}
		if val != "default" {
			t.Errorf("expected 'default', got %q", val)
		}
	})

	t.Run("UserInput", func(t *testing.T) {
		input := "user input\n"
		oldStdin := os.Stdin
		r, w, _ := os.Pipe()
		os.Stdin = r
		w.Write([]byte(input))
		w.Close()
		defer func() { os.Stdin = oldStdin }()

		val, err := e.Prompt("message", "default")
		if err != nil {
			t.Fatalf("Prompt failed: %v", err)
		}
		if val != "user input" {
			t.Errorf("expected 'user input', got %q", val)
		}
	})
	
	t.Run("InputError", func(t *testing.T) {
		// Mock a closed stdin or similar to trigger error
		oldStdin := os.Stdin
		r, w, _ := os.Pipe()
		os.Stdin = r
		r.Close() // Close reader to trigger error on ReadString
		w.Close()
		defer func() { os.Stdin = oldStdin }()

		_, err := e.Prompt("message", "default")
		if err == nil {
			t.Error("expected error for closed stdin, got nil")
		}
	})
}

func TestDefaultExecutor(t *testing.T) {
	if Default == nil {
		t.Error("Default executor should not be nil")
	}
}
