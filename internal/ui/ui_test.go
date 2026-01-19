package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/muesli/termenv"
)

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func TestModelUpdate(t *testing.T) {
	m := InitialModel(&config.Config{}, true)

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
	m = InitialModel(&config.Config{}, true)
	wmsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ = m.Update(wmsg)
	resModel = updatedModel.(model)
	if resModel.width != 100 || resModel.height != 40 {
		t.Errorf("expected width 100, height 40, got %d, %d", resModel.width, resModel.height)
	}
}

func TestModelView(t *testing.T) {
	m := InitialModel(&config.Config{}, true)
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
	if !strings.Contains(view, "Configuration") {
		t.Error("expected view to contain 'Configuration' section")
	}

	m.quitting = true
	view = m.View()
	if view != "" {
		t.Errorf("expected empty string when quitting, got %q", view)
	}
}

func TestSmallDimensions(t *testing.T) {
	m := InitialModel(&config.Config{}, true)

	// Test that small dimensions show the "too small" message
	m.width = 40
	m.height = 10
	view := m.View()
	if !strings.Contains(view, "Terminal too small") {
		t.Error("expected 'Terminal too small' message for undersized terminal")
	}

	// Test that adequate dimensions render normally
	m.width = 80
	m.height = 20
	view = m.View()
	if strings.Contains(view, "Terminal too small") {
		t.Error("expected normal render at minimum dimensions")
	}
	if !strings.Contains(view, "Commands") {
		t.Error("expected 'Commands' section in normal render")
	}
}

func TestNavigation(t *testing.T) {
	cfg := &config.Config{
		Npm: config.NpmConfig{
			Scripts: []string{"dev", "build"},
		},
	}
	m := InitialModel(cfg, true)

	if m.cursor != 0 {
		t.Errorf("expected initial cursor 0, got %d", m.cursor)
	}

	// Move down
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(model)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after Down, got %d", m.cursor)
	}

	// Move up
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after Up, got %d", m.cursor)
	}
}

func TestEnterKey(t *testing.T) {
	m := InitialModel(&config.Config{}, true)

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected non-nil command after Enter")
	}
	resModel := updatedModel.(model)
	if !resModel.quitting {
		t.Error("expected quitting to be true after Enter")
	}
	if resModel.selectedCommand != "build" {
		t.Errorf("expected selectedCommand to be 'build', got %q", resModel.selectedCommand)
	}
}

func TestTabbing(t *testing.T) {
	m := InitialModel(&config.Config{}, true)

	if m.focus != focusCommands {
		t.Error("expected initial focus to be commands")
	}

	// Tab -> Config
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focus != focusConfig {
		t.Errorf("expected focus to be config after Tab, got %v", m.focus)
	}

	// Shift+Tab -> Commands
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updatedModel.(model)
	if m.focus != focusCommands {
		t.Errorf("expected focus to be commands after Shift+Tab, got %v", m.focus)
	}
}

func TestNoConfigMessage(t *testing.T) {
	m := InitialModel(&config.Config{}, false)
	m.width = 100
	m.height = 40

	view := m.View()
	if !strings.Contains(view, "No cleat.yaml found") {
		t.Error("expected view to contain 'No cleat.yaml found' in config pane when cfgFound is false")
	}
	if !strings.Contains(view, "(no cleat.yaml)") {
		t.Error("expected view to contain '(no cleat.yaml)' in help bar when cfgFound is false")
	}
}

func TestConfigPaneAction(t *testing.T) {
	m := InitialModel(&config.Config{}, false)
	m.width = 100
	m.height = 40

	// Tab to config pane
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)

	view := m.View()
	if !strings.Contains(view, "Press Enter to create") {
		t.Error("expected 'Press Enter to create' hint when config not found and pane focused")
	}

	// With config found
	m2 := InitialModel(&config.Config{}, true)
	m2.width = 100
	m2.height = 40
	m2.focus = focusConfig

	view2 := m2.View()
	if !strings.Contains(view2, "Press Enter to edit") {
		t.Error("expected 'Press Enter to edit' hint when config exists and pane focused")
	}
}

func TestScrolling(t *testing.T) {
	cfg := &config.Config{
		Npm: config.NpmConfig{
			Scripts: []string{"script1", "script2", "script3", "script4", "script5", "script6", "script7", "script8", "script9", "script10"},
		},
	}
	m := InitialModel(cfg, true)
	m.width = 80
	m.height = 15 // Reduced to ensure scrolling

	if m.scrollOffset != 0 {
		t.Errorf("expected initial scrollOffset 0, got %d", m.scrollOffset)
	}

	// Navigate to the end
	numPresses := len(m.commands) - 1
	for i := 0; i < numPresses; i++ {
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updatedModel.(model)
	}

	if m.scrollOffset == 0 {
		t.Error("expected scrollOffset to increase after navigating down")
	}

	expectedCursor := len(m.commands) - 1
	if m.cursor != expectedCursor {
		t.Errorf("expected cursor at %d, got %d", expectedCursor, m.cursor)
	}

	// Navigate back up
	for i := 0; i < numPresses; i++ {
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
		m = updatedModel.(model)
	}

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.cursor)
	}
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0 after scrolling back up, got %d", m.scrollOffset)
	}
}

func TestCursorDimmedWhenUnfocused(t *testing.T) {
	m := InitialModel(&config.Config{}, true)
	m.width = 100
	m.height = 40

	// When commands pane is focused, cursor should be purple (#bd93f9)
	view1 := m.View()
	// TrueColor ANSI for #bd93f9 is 189;147;249
	if !strings.Contains(view1, "189;147;249") {
		t.Error("expected purple cursor color when commands pane is focused")
	}

	// Tab to config pane
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)

	// When config pane is focused, cursor should be dimmed (comment color #6272a4)
	view2 := m.View()
	lines := strings.Split(view2, "\n")
	foundDimmedCursor := false
	for _, line := range lines {
		if strings.Contains(line, ">") && strings.Contains(line, "build") {
			// TrueColor ANSI for #6272a4 might be 97;113;163 or 98;114;164 depending on profile
			if strings.Contains(line, "97;113;163") || strings.Contains(line, "98;114;164") {
				foundDimmedCursor = true
			}
		}
	}
	if !foundDimmedCursor {
		t.Error("expected dimmed cursor color when commands pane is not focused")
	}
}

func TestHelpOverlay(t *testing.T) {
	m := InitialModel(&config.Config{}, true)
	m.width = 100
	m.height = 40

	if m.showHelp {
		t.Error("expected showHelp to be false initially")
	}

	// Press '?' to show help
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = updatedModel.(model)

	if !m.showHelp {
		t.Error("expected showHelp to be true after pressing '?'")
	}

	view := m.View()
	if !strings.Contains(view, "Keyboard Shortcuts") {
		t.Error("expected help overlay to contain 'Keyboard Shortcuts'")
	}

	// Press any key to dismiss
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	m = updatedModel.(model)

	if m.showHelp {
		t.Error("expected showHelp to be false after pressing any key")
	}
}

func TestEscToQuit(t *testing.T) {
	m := InitialModel(&config.Config{}, true)

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	resModel := updatedModel.(model)

	if !resModel.quitting {
		t.Error("expected quitting to be true after pressing Esc")
	}
	if cmd == nil {
		t.Error("expected a non-nil command after pressing Esc")
	}
}
