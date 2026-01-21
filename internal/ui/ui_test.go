package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
	if !strings.Contains(view, "Tasks for build") {
		t.Error("expected view to contain 'Tasks for build' section")
	}
}

func TestConfigPreview(t *testing.T) {
	cfg := &config.Config{
		Version: 1,
		Docker:  true,
		Services: []config.ServiceConfig{
			{
				Name: "default",
				Modules: []config.ModuleConfig{
					{
						Python: &config.PythonConfig{
							Django:        true,
							DjangoService: "web",
						},
					},
					{
						Npm: &config.NpmConfig{
							Service: "node",
							Scripts: []string{"build"},
						},
					},
				},
			},
		},
	}
	m := InitialModel(cfg, true)
	m.width = 100
	m.height = 40

	view := m.View()

	if !strings.Contains(view, "version: 1") {
		t.Error("expected view to contain 'version: 1'")
	}
	if !strings.Contains(view, "docker: true") {
		t.Error("expected view to contain 'docker: true'")
	}
	if !strings.Contains(view, "python:") {
		t.Error("expected view to contain 'python:' block")
	}
	if !strings.Contains(view, "django:") {
		t.Error("expected view to contain 'django:' block under python")
	}
	if !strings.Contains(view, "django: true") {
		t.Error("expected view to contain 'django: true'")
	}
	if !strings.Contains(view, "django_service: web") {
		t.Error("expected view to contain 'django_service: web'")
	}
	if !strings.Contains(view, "npm:") {
		t.Error("expected view to contain 'npm:' block")
	}
	if !strings.Contains(view, "service: node") {
		t.Error("expected view to contain 'service: node'")
	}
}

func TestConfigPreviewFiltering(t *testing.T) {
	// Configuration with Django false and NPM disabled
	cfg := &config.Config{
		Docker: true,
		Services: []config.ServiceConfig{
			{
				Name: "default",
				Modules: []config.ModuleConfig{
					{Python: &config.PythonConfig{Django: false}},
					{Npm: &config.NpmConfig{Scripts: []string{}}},
				},
			},
		},
	}
	m := InitialModel(cfg, true)
	m.width = 100
	m.height = 40

	view := m.View()

	if strings.Contains(view, "python:") {
		t.Error("expected view NOT to contain 'python:' block when Django is false")
	}
	if strings.Contains(view, "django: false") {
		t.Error("expected view NOT to contain 'django: false' (it should be filtered out)")
	}
	if strings.Contains(view, "npm:") {
		t.Error("expected view NOT to contain 'npm:' block when no scripts are enabled")
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
	m.width = 60
	m.height = 20
	view = m.View()
	if strings.Contains(view, "Terminal too small") {
		t.Error("expected normal render at minimum dimensions")
	}
	if !strings.Contains(view, "Commands") {
		t.Error("expected 'Commands' section in normal render")
	}
}

func TestCommandTreeNesting(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
		Services: []config.ServiceConfig{
			{
				Name: "backend",
				Modules: []config.ModuleConfig{
					{Python: &config.PythonConfig{Django: true}},
				},
			},
		},
	}
	m := InitialModel(cfg, true)

	// Expected visible items (all expanded):
	// build (level 0)
	// run (level 0)
	// docker (level 0)
	//   down (level 1)
	//   rebuild (level 1)
	// backend (level 0)
	//   django (level 1)
	//     create-user-dev (level 2)
	//     collectstatic (level 2)
	//     migrate (level 2)

	expected := []struct {
		label string
		level int
	}{
		{"build", 0},
		{"run", 0},
		{"docker", 0},
		{"down", 1},
		{"rebuild", 1},
		{"backend", 0},
		{"django", 1},
		{"create-user-dev", 2},
		{"collectstatic", 2},
		{"migrate", 2},
	}

	if len(m.visibleItems) != len(expected) {
		t.Fatalf("expected %d visible items, got %d", len(expected), len(m.visibleItems))
	}

	for i, exp := range expected {
		if m.visibleItems[i].item.Label != exp.label {
			t.Errorf("item %d: expected label %q, got %q", i, exp.label, m.visibleItems[i].item.Label)
		}
		if m.visibleItems[i].level != exp.level {
			t.Errorf("item %d: expected level %d, got %d", i, exp.level, m.visibleItems[i].level)
		}
	}
}

