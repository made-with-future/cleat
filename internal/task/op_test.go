package task

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShouldUseOp(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cleat-op-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	envsDir := filepath.Join(tempDir, ".envs")
	if err := os.Mkdir(envsDir, 0755); err != nil {
		t.Fatalf("failed to create .envs dir: %v", err)
	}

	// Test case 1: No op CLI (hard to test without mocking exec.LookPath)
	// We'll assume op is NOT in path for this test environment unless it's there.

	// Test case 2: op:// present in .env file
	err = os.WriteFile(filepath.Join(envsDir, "dev.env"), []byte("DB_PASSWORD=op://vault/item/password\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	// Note: We can't easily test ShouldUseOp returning true without 'op' in PATH.
	// But we can test it returns false when op:// is NOT present.

	t.Run("op:// present", func(t *testing.T) {
		err = os.WriteFile(filepath.Join(envsDir, "dev.env"), []byte("DB_PASSWORD=op://vault/item/password\n"), 0644)
		if err != nil {
			t.Fatalf("failed to write .env file: %v", err)
		}
		// Since 'op' is in PATH in this environment, this should return true
		if !ShouldUseOp(tempDir) {
			t.Errorf("ShouldUseOp() should return true when op:// is present and op CLI exists")
		}
	})

	t.Run("op:// not present", func(t *testing.T) {
		err = os.WriteFile(filepath.Join(envsDir, "dev.env"), []byte("DB_PASSWORD=secret\n"), 0644)
		if err != nil {
			t.Fatalf("failed to write .env file: %v", err)
		}
		if ShouldUseOp(tempDir) {
			t.Errorf("ShouldUseOp() should return false when op:// is not present")
		}
	})

	t.Run("op:// present but no .envs dir", func(t *testing.T) {
		if ShouldUseOp("/nonexistent") {
			t.Errorf("ShouldUseOp() should return false for nonexistent dir")
		}
	})
}
