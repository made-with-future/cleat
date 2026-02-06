package ui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/history"
	"github.com/madewithfuture/cleat/internal/task"
	"github.com/muesli/termenv"
)

func ptrBool(b bool) *bool {
	return &b
}

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
	testHomeDir, _ := os.MkdirTemp("", "cleat-ui-test-home-*")
	history.UserHomeDir = func() (string, error) {
		return testHomeDir, nil
	}
}

func TestModelUpdate(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	updatedModel, cmd := m.Update(msg)

	resModel := updatedModel.(model)
	if !resModel.quitting {
		t.Error("expected quitting to be true after pressing 'q'")
	}
	if cmd == nil {
		t.Error("expected a non-nil command after pressing 'q'")
	}

	m = InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	wmsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ = m.Update(wmsg)
	resModel = updatedModel.(model)
	if resModel.width != 100 || resModel.height != 40 {
		t.Errorf("expected width 100, height 40, got %d, %d", resModel.width, resModel.height)
	}
}

func TestModelView(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	m.width = 100
	m.height = 40

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}

	if !strings.Contains(view, "Cleat") {
		t.Error("expected view to contain title 'Cleat'")
	}
	if !strings.Contains(view, "q: quit") {
		t.Error("expected view to contain help text 'q: quit'")
	}
	if !strings.Contains(view, "Commands") {
		t.Error("expected view to contain 'Commands' section")
	}

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	m = updatedModel.(model)
	view = m.View()
	if !strings.Contains(view, "Configuration") {
		t.Error("expected view to contain 'Configuration' section after pressing 'c'")
	}
}

func TestErrorBar(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	m.width = 100
	m.height = 40
	m.fatalError = errors.New("test error occurred")

	view := m.View()
	if !strings.Contains(view, "ERROR: test error occurred") {
		t.Error("expected view to contain error message")
	}

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = updatedModel.(model)
	if m.fatalError != nil {
		t.Error("expected fatalError to be cleared after key press")
	}
}

func TestFiltering(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
	}
	m := InitialModel(cfg, true, "0.1.0", &executor.ShellExecutor{})
	m.width = 100
	m.height = 40

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = updatedModel.(model)
	if !m.filtering {
		t.Error("expected filtering to be true")
	}

	for _, r := range "down" {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updatedModel.(model)
	}
	if m.filterText != "down" {
		t.Errorf("expected filterText 'down', got %q", m.filterText)
	}

	foundDown := false
	for _, item := range m.visibleItems {
		if strings.Contains(item.item.Label, "down") || strings.Contains(item.path, "down") {
			foundDown = true
		}
	}
	if !foundDown {
		t.Error("expected to find item matching 'down' in filtered results")
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updatedModel.(model)
	if m.filterText != "dow" {
		t.Errorf("expected filterText 'dow' after backspace, got %q", m.filterText)
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(model)
	if m.filtering {
		t.Error("expected filtering to be false after Esc")
	}
}

func TestTabbing(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})

	if m.focus != focusCommands {
		t.Error("expected initial focus to be commands")
	}

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focus != focusHistory {
		t.Errorf("expected focus to be history after Tab, got %v", m.focus)
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focus != focusCommands {
		t.Errorf("expected focus to be commands after 2nd Tab, got %v", m.focus)
	}
	
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updatedModel.(model)
	if m.focus != focusHistory {
		t.Errorf("expected focus history after ShiftTab, got %v", m.focus)
	}
}

func TestHelpOverlay(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	m.width = 100
	m.height = 40

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = updatedModel.(model)
	if !m.showHelp {
		t.Error("expected showHelp to be true")
	}

	view := m.View()
	if !strings.Contains(view, "Keyboard Shortcuts") {
		t.Error("expected view to contain 'Keyboard Shortcuts'")
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	m = updatedModel.(model)
	if m.showHelp {
		t.Error("expected showHelp to be false after pressing a key")
	}
}

func TestTUIKeys(t *testing.T) {
	cfg := &config.Config{Docker: true}
	m := InitialModel(cfg, true, "0.1.0", &executor.ShellExecutor{})
	m.width = 100
	m.height = 40

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("C")})
	
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	m = updatedModel.(model)
	if m.focus != focusTasks {
		t.Error("expected focus tasks after 't'")
	}
}

