package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelUpdate(t *testing.T) {
	m := InitialModel()

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

	// Test quitting with 'ctrl+c'
	msg = tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedModel, cmd = m.Update(msg)

	resModel = updatedModel.(model)
	if !resModel.quitting {
		t.Error("expected quitting to be true after pressing 'ctrl+c'")
	}
}

func TestModelView(t *testing.T) {
	m := InitialModel()
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}

	m.quitting = true
	view = m.View()
	if view != "Bye!\n" {
		t.Errorf("expected 'Bye!\\n', got %q", view)
	}
}
