package history

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"gopkg.in/yaml.v3"
)

func TestLoadWorkflows_Validation(t *testing.T) {
	// Setup tmp dir
	tmpDir, err := os.MkdirTemp("", "cleat-history-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create a workflow file with invalid workflow (empty commands)
	invalidWf := []config.Workflow{
		{
			Name:     "invalid-wf",
			Commands: []string{}, // Empty commands
		},
	}
	data, _ := yaml.Marshal(invalidWf)
	os.WriteFile("cleat.workflows.yaml", data, 0644)

	cfg := &config.Config{}
	workflows, err := LoadWorkflows(cfg)
	if err != nil {
		// It currently doesn't return error, so this might pass
	}

	// We want to assert that invalid workflows are filtered out or error is returned.
	// For this "Red" phase, we'll assert that we WANT an error or the workflow to be rejected.
	
	// Let's assert that we expect LoadWorkflows to validate and return error or filter it.
	// Since the current implementation doesn't validate, we expect this test to FAIL if we check for validation.
	
	found := false
	for _, w := range workflows {
		if w.Name == "invalid-wf" {
			found = true
			break
		}
	}

	if found {
		t.Fatal("Expected invalid workflow 'invalid-wf' to be rejected due to empty commands, but it was loaded")
	}
}

func TestLoadWorkflows_Malformed(t *testing.T) {
	// Setup tmp dir
	tmpDir, err := os.MkdirTemp("", "cleat-history-malformed-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create a malformed workflow file
	os.WriteFile("cleat.workflows.yaml", []byte("invalid: yaml: content: ["), 0644)

	cfg := &config.Config{}
	_, err = LoadWorkflows(cfg)
	if err == nil {
		t.Fatal("Expected error for malformed workflow file, got nil")
	}
}

func TestSaveWorkflowToProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-save-project-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	wf := config.Workflow{
		Name:     "test-wf",
		Commands: []string{"echo hello"},
	}

	// 1. Save new workflow
	err = SaveWorkflowToProject(wf)
	if err != nil {
		t.Fatalf("SaveWorkflowToProject failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat("cleat.workflows.yaml"); os.IsNotExist(err) {
		t.Fatal("cleat.workflows.yaml was not created")
	}

	// 2. Load and verify
	workflows, err := LoadWorkflows(nil)
	if err != nil {
		t.Fatalf("LoadWorkflows failed: %v", err)
	}
	if len(workflows) != 1 || workflows[0].Name != "test-wf" {
		t.Fatalf("Expected 1 workflow 'test-wf', got %v", workflows)
	}

	// 3. Update existing
	wf.Commands = []string{"echo updated"}
	err = SaveWorkflowToProject(wf)
	if err != nil {
		t.Fatalf("SaveWorkflowToProject update failed: %v", err)
	}

	workflows, _ = LoadWorkflows(nil)
	if len(workflows) != 1 || workflows[0].Commands[0] != "echo updated" {
		t.Fatalf("Expected updated workflow, got %v", workflows)
	}

	// 4. Fallback to .yml
	os.Remove("cleat.workflows.yaml")
	os.WriteFile("cleat.workflows.yml", []byte("- name: yml-wf\n  commands: [ls]"), 0644)
	
	err = SaveWorkflowToProject(wf)
	if err != nil {
		t.Fatalf("SaveWorkflowToProject with .yml failed: %v", err)
	}

	if _, err := os.Stat("cleat.workflows.yaml"); err == nil {
		t.Fatal("cleat.workflows.yaml should not have been created when .yml exists")
	}

	data, _ := os.ReadFile("cleat.workflows.yml")
	var ymlWorkflows []config.Workflow
	yaml.Unmarshal(data, &ymlWorkflows)
	
	found := false
	for _, w := range ymlWorkflows {
		if w.Name == "test-wf" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Workflow was not saved to cleat.workflows.yml")
	}
}

func TestDeleteWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-delete-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	wf := config.Workflow{Name: "to-delete", Commands: []string{"ls"}}
	SaveWorkflowToProject(wf)

	err = DeleteWorkflow("to-delete")
	if err != nil {
		t.Fatalf("DeleteWorkflow failed: %v", err)
	}

	workflows, _ := LoadWorkflows(nil)
	if len(workflows) != 0 {
		t.Fatal("Workflow was not deleted")
	}
}

func TestSaveWorkflowToUser(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-save-user-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock UserHomeDir
	oldUserHomeDir := UserHomeDir
	UserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { UserHomeDir = oldUserHomeDir }()

	wf := config.Workflow{
		Name:     "user-wf",
		Commands: []string{"echo user"},
	}

	err = SaveWorkflowToUser(wf)
	if err != nil {
		t.Fatalf("SaveWorkflowToUser failed: %v", err)
	}

	// Load and verify
	workflows, err := LoadWorkflows(nil)
	if err != nil {
		t.Fatalf("LoadWorkflows failed: %v", err)
	}

	found := false
	for _, w := range workflows {
		if w.Name == "user-wf" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("User workflow was not loaded")
	}

	// 3. Update existing user workflow
	wf.Commands = []string{"echo user-updated"}
	err = SaveWorkflowToUser(wf)
	if err != nil {
		t.Fatalf("SaveWorkflowToUser update failed: %v", err)
	}

	workflows, _ = LoadWorkflows(nil)
	found = false
	for _, w := range workflows {
		if w.Name == "user-wf" && w.Commands[0] == "echo user-updated" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("User workflow was not updated")
	}
}

func TestSaveWorkflow_IDBasedUpsert(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-upsert-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// 1. Save initial workflow
	wf1 := config.Workflow{
		ID:       "my-workflow",
		Name:     "My Workflow",
		Commands: []string{"echo 1"},
	}
	err = SaveWorkflowToProject(wf1)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Save workflow with SAME ID but DIFFERENT Name
	wf2 := config.Workflow{
		ID:       "my-workflow",
		Name:     "Updated Name",
		Commands: []string{"echo 2"},
	}
	err = SaveWorkflowToProject(wf2)
	if err != nil {
		t.Fatal(err)
	}

	// 3. Verify we have only 1 workflow and it has the updated name
	workflows, _ := LoadWorkflows(nil)
	if len(workflows) != 1 {
		t.Fatalf("Expected 1 workflow, got %d", len(workflows))
	}
	if workflows[0].Name != "Updated Name" {
		t.Fatalf("Expected name 'Updated Name', got %q", workflows[0].Name)
	}
}

func TestSaveWorkflowToUser_Errors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-save-user-errors-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 1. GetUserWorkflowFilePath error
	oldUserHomeDir := UserHomeDir
	UserHomeDir = func() (string, error) {
		return "", fmt.Errorf("home error")
	}
	err = SaveWorkflowToUser(config.Workflow{Name: "err"})
	UserHomeDir = oldUserHomeDir
	if err == nil {
		t.Fatalf("Expected home error, got %v", err)
	}

	// 2. MkdirAll error
	// Create a file where the directory should be
	UserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { UserHomeDir = oldUserHomeDir }()

	cleatDir := filepath.Join(tmpDir, ".cleat")
	os.WriteFile(cleatDir, []byte("im a file"), 0644)

	err = SaveWorkflowToUser(config.Workflow{Name: "err"})
	if err == nil {
		t.Fatal("Expected error from MkdirAll, got nil")
	}
}