func TestConfirmClearHistory(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	m.state = stateConfirmClearHistory
	
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = updatedModel.(model)
	if m.state != stateBrowsing {
		t.Error("expected state browsing after cancel")
	}
	
	m.state = stateConfirmClearHistory
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m = updatedModel.(model)
	if m.state != stateBrowsing {
		t.Error("expected state browsing after confirm")
	}
}

func TestWorkflowCreationFlow(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	m.history = []history.HistoryEntry{{Command: "build", Timestamp: time.Now()}}
	m.state = stateCreatingWorkflow
	
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if len(m.selectedWorkflowIndices) != 1 {
		t.Error("expected 1 selected index")
	}
	
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	m = updatedModel.(model)
	if m.state != stateWorkflowNameInput {
		t.Errorf("expected state stateWorkflowNameInput, got %v", m.state)
	}
	
	m.textInput.SetValue("my-wf")
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if m.state != stateWorkflowLocationSelection {
		t.Errorf("expected state WorkflowLocationSelection, got %v", m.state)
	}
}

func TestInputCollection(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	m.state = stateInputCollection
	m.requirements = []task.InputRequirement{{Key: "foo", Prompt: "Foo"}}
	m.requirementIdx = 0
	
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if !m.quitting {
		t.Error("expected quitting after filling requirements")
	}
}

func TestShowingConfig(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	m.state = stateShowingConfig
	
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(model)
	if m.state != stateBrowsing {
		t.Error("expected state browsing after Esc")
	}
}

func TestJumpWithNumbers(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	m.history = []history.HistoryEntry{{Command: "h1"}, {Command: "h2"}, {Command: "h3"}}
	
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	m = updatedModel.(model)
	if m.focus != focusHistory || m.historyCursor != 1 {
		t.Errorf("jump to 2 failed: focus=%v, cursor=%d", m.focus, m.historyCursor)
	}
}