func TestNavigation(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "frontend",
				Modules: []config.ModuleConfig{
					{Npm: &config.NpmConfig{Scripts: []string{"dev", "build"}}},
				},
			},
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
		Services: []config.ServiceConfig{
			{
				Name: "frontend",
				Modules: []config.ModuleConfig{
					{
						Npm: &config.NpmConfig{
							Scripts: []string{"script1", "script2", "script3", "script4", "script5", "script6", "script7", "script8", "script9", "script10"},
						},
					},
				},
			},
		},
	}
	m := InitialModel(cfg, true)
	m.width = 80
	m.height = 15 // Reduced to ensure scrolling

	if m.scrollOffset != 0 {
		t.Errorf("expected initial scrollOffset 0, got %d", m.scrollOffset)
	}

	// Navigate to the end
	numPresses := len(m.visibleItems) - 1
	for i := 0; i < numPresses; i++ {
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updatedModel.(model)
	}

	if m.scrollOffset == 0 {
		t.Error("expected scrollOffset to increase after navigating down")
	}

	expectedCursor := len(m.visibleItems) - 1
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

func TestTaskPreviewWrapping(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
	}
	m := InitialModel(cfg, true)
	m.width = 60
	m.height = 20

	// Trigger preview update
	m.updateTaskPreview()

	// Find a long command (like docker:rebuild)
	m.cursor = 0
	foundRebuild := false
	for i, item := range m.visibleItems {
		if item.item.Command == "docker rebuild" {
			m.cursor = i
			foundRebuild = true
			break
		}
	}

	if !foundRebuild {
		// If not found (maybe docker tree is collapsed), expand it
		for _, item := range m.visibleItems {
			if item.item.Label == "docker" {
				item.item.Expanded = true
				m.updateVisibleItems()
				// Try again
				for j, item2 := range m.visibleItems {
					if item2.item.Command == "docker rebuild" {
						m.cursor = j
						foundRebuild = true
						break
					}
				}
				break
			}
		}
	}

	if !foundRebuild {
		t.Skip("docker rebuild command not found in tree")
	}

	m.updateTaskPreview()

	// Verify that some lines in taskPreview are wrapped
	// docker rebuild has long commands
	hasWrappedLines := false
	for _, line := range m.taskPreview {
		// Strip ANSI codes for length check
		plainLine := lipgloss.Width(line)
		if plainLine > 0 {
			// paneWidth = (60-2)/2 = 29. availableWidth = 26.
			if plainLine > 26 {
				t.Errorf("line exceeds available width: %d > 26: %q", plainLine, line)
			}
			// Check if we have any continuation lines (starting with 8 spaces for commands)
			if strings.HasPrefix(ansi.Strip(line), "        ") {
				hasWrappedLines = true
			}
		}
	}

	if !hasWrappedLines {
		// It might be that the command is not long enough for 26 chars.
		// Docker rebuild's cleanup command:
		// "docker compose --profile * down --remove-orphans --rmi all --volumes"
		// Length: 69. Definitely should wrap.
		t.Error("expected wrapped lines in task preview, but found none")
	}
}

func TestConfigScrolling(t *testing.T) {
	cfg := &config.Config{
		Services: make([]config.ServiceConfig, 20),
	}
	for i := 0; i < 20; i++ {
		cfg.Services[i] = config.ServiceConfig{Name: "service-" + string(rune('a'+i))}
	}
	m := InitialModel(cfg, true)
	m.width = 100
	m.height = 20
	m.focus = focusConfig

	if m.configScrollOffset != 0 {
		t.Errorf("expected initial configScrollOffset 0, got %d", m.configScrollOffset)
	}

	// Move down
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(model)
	if m.configScrollOffset != 1 {
		t.Errorf("expected configScrollOffset 1 after Down, got %d", m.configScrollOffset)
	}

	// Move up
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(model)
	if m.configScrollOffset != 0 {
		t.Errorf("expected configScrollOffset 0 after Up, got %d", m.configScrollOffset)
	}

	// Move down multiple times
	for i := 0; i < 100; i++ {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updatedModel.(model)
	}

	if m.configScrollOffset == 0 {
		t.Error("expected configScrollOffset to be > 0 after multiple Down presses")
	}

	lastOffset := m.configScrollOffset

	// One more down should not increase offset if at end
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(model)
	if m.configScrollOffset != lastOffset {
		t.Errorf("expected configScrollOffset to be capped, but it changed from %d to %d", lastOffset, m.configScrollOffset)
	}
}

