package ui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
)

func TestDeleteWorkflow(t *testing.T) {
	// Setup temp dir
	tmpDir, err := os.MkdirTemp("", "cleat-test-delete-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp dir
	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	// Create cleat.yaml to make it a valid project root
	if err := os.WriteFile("cleat.yaml", []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create cleat.workflows.yaml with a test workflow
	workflowContent := `
- name: test-workflow
  description: A test workflow
  commands:
    - echo hello
`
	if err := os.WriteFile("cleat.workflows.yaml", []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize model
	// We need to load config first to pass it to InitialModel
	cfg, err := config.LoadConfig("cleat.yaml")
	if err != nil {
		t.Fatal(err)
	}

	m := InitialModel(cfg, true, "0.1.0", &executor.ShellExecutor{})

	// Check if workflow is loaded and visible
	// We expect "workflows" category and inside it "test-workflow"
	// Since InitialModel expands if only one item, but here we likely have "recent" and "workflows".
	// "workflows" category usually has label "workflows".

	// We need to expand "workflows" item if it's collapsed.
	// Or just check if we can navigate to it.

	// Let's print visible items to debug if needed.
	// t.Logf("Visible items: %v", m.visibleItems)

	// Find "workflows" item and expand it if needed
	workflowsIdx := -1
	for i, v := range m.visibleItems {
		if v.item.Label == "workflows" {
			workflowsIdx = i
			break
		}
	}

	if workflowsIdx != -1 {
		// Move cursor to workflows
		m.cursor = workflowsIdx
		// Expand
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	}

	// Now look for "test-workflow"
	found := false
	for i := 0; i < len(m.visibleItems); i++ {
		item := m.visibleItems[i]
		// Workflow commands are prefixed with "workflow:"
		if item.item.Command == "workflow:test-workflow" {
			m.cursor = i
			found = true
			break
		}
	}

	if !found {
		// Maybe it's not prefixed in visibleItems?
		// logic in buildCommandTree assigns Command field.
		// Let's assume it is. If not, the test will fail and I'll debug.

		// try to dump items
		for i, v := range m.visibleItems {
			t.Logf("Item %d: %s cmd=%s", i, v.item.Label, v.item.Command)
		}
		t.Fatal("workflow not found in visible items")
	}

	item := m.visibleItems[m.cursor]
	if item.item.Command != "workflow:test-workflow" {
		t.Fatalf("failed to navigate to workflow, current: %s", item.item.Command)
	}

	// Press 'd'
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})

	if m.state != stateConfirmDeleteWorkflow {
		t.Errorf("expected stateConfirmDeleteWorkflow, got %v", m.state)
	}

	// Press 'y' to confirm
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})

	if m.state != stateBrowsing {
		t.Errorf("expected stateBrowsing, got %v", m.state)
	}

	// Verify workflow is gone from model
	found = false
	for _, item := range m.visibleItems {
		if item.item.Command == "workflow:test-workflow" {
			found = true
		}
	}
	if found {
		t.Error("workflow still present in visible items")
	}

	// Verify workflow is gone from file
	content, err := os.ReadFile("cleat.workflows.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "test-workflow") {
		t.Error("workflow still present in file")
	}
}

func updateModel(m model, msg tea.Msg) (model, tea.Cmd) {
	mod, cmd := m.Update(msg)
	return mod.(model), cmd
}
