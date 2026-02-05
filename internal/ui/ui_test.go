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

func TestWorkflowLocationSelection(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-ui-wf-loc-*")
	defer os.RemoveAll(tmpDir)
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	
	os.WriteFile("cleat.yaml", []byte("version: 1"), 0644)

	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	m.state = stateWorkflowLocationSelection
	m.selectedWorkflowIndices = []int{0}
	m.history = []history.HistoryEntry{{Command: "build", Timestamp: time.Now()}}
	m.textInput.SetValue("my-wf")
	
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if m.state != stateBrowsing {
		t.Error("expected state browsing after saving")
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

func TestBuildHistoryContent(t *testing.T) {
    m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
    m.width = 100
    m.history = []history.HistoryEntry{{Command: "c1", Success: true, Timestamp: time.Now()}}
    m.buildHistoryContent(lipgloss.Color("1"), lipgloss.Color("2"), lipgloss.Color("3"), lipgloss.Color("4"), lipgloss.Color("5"), 50)
    
    m.state = stateCreatingWorkflow
    m.buildHistoryContent(lipgloss.Color("1"), lipgloss.Color("2"), lipgloss.Color("3"), lipgloss.Color("4"), lipgloss.Color("5"), 50)
}