func TestInputCollectionModal(t *testing.T) {
	cfg := &config.Config{
		GoogleCloudPlatform: &config.GCPConfig{
			ProjectName: "test-project",
		},
	}
	m := InitialModel(cfg, true)

	// Navigate to gcp set-config
	found := false
	for i, item := range m.visibleItems {
		if item.item.Command == "gcp set-config" {
			m.cursor = i
			found = true
			break
		}
	}

	if !found {
		t.Fatal("gcp set-config not found in visible items")
	}

	// Press Enter to select command
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if m.state != stateInputCollection {
		t.Fatalf("expected state stateInputCollection, got %v", m.state)
	}

	if len(m.requirements) == 0 {
		t.Fatal("expected requirements for gcp set-config when account is missing")
	}

	// Simulate typing input "test@example.com"
	for _, r := range "test@example.com" {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updatedModel.(model)
	}

	// Press Enter to submit input
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)

	if !m.quitting {
		t.Error("expected quitting after filling all requirements")
	}
	if m.collectedInputs["gcp:account"] != "test@example.com" {
		t.Errorf("expected gcp:account to be 'test@example.com', got %q", m.collectedInputs["gcp:account"])
	}
}

func TestNestedCommandPathTitle(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
		Services: []config.ServiceConfig{
			{
				Name: "api",
				Modules: []config.ModuleConfig{
					{Python: &config.PythonConfig{Django: true, DjangoService: "backend"}},
				},
			},
		},
	}
	m := InitialModel(cfg, true)
	m.width = 100
	m.height = 40

	// Initial selection is "build"
	view := m.View()
	if !strings.Contains(view, "Tasks for build") {
		t.Errorf("expected 'Tasks for build', got something else")
	}

	// Move to docker > down
	// Tree: build, run, docker (down, rebuild), api (django (collectstatic, migrate))
	// In InitialModel, tree is built.
	// docker is at index 2 (0: build, 1: run, 2: docker)
	// it's expanded by default.
	// visible items:
	// 0: build
	// 1: run
	// 2: docker
	// 3:   down
	// 4:   rebuild

	m.cursor = 3 // docker.down
	m.updateTaskPreview()
	view = m.View()
	if !strings.Contains(view, "Tasks for docker.down") {
		t.Errorf("expected title 'Tasks for docker.down', view content:\n%s", view)
	}

	// Move to api > django > migrate
	// 5: api
	// 6:   django
	// 7:     create-user-dev (because Docker is true)
	// 8:     collectstatic
	// 9:     migrate

	m.cursor = 9
	m.updateTaskPreview()
	view = m.View()
	if !strings.Contains(view, "Tasks for api.django.migrate") {
		t.Errorf("expected title 'Tasks for api.django.migrate', view content:\n%s", view)
	}

	// Test with filtering
	m.filterText = "migrate"
	m.updateVisibleItems()
	m.cursor = 0 // api (because it's a parent of a match)
	// Actually api matches too because of anyDescendantMatches?
	// api -> django -> migrate.
	// if filter is "migrate":
	// api: anyDescendantMatches = true -> visible
	//   django: anyDescendantMatches = true -> visible
	//     migrate: matches = true -> visible

	// Let's find migrate in visible items
	found := false
	for i, v := range m.visibleItems {
		if v.path == "api.django.migrate" {
			m.cursor = i
			found = true
			break
		}
	}
	if !found {
		t.Fatal("could not find api.django.migrate in visible items after filtering")
	}

	m.updateTaskPreview()
	view = m.View()
	if !strings.Contains(view, "Tasks for api.django.migrate") {
		t.Errorf("expected title 'Tasks for api.django.migrate' with filter, view content:\n%s", view)
	}
}
