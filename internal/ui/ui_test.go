package ui

import (
	"errors"
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/history"
	"github.com/muesli/termenv"
)

func ptrBool(b bool) *bool {
	return &b
}

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
	// Mock home directory for tests to avoid loading real history/workflows
	testHomeDir, _ := os.MkdirTemp("", "cleat-ui-test-home-*")
	history.UserHomeDir = func() (string, error) {
		return testHomeDir, nil
	}
}

func TestModelUpdate(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})

	// Test quitting with 'q'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	updatedModel, cmd := m.Update(msg)

	resModel := updatedModel.(model)
	if !resModel.quitting {
		t.Error("expected quitting to be true after pressing 'q'")
	}
	if cmd == nil {
		t.Error("expected a non-nil command after pressing 'q'")
	}

	// Test window resize
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
	// Configuration is now in a modal, not visible initially
	if strings.Contains(view, "Configuration") {
		t.Error("expected view NOT to contain 'Configuration' section initially")
	}

	if !strings.Contains(view, "Command History") {
		t.Error("expected view to contain 'Command History' section")
	}
	if !strings.Contains(view, "Tasks for build") {
		t.Error("expected view to contain 'Tasks for build' section")
	}

	// Press 'c' to open configuration modal
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

	// Check if any key clears it
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

	// Press '/' to start filtering
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = updatedModel.(model)
	if !m.filtering {
		t.Error("expected filtering to be true")
	}

	// Type 'down'
	for _, r := range "down" {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updatedModel.(model)
	}
	if m.filterText != "down" {
		t.Errorf("expected filterText 'down', got %q", m.filterText)
	}

	// Verify only matching items (or parents of matching items) visible
	foundDown := false
	for _, item := range m.visibleItems {
		if strings.Contains(item.item.Label, "down") || strings.Contains(item.path, "down") {
			foundDown = true
		}
	}
	if !foundDown {
		t.Error("expected to find item matching 'down' in filtered results")
	}

	// Backspace
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updatedModel.(model)
	if m.filterText != "dow" {
		t.Errorf("expected filterText 'dow' after backspace, got %q", m.filterText)
	}

	// Escape to exit filter
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(model)
	if m.filtering {
		t.Error("expected filtering to be false after Esc")
	}
}

func TestModelInit(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	cmd := m.Init()
	if cmd != nil {
		t.Error("expected Init to return nil")
	}
}

func TestTabbing(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})

	if m.focus != focusCommands {
		t.Error("expected initial focus to be commands")
	}

	// Tab -> History
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focus != focusHistory {
		t.Errorf("expected focus to be history after Tab, got %v", m.focus)
	}

	// Tab -> Commands
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focus != focusCommands {
		t.Errorf("expected focus to be commands after 2nd Tab, got %v", m.focus)
	}
}

func TestHelpOverlay(t *testing.T) {
	m := InitialModel(&config.Config{}, true, "0.1.0", &executor.ShellExecutor{})
	m.width = 100
	m.height = 40

	// Press '?'
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = updatedModel.(model)
	if !m.showHelp {
		t.Error("expected showHelp to be true")
	}

	view := m.View()
	if !strings.Contains(view, "Keyboard Shortcuts") {
		t.Error("expected view to contain 'Keyboard Shortcuts'")
	}

	// Press any key to dismiss
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	m = updatedModel.(model)
	if m.showHelp {
		t.Error("expected showHelp to be false after pressing a key")
	}
}

func TestModelUpdateNavigation(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
	}
	m := InitialModel(cfg, true, "0.1.0", &executor.ShellExecutor{})
	m.width = 100
	m.height = 40

	// Down
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(model)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after Down, got %d", m.cursor)
	}

	// Up
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after Up, got %d", m.cursor)
	}
}

func TestModelUpdateEnter(t *testing.T) {
	cfg := &config.Config{}
	m := InitialModel(cfg, true, "0.1.0", &executor.ShellExecutor{})
	m.width = 100
	m.height = 40

	// Press Enter on 'build'
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("expected a command after pressing Enter")
	}
	resModel := updatedModel.(model)
	if !resModel.quitting {
		t.Error("expected quitting after selecting a command")
	}
}

func TestTUIKeys(t *testing.T) {
	cfg := &config.Config{Docker: true}
	m := InitialModel(cfg, true, "0.1.0", &executor.ShellExecutor{})
	m.width = 100
	m.height = 40

	// Expand all 'e'
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = updatedModel.(model)
	
	// Collapse all 'C'
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("C")})
	m = updatedModel.(model)
	
	// Jump to tasks 't'
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	m = updatedModel.(model)
	if m.focus != focusTasks {
		t.Error("expected focus tasks after 't'")
	}
}

func TestTerraformTree(t *testing.T) {
	cfg := &config.Config{
		Terraform: &config.TerraformConfig{
			UseFolders: true,
			Envs: []string{"prod"},
		},
	}
	m := InitialModel(cfg, true, "0.1.0", &executor.ShellExecutor{})
	m.expandAll()
	m.updateVisibleItems()
	
	foundProd := false
	for _, item := range m.visibleItems {
		if strings.Contains(item.path, "terraform.prod") {
			foundProd = true
			break
		}
	}
	if !foundProd {
		t.Error("expected terraform.prod in tree")
	}
}
