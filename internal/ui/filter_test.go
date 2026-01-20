package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/madewithfuture/cleat/internal/config"
)

func TestFiltering(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
		Services: []config.ServiceConfig{
			{
				Name: "backend",
				Modules: []config.ModuleConfig{
					{Python: &config.PythonConfig{Django: true}},
				},
			},
			{
				Name: "frontend",
				Modules: []config.ModuleConfig{
					{Npm: &config.NpmConfig{Scripts: []string{"dev"}}},
				},
			},
		},
	}
	m := InitialModel(cfg, true)

	// Initial count: build, run, docker (down, rebuild), backend (django (create-user-dev, collectstatic, migrate)), frontend (npm (run dev))
	// Total: 1 (build) + 1 (run) + 1 (docker) + 2 (docker children) + 1 (backend) + 1 (django) + 3 (django children) + 1 (frontend) + 1 (npm) + 1 (npm child) = 13 items
	if len(m.visibleItems) == 0 {
		t.Fatal("expected visible items")
	}

	// 1. Enter filtering mode
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = updatedModel.(model)

	if !m.filtering {
		t.Error("expected filtering to be true")
	}

	// 2. Type "mig"
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	updatedModel, _ = updatedModel.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	updatedModel, _ = updatedModel.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updatedModel.(model)

	if m.filterText != "mig" {
		t.Errorf("expected filterText 'mig', got %q", m.filterText)
	}

	// Should show backend -> django -> migrate
	foundMigrate := false
	for _, item := range m.visibleItems {
		if item.item.Label == "migrate" {
			foundMigrate = true
			break
		}
	}
	if !foundMigrate {
		t.Error("expected to find 'migrate' in visible items when filtering for 'mig'")
	}

	// 3. Backspace
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updatedModel.(model)
	if m.filterText != "mi" {
		t.Errorf("expected filterText 'mi' after backspace, got %q", m.filterText)
	}

	// 4. Clear filter
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})                    // "mi" -> "m"
	updatedModel, _ = updatedModel.(model).Update(tea.KeyMsg{Type: tea.KeyBackspace}) // "m" -> ""
	m = updatedModel.(model)
	if !m.filtering || m.filterText != "" {
		t.Error("expected to still be filtering with empty text after backspacing to empty")
	}
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace}) // "" -> exit filtering
	m = updatedModel.(model)
	if m.filtering {
		t.Error("expected filtering to be false after backspacing when already empty")
	}

	// 5. Cancel with Esc
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	updatedModel, _ = updatedModel.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	updatedModel, _ = updatedModel.(model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(model)
	if m.filtering || m.filterText != "" {
		t.Error("expected filtering to be reset after Esc")
	}

	// 6. Enter to run
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	updatedModel, _ = updatedModel.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	updatedModel, _ = updatedModel.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	// Cursor should be on "build" (first item)
	m = updatedModel.(model)
	if m.visibleItems[m.cursor].item.Label != "build" {
		t.Errorf("expected cursor on 'build', got %q", m.visibleItems[m.cursor].item.Label)
	}
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if !m.quitting {
		t.Error("expected model to be quitting after Enter on command")
	}
	if m.selectedCommand != "build" {
		t.Errorf("expected selectedCommand 'build', got %q", m.selectedCommand)
	}
}

func TestFilterView(t *testing.T) {
	m := InitialModel(&config.Config{}, true)
	m.width = 100
	m.height = 40

	// Start filtering
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	updatedModel, _ = updatedModel.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	updatedModel, _ = updatedModel.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	updatedModel, _ = updatedModel.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m = updatedModel.(model)

	view := m.View()
	if !strings.Contains(view, "/tes█") {
		t.Errorf("expected view to contain search string and cursor, got %q", view)
	}

	// Stop filtering but keep text
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	view = m.View()
	if !strings.Contains(view, "/tes") {
		t.Error("expected view to contain search string after Enter")
	}
	if strings.Contains(view, "█") {
		t.Error("expected view NOT to contain cursor after Enter")
	}
}
