package history

import (
	"os"
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
