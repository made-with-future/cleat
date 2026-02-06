package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madewithfuture/cleat/internal/executor"
)

type mockTFExecutor struct {
	executor.ShellExecutor
	runCalled bool
}

func (m *mockTFExecutor) Run(name string, args ...string) error {
	m.runCalled = true
	return nil
}

func (m *mockTFExecutor) RunWithDir(dir string, name string, args ...string) error {
	m.runCalled = true
	return nil
}

func TestTerraformCmd(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-tf-cmd-*")
	defer os.RemoveAll(tmpDir)
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	oldExec := executor.Default
	mock := &mockTFExecutor{}
	executor.Default = mock
	defer func() { executor.Default = oldExec }()
	
	oldConfigPath := ConfigPath
	defer func() { ConfigPath = oldConfigPath }()

	t.Run("NoTerraformConfig", func(t *testing.T) {
		ConfigPath = filepath.Join(tmpDir, "no-tf.yaml")
		os.WriteFile(ConfigPath, []byte("version: 1"), 0644)
		rootCmd.SetArgs([]string{"terraform", "plan"})
		err := rootCmd.Execute()
		if err == nil || err.Error() != "terraform is not configured in cleat.yaml" {
			t.Errorf("expected terraform config error, got %v", err)
		}
	})

	t.Run("MultiEnvRequired", func(t *testing.T) {
		ConfigPath = filepath.Join(tmpDir, "multi-tf.yaml")
		os.WriteFile(ConfigPath, []byte("version: 1\nterraform: {envs: [dev, prod], dir: .}"), 0644)
		
		// Create a directory structure that triggers UseFolders
		iacDir := filepath.Join(tmpDir, ".iac")
		os.Mkdir(iacDir, 0755)
		os.Mkdir(filepath.Join(iacDir, "dev"), 0755)
		os.Mkdir(filepath.Join(iacDir, "prod"), 0755)
		os.WriteFile(filepath.Join(iacDir, "dev", "main.tf"), []byte(""), 0644)
		os.WriteFile(filepath.Join(iacDir, "prod", "main.tf"), []byte(""), 0644)
		
		// Since we're using ConfigPath, we need to make sure detector runs on the right baseDir
		// LoadConfig will run detector on filepath.Dir(ConfigPath) which is tmpDir.
		
		rootCmd.SetArgs([]string{"terraform", "plan"})
		err := rootCmd.Execute()
		if err == nil {
			t.Error("expected error for missing env when multiple exist")
		}
	})

	t.Run("ValidEnv", func(t *testing.T) {
		ConfigPath = filepath.Join(tmpDir, "valid-tf.yaml")
		os.WriteFile(ConfigPath, []byte("version: 1\nterraform: {envs: [prod], dir: .}"), 0644)
		rootCmd.SetArgs([]string{"terraform", "plan", "prod"})
		mock.runCalled = false
		err := rootCmd.Execute()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !mock.runCalled {
			t.Error("expected executor.Run to be called")
		}
	})
}