func TestWorkflowCreationFlow_Advanced(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	m.history = []history.HistoryEntry{
		{Command: "ls", Timestamp: time.Now()},
		{Command: "pwd", Timestamp: time.Now()},
	}

	// 1. Selection in stateCreatingWorkflow
	m.state = stateCreatingWorkflow
	m.historyCursor = 0
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if len(m.selectedWorkflowIndices) != 1 || m.selectedWorkflowIndices[0] != 0 {
		t.Error("Expected index 0 to be selected")
	}

	m.historyCursor = 1
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = updatedModel.(model)
	if len(m.selectedWorkflowIndices) != 2 {
		t.Error("Expected 2 indices to be selected")
	}

	// Toggle off
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if len(m.selectedWorkflowIndices) != 1 || m.selectedWorkflowIndices[0] != 0 {
		t.Error("Expected index 1 to be deselected")
	}

	// 2. Transition to name input
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	m = updatedModel.(model)
	if m.state != stateWorkflowNameInput {
		t.Error("Expected transition to stateWorkflowNameInput")
	}

	// 3. Invalid name input
	m.textInput.SetValue("   ")
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if m.state != stateWorkflowNameInput || m.fatalError == nil {
		t.Error("Expected error for empty name")
	}

	// Dismiss error
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = updatedModel.(model)
	if m.fatalError != nil {
		t.Error("Expected error to be dismissed")
	}

	// 4. Valid name input
	m.textInput.SetValue("my-wf")
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if m.state != stateWorkflowLocationSelection {
		t.Error("Expected transition to stateWorkflowLocationSelection")
	}

	// 5. Location selection
	// Default is Project (0)
	// Go down to User (1)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updatedModel.(model)
	if m.workflowLocationIdx != 1 {
		t.Error("Expected User location selected")
	}

	// Go up to Project (0)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updatedModel.(model)
	if m.workflowLocationIdx != 0 {
		t.Error("Expected Project location selected")
	}

	// 6. Confirm Save
	tmpDir, _ := os.MkdirTemp("", "cleat-ui-wf-save-*")
	defer os.RemoveAll(tmpDir)
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	os.WriteFile("cleat.yaml", []byte("version: 1"), 0644)

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if m.state != stateBrowsing {
		t.Error("Expected stateBrowsing after save")
	}
	
	if _, err := os.Stat("cleat.workflows.yaml"); os.IsNotExist(err) {
		t.Error("cleat.workflows.yaml was not created")
	}

	// 7. Save to User
	m.state = stateWorkflowLocationSelection
	m.workflowLocationIdx = 1 // User
	
	// Mock UserHomeDir
	userHome, _ := os.MkdirTemp("", "cleat-ui-user-home-*")
	defer os.RemoveAll(userHome)
	oldUserHomeDir := history.UserHomeDir
	history.UserHomeDir = func() (string, error) { return userHome, nil }
	defer func() { history.UserHomeDir = oldUserHomeDir }()

	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	
	// 8. Esc cases
	m.state = stateWorkflowNameInput
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(model)
	if m.state != stateBrowsing {
		t.Error("Esc failed in stateWorkflowNameInput")
	}

	m.state = stateWorkflowLocationSelection
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(model)
	if m.state != stateBrowsing {
		t.Error("Esc failed in stateWorkflowLocationSelection")
	}

	m.state = stateCreatingWorkflow
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(model)
	if m.state != stateBrowsing {
		t.Error("Esc failed in stateCreatingWorkflow")
	}

	// 9. Confirm with no selection
	m.state = stateCreatingWorkflow
	m.selectedWorkflowIndices = []int{}
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	m = updatedModel.(model)
	// 10. Navigation in stateCreatingWorkflow
	m.state = stateCreatingWorkflow
	m.focus = focusHistory
	m.historyCursor = 1
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(model)
	if m.historyCursor != 0 {
		t.Error("KeyUp failed in stateCreatingWorkflow")
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(model)
	if m.historyCursor != 1 {
		t.Error("KeyDown failed in stateCreatingWorkflow")
	}
}

func TestNavigationKeys(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	m.history = []history.HistoryEntry{{Command: "h1"}, {Command: "h2"}}
	
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	
	m.focus = focusHistory
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
}

func TestEditorFinished(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-ui-editor-*")
	defer os.RemoveAll(tmpDir)
	
	path := filepath.Join(tmpDir, "cleat.yaml")
	os.WriteFile(path, []byte("version: 1"), 0644)
	
	m := InitialModel(&config.Config{SourcePath: path}, true, "0.1.0", &executor.ShellExecutor{})
	
	updatedModel, _ := m.Update(editorFinishedMsg{err: nil})
	m = updatedModel.(model)
	if !m.cfgFound {
		t.Error("expected config to be found after editor finished")
	}
}

func TestRenderingHelpers(t *testing.T) {
    m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
    m.visibleTail("styled string", 2)
    m.overlayLine("bg line", "fg", 2)
    m.overlayLine("bg", "fg", 10)
}

func TestBuildCommandTreeDeep(t *testing.T) {
    cfg := &config.Config{
        Docker: true,
        GoogleCloudPlatform: &config.GCPConfig{ProjectName: "p"},
        AppYaml: "app.yaml",
        Terraform: &config.TerraformConfig{UseFolders: true, Envs: []string{"dev"}},
        Services: []config.ServiceConfig{
            {
                Name: "s1",
                Modules: []config.ModuleConfig{
                    {Python: &config.PythonConfig{Django: true}},
                    {Npm: &config.NpmConfig{Scripts: []string{"s1"}}},
                },
                AppYaml: "s1.yaml",
                Docker: ptrBool(true),
            },
        },
    }
    tree := buildCommandTree(cfg, []config.Workflow{{Name: "w1"}})
    if len(tree) == 0 {
        t.Error("tree should not be empty")
    }
}

func TestTUIStateRendering(t *testing.T) {
    m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
    m.width = 100
    m.height = 40
    
    states := []uiState{
        stateInputCollection,
        stateConfirmClearHistory,
        stateWorkflowNameInput,
        stateWorkflowLocationSelection,
        stateShowingConfig,
    }
    
    for _, s := range states {
        m.state = s
        m.View()
    }
}

func TestHandleEnterKey(t *testing.T) {
    cfg := &config.Config{
        Services: []config.ServiceConfig{
            {Name: "svc", Modules: []config.ModuleConfig{{Python: &config.PythonConfig{Django: true}}}},
        },
    }
    m := InitialModel(cfg, true, "0.1.0", &executor.ShellExecutor{})
    m.expandAll()
    m.updateVisibleItems()
    
    // Find a leaf command
    for i, item := range m.visibleItems {
        if item.item.Command != "" {
            m.cursor = i
            break
        }
    }
    
    m.handleEnterKey()
    
    // Test with history focus
    m.focus = focusHistory
    m.history = []history.HistoryEntry{{Command: "build"}}
    m.handleEnterKey()
}

func TestBuildConfigLines(t *testing.T) {
    cfg := &config.Config{
        Version: 1,
        Docker: true,
        GoogleCloudPlatform: &config.GCPConfig{ProjectName: "p"},
        Terraform: &config.TerraformConfig{Envs: []string{"e"}},
        Services: []config.ServiceConfig{{Name: "s", Modules: []config.ModuleConfig{{Python: &config.PythonConfig{Django: true}}}}},
    }
    m := InitialModel(cfg, true, "0.1.0", &executor.ShellExecutor{})
    m.buildConfigLines()
}

func TestOpenEditor(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-test-editor-*")
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "cleat.yaml")
	m := InitialModel(&config.Config{SourcePath: configPath}, false, "0.1.0", &executor.ShellExecutor{})
	m.cfgFound = false

	// Test default config creation
	os.Setenv("EDITOR", "true") // Use 'true' as it exits immediately with 0
	defer os.Unsetenv("EDITOR")

	cmd := m.openEditor()
	if cmd == nil {
		t.Fatal("expected non-nil tea.Cmd from openEditor")
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected default config file to be created")
	}

	// Test with existing file
	m.cfgFound = true
	os.WriteFile(configPath, []byte("existing"), 0644)
	m.openEditor()
	data, _ := os.ReadFile(configPath)
	if string(data) != "existing" {
		t.Error("openEditor should not overwrite existing config")
	}
}

