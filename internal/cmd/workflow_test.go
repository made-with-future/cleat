package cmd

import (
	"os"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"gopkg.in/yaml.v3"
)

type MockExecutor struct {
	executor.ShellExecutor
	RunCalled bool
	LastCmd   string
}

func (e *MockExecutor) Run(name string, args ...string) error {
	e.RunCalled = true
	e.LastCmd = name
	return nil
}

func (e *MockExecutor) RunWithDir(dir string, name string, args ...string) error {
	e.RunCalled = true
	e.LastCmd = name
	return nil
}

func TestRunWorkflowCmd_ExternalFile(t *testing.T) {
	// Setup a temporary directory
	tmpDir, err := os.MkdirTemp("", "cleat-workflow-cmd-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// 1. Create a basic cleat.yaml (without the workflow)
	os.WriteFile("cleat.yaml", []byte("version: 1\n"), 0644)

	// 2. Create cleat.workflows.yaml with the workflow we want to run
	externalWf := []config.Workflow{
		{
			Name:     "external-wf",
			Commands: []string{"echo test"},
		},
	}
	data, _ := yaml.Marshal(externalWf)
	os.WriteFile("cleat.workflows.yaml", data, 0644)

	// 3. Mock the executor
	mockExec := &MockExecutor{}
	// We need to inject this mock executor into the session created by createSessionAndMerge.
	// However, createSessionAndMerge uses executor.Default.
	// We can temporarily swap executor.Default for this test.
	oldDefault := executor.Default
	executor.Default = mockExec
	defer func() { executor.Default = oldDefault }()

	// 4. Run the command
	rootCmd.SetArgs([]string{"workflow", "external-wf"})
	err = rootCmd.Execute()

	// 5. Verify
	if err != nil {
		t.Fatalf("run-workflow failed: %v", err)
	}

	if !mockExec.RunCalled {
		t.Error("External workflow command was not executed")
	}
}
