package history

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"gopkg.in/yaml.v3"
)

func TestLoadWorkflowsMultiLocation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-workflow-test-*")
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

	// Mock project root
	projectRoot := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectRoot, 0755)

	oldWd, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

	// 1. Prepare cleat.yaml with a workflow
	cfg := &config.Config{
		Workflows: []config.Workflow{
			{
				Name: "cleat-yml-wf",
				Commands: []string{
					"cmd1",
				},
			},
			{
				Name: "overridden-wf",
				Commands: []string{
					"original",
				},
			},
		},
	}

	// 2. Prepare cleat.workflows.yaml
	projectWorkflows := []config.Workflow{
		{
			Name: "project-wf",
			Commands: []string{
				"cmd2",
			},
		},
	}
	pwData, _ := yaml.Marshal(projectWorkflows)
	os.WriteFile(filepath.Join(projectRoot, "cleat.workflows.yaml"), pwData, 0644)

	// 3. Prepare user workflows (YAML in home dir)
	userWorkflows := []config.Workflow{
		{
			Name: "user-wf",
			Commands: []string{
				"cmd3",
			},
		},
		{
			Name: "overridden-wf",
			Commands: []string{
				"overridden-by-user",
			},
		},
	}
	uData, _ := yaml.Marshal(userWorkflows)
	userPath, _ := GetUserWorkflowFilePath()
	os.MkdirAll(filepath.Dir(userPath), 0755)
	os.WriteFile(userPath, uData, 0644)

	// Test loading
	workflows, err := LoadWorkflows(cfg)
	if err != nil {
		t.Fatalf("LoadWorkflows failed: %v", err)
	}

	// Check counts (cleat-yml-wf, project-wf, user-wf, overridden-wf)
	if len(workflows) != 4 {
		t.Errorf("Expected 4 workflows, got %d", len(workflows))
	}

	// Verify overrides
	var foundOverridden bool
	for _, w := range workflows {
		if w.Name == "overridden-wf" {
			foundOverridden = true
			if w.Commands[0] != "overridden-by-user" {
				t.Errorf("Expected overridden-wf to have command 'overridden-by-user', got '%s'", w.Commands[0])
			}
		}
	}
	if !foundOverridden {
		t.Error("overridden-wf not found")
	}

	// Verify user-wf was loaded
	foundUser := false
	for _, w := range workflows {
		if w.Name == "user-wf" {
			foundUser = true
			break
		}
	}
	if !foundUser {
		t.Error("user-wf not found")
	}
}

func TestSaveWorkflowToProjectFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-workflow-save-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	projectRoot := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectRoot, 0755)

	oldWd, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

	wf := config.Workflow{
		Name: "new-wf",
		Commands: []string{
			"test-cmd",
		},
	}

	err = SaveWorkflowToProject(wf)
	if err != nil {
		t.Fatalf("SaveWorkflowToProject failed: %v", err)
	}

	// Verify file exists and has correct content
	projectFile := filepath.Join(projectRoot, "cleat.workflows.yaml")
	if _, err := os.Stat(projectFile); os.IsNotExist(err) {
		t.Fatal("cleat.workflows.yaml was not created")
	}

	data, _ := os.ReadFile(projectFile)
	var savedWorkflows []config.Workflow
	yaml.Unmarshal(data, &savedWorkflows)

	if len(savedWorkflows) != 1 {
		t.Fatalf("Expected 1 saved workflow, got %d", len(savedWorkflows))
	}

	if savedWorkflows[0].Name != "new-wf" {
		t.Errorf("Expected name 'new-wf', got '%s'", savedWorkflows[0].Name)
	}
}

func TestSaveWorkflowToUserFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-workflow-save-user-test-*")
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

	wf := config.Workflow{
		Name: "user-wf",
		Commands: []string{
			"test-cmd-user",
		},
	}

	err = SaveWorkflowToUser(wf)
	if err != nil {
		t.Fatalf("SaveWorkflowToUser failed: %v", err)
	}

	// Verify file exists and has correct content
	userFile, _ := GetUserWorkflowFilePath()
	if _, err := os.Stat(userFile); os.IsNotExist(err) {
		t.Fatal("user workflow file was not created")
	}

	data, _ := os.ReadFile(userFile)
	var savedWorkflows []config.Workflow
	yaml.Unmarshal(data, &savedWorkflows)

	if len(savedWorkflows) != 1 {
		t.Fatalf("Expected 1 saved workflow, got %d", len(savedWorkflows))
	}

	if savedWorkflows[0].Name != "user-wf" {
		t.Errorf("Expected name 'user-wf', got '%s'", savedWorkflows[0].Name)
	}
}