func TestHandleInputCollection_Advanced(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	m.state = stateInputCollection
	m.requirements = []task.InputRequirement{
		{Key: "k1", Prompt: "P1"},
		{Key: "k2", Prompt: "P2"},
	}
	m.requirementIdx = 0
	m.collectedInputs = make(map[string]string)

	// Submit first input
	m.textInput.SetValue("v1")
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if m.collectedInputs["k1"] != "v1" {
		t.Errorf("expected k1=v1, got %v", m.collectedInputs["k1"])
	}
	if m.requirementIdx != 1 {
		t.Errorf("expected requirementIdx 1, got %d", m.requirementIdx)
	}

	// Submit second input (last)
	m.textInput.SetValue("v2")
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if m.collectedInputs["k2"] != "v2" {
		t.Errorf("expected k2=v2, got %v", m.collectedInputs["k2"])
	}
	if !m.quitting {
		t.Error("expected quitting to be true after last input")
	}
	
	// Test Esc
	m.state = stateInputCollection
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(model)
	if m.state != stateBrowsing {
		t.Error("expected browsing state after Esc")
	}

	// Test Ctrl+C
	m.state = stateInputCollection
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updatedModel.(model)
	if !m.quitting {
		t.Error("expected quitting after Ctrl+C")
	}
}

func TestHandleShowingConfig_Advanced(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	m.state = stateShowingConfig
	m.focus = focusConfig
	m.previousFocus = focusCommands

	// Test Esc
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(model)
	if m.state != stateBrowsing || m.focus != focusCommands {
		t.Error("failed to return to browsing/prevFocus after Esc")
	}

	// Test 'q'
	m.state = stateShowingConfig
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = updatedModel.(model)
	if m.state != stateBrowsing {
		t.Error("failed to close config with 'q'")
	}

	// Test Ctrl+C
	m.state = stateShowingConfig
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updatedModel.(model)
	if !m.quitting {
		t.Error("expected quitting after Ctrl+C")
	}
	
	// Test 'g' jump to top
	m.state = stateShowingConfig
	m.configScrollOffset = 10
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}) // pending
	m = updatedModel.(model)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updatedModel.(model)
	if m.configScrollOffset != 0 {
		t.Errorf("expected configScrollOffset 0 after 'gg', got %d", m.configScrollOffset)
	}
}

