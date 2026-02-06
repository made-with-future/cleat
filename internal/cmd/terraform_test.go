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
		tmpDir, _ := os.MkdirTemp("", "cleat-tf-cmd-no-tf-*")
		defer os.RemoveAll(tmpDir)
		ConfigPath = filepath.Join(tmpDir, "no-tf.yaml")
		os.WriteFile(ConfigPath, []byte("version: 1"), 0644)
		rootCmd.SetArgs([]string{"terraform", "plan"})
		err := rootCmd.Execute()
		if err == nil || err.Error() != "terraform is not configured in cleat.yaml" {
			t.Errorf("expected terraform config error, got %v", err)
		}
	})

	t.Run("MultiEnvRequired", func(t *testing.T) {
		tmpDir, _ := os.MkdirTemp("", "cleat-tf-cmd-multi-*")
		defer os.RemoveAll(tmpDir)
		ConfigPath = filepath.Join(tmpDir, "multi-tf.yaml")
		os.WriteFile(ConfigPath, []byte("version: 1\nterraform: {envs: [dev, prod], dir: .}"), 0644)

		// Create a directory structure that triggers UseFolders
		iacDir := filepath.Join(tmpDir, ".iac")
		os.Mkdir(iacDir, 0755)
		os.Mkdir(filepath.Join(iacDir, "dev"), 0755)
		os.Mkdir(filepath.Join(iacDir, "prod"), 0755)
		os.WriteFile(filepath.Join(iacDir, "dev", "main.tf"), []byte(""), 0644)
		os.WriteFile(filepath.Join(iacDir, "prod", "main.tf"), []byte(""), 0644)

		rootCmd.SetArgs([]string{"terraform", "plan"})
		err := rootCmd.Execute()
		if err == nil {
			t.Error("expected error for missing env when multiple exist")
		}
	})

	t.Run("ValidEnv", func(t *testing.T) {
		tmpDir, _ := os.MkdirTemp("", "cleat-tf-cmd-valid-*")
		defer os.RemoveAll(tmpDir)
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

	t.Run("EnvFromGeneralEnvs", func(t *testing.T) {
		tmpDir, _ := os.MkdirTemp("", "cleat-tf-cmd-general-*")
		defer os.RemoveAll(tmpDir)
		ConfigPath = filepath.Join(tmpDir, "general-envs-tf.yaml")
		// No envs in terraform config
		os.WriteFile(ConfigPath, []byte("version: 1\nterraform: {dir: .}"), 0644)

		// Create .envs/dev.env to populate cfg.Envs
		envsDir := filepath.Join(tmpDir, ".envs")
		os.MkdirAll(envsDir, 0755)
		os.WriteFile(filepath.Join(envsDir, "dev.env"), []byte("OP_SECRET=op://vault/item/secret"), 0644)

		rootCmd.SetArgs([]string{"terraform", "plan", "dev"})
		mock.runCalled = false
		err := rootCmd.Execute()

		if err != nil {
			t.Errorf("expected dev env to be valid from general envs, got error: %v", err)
		}
		if !mock.runCalled {
			t.Error("expected executor.Run to be called")
		}
	})
}
