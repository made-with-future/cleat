package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
)

func TestFilteringWithSpace(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
	}
	m := InitialModel(cfg, true, "0.1.0", &executor.ShellExecutor{})
	m.width = 100
	m.height = 40

	// Start filtering
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = updatedModel.(model)

	// Type "docker"
	for _, r := range "docker" {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updatedModel.(model)
	}

	// Type space
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	m = updatedModel.(model)

	if !strings.HasSuffix(m.filterText, " ") {
		t.Errorf("expected filterText to end with space, got %q", m.filterText)
	}

	// Type "up"
	for _, r := range "up" {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updatedModel.(model)
	}

	if m.filterText != "docker up" {
		t.Errorf("expected filterText 'docker up', got %q", m.filterText)
	}

	// Verify it matches "docker up"
	found := false
	for _, item := range m.visibleItems {
		if strings.Contains(strings.ToLower(item.item.Command), "docker up") ||
			(strings.Contains(strings.ToLower(item.path), "docker") && strings.Contains(strings.ToLower(item.path), "up")) {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find item matching 'docker up'")
	}
}

func TestFilteringWildcard(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
	}
	m := InitialModel(cfg, true, "0.1.0", &executor.ShellExecutor{})
	m.width = 100
	m.height = 40

	// Start filtering
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = updatedModel.(model)

	// Type "doc up" (wildcard matching "docker up")
	for _, r := range "doc up" {
		var msg tea.KeyMsg
		if r == ' ' {
			msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
		} else {
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		}
		updatedModel, _ = m.Update(msg)
		m = updatedModel.(model)
	}

	found := false
	for _, item := range m.visibleItems {
		// "doc up" should match "docker up" command
		if strings.Contains(strings.ToLower(item.item.Command), "docker up") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'doc up' to match 'docker up'")
	}
}

func TestFilteringOrderSensitive(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
	}
	m := InitialModel(cfg, true, "0.1.0", &executor.ShellExecutor{})
	m.width = 100
	m.height = 40

	// Start filtering
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = updatedModel.(model)

	// Type "up doc" (should NOT match "docker up" if order sensitive)
	for _, r := range "up doc" {
		var msg tea.KeyMsg
		if r == ' ' {
			msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
		} else {
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		}
		updatedModel, _ = m.Update(msg)
		m = updatedModel.(model)
	}

	found := false
	for _, item := range m.visibleItems {
		if strings.Contains(strings.ToLower(item.item.Command), "docker up") {
			found = true
			break
		}
	}
	if found {
		t.Error("expected 'up doc' NOT to match 'docker up' (order sensitive)")
	}
}