func TestOpenEditor_Fallback(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-test-editor-fallback-*")
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "cleat.yaml")
	m := InitialModel(&config.Config{SourcePath: configPath}, false, "0.1.0", &executor.ShellExecutor{})
	
	oldEditor := os.Getenv("EDITOR")
	os.Unsetenv("EDITOR")
	defer os.Setenv("EDITOR", oldEditor)

	m.openEditor()
	// Just ensuring it doesn't crash and uses "vi" internally
}

func TestOpenEditor_WriteError(t *testing.T) {
	// Attempt to write to a path that should fail (e.g., a directory that exists)
	tmpDir, _ := os.MkdirTemp("", "cleat-test-editor-err-*")
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "bad-dir")
	os.Mkdir(configPath, 0755)
	
	m := InitialModel(&config.Config{SourcePath: configPath}, false, "0.1.0", &executor.ShellExecutor{})
	m.cfgFound = false
	
	m.openEditor()
	// Should log error but not crash
}

type mockRunner struct {
	m tea.Model
}

func (r *mockRunner) Run() (tea.Model, error) {
	return r.m, nil
}

func TestStart(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-test-start-*")
	defer os.RemoveAll(tmpDir)

	// Mock runner factory
	oldFactory := runnerFactory
	defer func() { 
		runnerFactory = oldFactory 
	}()

	t.Run("NoConfig", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "nonexistent.yaml")
		runnerFactory = func(m tea.Model) programRunner {
			mod := m.(model)
			if mod.cfgFound {
				t.Error("expected cfgFound to be false")
			}
			mod.selectedCommand = "ls"
			return &mockRunner{m: mod}
		}
		cmd, _, err := Start("0.1.0", configPath)
		if err != nil {
			t.Fatal(err)
		}
		if cmd != "ls" {
			t.Errorf("expected ls, got %q", cmd)
		}
	})

	t.Run("MalformedConfig", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "malformed.yaml")
		os.WriteFile(configPath, []byte("invalid yaml: ["), 0644)
		runnerFactory = func(m tea.Model) programRunner {
			mod := m.(model)
			if mod.fatalError == nil {
				t.Error("expected fatalError for malformed config")
			}
			return &mockRunner{m: mod}
		}
		Start("0.1.0", configPath)
	})

	t.Run("ValidConfig", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "valid.yaml")
		os.WriteFile(configPath, []byte("version: 1\ndocker: true"), 0644)
		runnerFactory = func(m tea.Model) programRunner {
			mod := m.(model)
			if !mod.cfgFound {
				t.Error("expected cfgFound to be true")
			}
			return &mockRunner{m: mod}
		}
		Start("0.1.0", configPath)
	})
}

func TestInit(t *testing.T) {
	m := model{}
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestNavigationHandlers(t *testing.T) {
    cfg := &config.Config{
        Services: []config.ServiceConfig{{Name: "s1"}, {Name: "s2"}},
    }
    m := InitialModel(cfg, true, "0.1.0", &executor.ShellExecutor{})
    m.updateVisibleItems()
    
    // Down
    m.handleDownKey()
    if m.cursor != 1 {
        t.Errorf("expected cursor 1, got %d", m.cursor)
    }
    
    // Up
    m.handleUpKey()
    if m.cursor != 0 {
        t.Errorf("expected cursor 0, got %d", m.cursor)
    }

    // History navigation
    m.focus = focusHistory
    m.history = []history.HistoryEntry{{Command: "h1"}, {Command: "h2"}}
    m.historyCursor = 0
    m.handleDownKey()
    if m.historyCursor != 1 {
        t.Errorf("expected historyCursor 1, got %d", m.historyCursor)
    }
    m.handleUpKey()
    if m.historyCursor != 0 {
        t.Errorf("expected historyCursor 0, got %d", m.historyCursor)
    }

    // Config navigation
    m.focus = focusConfig
    m.cfgFound = true
    m.cfg = &config.Config{Version: 1} // buildConfigLines will return some lines
    m.configScrollOffset = 1
    m.handleUpKey()
    if m.configScrollOffset != 0 {
        t.Errorf("expected configScrollOffset 0, got %d", m.configScrollOffset)
    }
    m.handleDownKey()
    // handleDownKey for config depends on buildConfigLines length
}