func TestModifyWorkflowFile_Errors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-modify-errors-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "bad.yaml")

	// 1. Malformed YAML
	os.WriteFile(path, []byte("invalid: ["), 0644)
	err = modifyWorkflowFile(path, func(w []config.Workflow) ([]config.Workflow, bool, error) {
		return nil, false, nil
	})
	if err == nil {
		t.Fatal("Expected error for malformed YAML, got nil")
	}

	// 2. Op returns error
	os.WriteFile(path, []byte("[]"), 0644)
	expectedErr := fmt.Errorf("op error")
	err = modifyWorkflowFile(path, func(w []config.Workflow) ([]config.Workflow, bool, error) {
		return nil, false, expectedErr
	})
	if err != expectedErr {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}

	// 3. Modified false (no write)
	err = modifyWorkflowFile(path, func(w []config.Workflow) ([]config.Workflow, bool, error) {
		return nil, false, nil
	})
	if err != nil {
		t.Fatalf("modifyWorkflowFile failed: %v", err)
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "Deploy Project", "deploy-project"},
		{"lowercase", "build", "build"},
		{"special chars", "My Workflow! @2026", "my-workflow-2026"},
		{"multiple spaces", "test   multiple   spaces", "test-multiple-spaces"},
		{"leading/trailing spaces", "  trim me  ", "trim-me"},
		{"numeric", "123", "123"},
		{"mixed alphanumeric", "v1.0.0-beta", "v100-beta"}, // Removing dots as per "remove non-alphanumeric"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Slugify(tt.input)
			if got != tt.expected {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestValidateWorkflowName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "Valid Name", false},
		{"empty", "", true},
		{"spaces only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkflowName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWorkflowName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
