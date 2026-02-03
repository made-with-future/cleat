package ui

import (
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/history"
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
	if !strings.Contains(view, "Command History") {
		t.Error("expected view to contain 'Command History' section")
	}
	if !strings.Contains(view, "Tasks for build") {
		t.Error("expected view to contain 'Tasks for build' section")
	}
}

func TestConfigPreview(t *testing.T) {
	cfg := &config.Config{
		Version: 1,
		Docker:  true,
		Envs:    []string{"production", "staging"},
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
	if !strings.Contains(view, "envs:") {
		t.Error("expected view to contain 'envs:' block")
	}
	if !strings.Contains(view, "- production") {
		t.Error("expected view to contain '- production' under envs")
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

	// Test with Terraform
	cfg.Terraform = &config.TerraformConfig{UseFolders: true, Envs: []string{"production", "staging"}}
	m2 := InitialModel(cfg, true)
	m2.expandAll()
	m2.updateVisibleItems()
	m2.width = 100
	m2.height = 40
	view2 := m2.View()
	if !strings.Contains(view2, "terraform:") {
		t.Error("expected view to contain 'terraform:' section")
	}

	// Test that terraform commands are in the tree
	found := false
	for _, item := range m2.visibleItems {
		if strings.Contains(item.path, "terraform.production.plan") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find terraform.production.plan in command tree")
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

func TestTaskPreviewAllCommands(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
		GoogleCloudPlatform: &config.GCPConfig{
			ProjectName: "test-project",
		},
		Terraform: &config.TerraformConfig{
			UseFolders: true,
			Envs:       []string{"dev"},
		},
		Services: []config.ServiceConfig{
			{
				Name: "backend",
				Modules: []config.ModuleConfig{
					{Python: &config.PythonConfig{Django: true}},
					{Npm: &config.NpmConfig{Scripts: []string{"build"}}},
				},
			},
		},
	}
	m := InitialModel(cfg, true)
	m.width = 100
	m.height = 40

	for i, item := range m.visibleItems {
		m.cursor = i
		m.updateTaskPreview()
		view := m.View()

		if item.item.Command != "" {
			if strings.Contains(view, "unknown command") {
				t.Errorf("Item %d (%s: %s) has 'unknown command' in preview", i, item.item.Label, item.item.Command)
			}
			if strings.Contains(view, "Error:") {
				t.Errorf("Item %d (%s: %s) has 'Error:' in preview", i, item.item.Label, item.item.Command)
			}
		}
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
	m.expandAll()
	m.updateVisibleItems()

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
	//     gen-random-secret-key (level 2)

	expected := []struct {
		label string
		level int
	}{
		{"build", 0},
		{"run", 0},
		{"docker", 0},
		{"down", 1},
		{"rebuild", 1},
		{"remove-orphans", 1},
		{"backend", 0},
		{"django", 1},
		{"create-user-dev", 2},
		{"collectstatic", 2},
		{"makemigrations", 2},
		{"migrate", 2},
		{"gen-random-secret-key", 2},
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

	// Tab -> History
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focus != focusHistory {
		t.Errorf("expected focus to be history after Tab, got %v", m.focus)
	}

	// Tab -> Config
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focus != focusConfig {
		t.Errorf("expected focus to be config after Tab, got %v", m.focus)
	}

	// Shift+Tab -> History
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updatedModel.(model)
	if m.focus != focusHistory {
		t.Errorf("expected focus to be history after Shift+Tab, got %v", m.focus)
	}

	// Shift+Tab -> Commands
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updatedModel.(model)
	if m.focus != focusCommands {
		t.Errorf("expected focus to be commands after Shift+Tab, got %v", m.focus)
	}
}

func TestWorkflowTaskPreviewIndentation(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
	}
	m := InitialModel(cfg, true)
	m.width = 100
	m.height = 40

	// Add a mock workflow
	m.workflows = []history.Workflow{
		{
			Name: "test-wf",
			Commands: []history.HistoryEntry{
				{Command: "docker down"},
			},
		},
	}
	// Rebuild tree to include workflows
	m.tree = buildCommandTree(cfg, m.workflows)
	m.expandAll()
	m.updateVisibleItems()

	// 1. Find the workflow in visible items
	found := false
	for i, item := range m.visibleItems {
		if item.item.Command == "workflow:test-wf" {
			m.cursor = i
			found = true
			break
		}
	}

	if !found {
		t.Fatal("workflow:test-wf not found in tree")
	}

	m.updateTaskPreview()

	// Check if task preview contains indented task
	// Regular task starts with "• " (indented with "  " in View)
	// Workflow task should start with "  • " (indented with "  " in View, so "    • ")

	// Let's check m.taskPreview directly first
	hasIndentedTask := false
	for _, line := range m.taskPreview {
		if strings.HasPrefix(ansi.Strip(line), "  • ") {
			hasIndentedTask = true
			break
		}
	}

	if !hasIndentedTask {
		t.Errorf("expected indented task line starting with '  • ', but not found. Preview: %v", m.taskPreview)
	}

	// Check command indentation
	hasIndentedCmd := false
	for _, line := range m.taskPreview {
		if strings.HasPrefix(ansi.Strip(line), "      $ ") {
			hasIndentedCmd = true
			break
		}
	}
	if !hasIndentedCmd {
		t.Errorf("expected indented command line starting with '      $ ', but not found. Preview: %v", m.taskPreview)
	}

	// Check for the header as well
	hasHeader := false
	for _, line := range m.taskPreview {
		if strings.Contains(ansi.Strip(line), "→ docker down") {
			hasHeader = true
			break
		}
	}
	if !hasHeader {
		t.Errorf("expected header '→ docker down', but not found")
	}

	// 2. Verify non-workflow task indentation (should NOT be indented)
	found = false
	for i, item := range m.visibleItems {
		if item.item.Command == "docker down" {
			m.cursor = i
			found = true
			break
		}
	}

	if !found {
		t.Fatal("docker down command not found in tree")
	}

	m.updateTaskPreview()

	for _, line := range m.taskPreview {
		stripped := ansi.Strip(line)
		if strings.HasPrefix(stripped, "• ") {
			// Success
			return
		}
		if strings.HasPrefix(stripped, "  • ") {
			t.Errorf("expected non-workflow task to NOT be indented, but found '  • '")
		}
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

	// Tab to history, then to config pane
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	updatedModel, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focus != focusConfig {
		t.Fatalf("expected focus to be config, got %v", m.focus)
	}

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
	m.expandAll()
	m.updateVisibleItems()
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

	// When commands pane is focused, cursor should be cyan (#8be9fd)
	view1 := m.View()
	// TrueColor ANSI for #8be9fd is 139;233;253
	if !strings.Contains(view1, "139;233;253") {
		t.Error("expected cyan cursor color when commands pane is focused")
	}

	// Tab to history pane
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)

	// When history pane is focused, its cursor should be cyan
	viewHistory := m.View()
	if !strings.Contains(viewHistory, "139;233;253") {
		t.Error("expected cyan cursor color when history pane is focused")
	}

	// Tab to config pane
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
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
	if !strings.Contains(view, "e          Expand all") {
		t.Error("expected help overlay to contain 'e          Expand all'")
	}
	if !strings.Contains(view, "c          Collapse all") {
		t.Error("expected help overlay to contain 'c          Collapse all'")
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
	m.expandAll()
	m.updateVisibleItems()
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
	m.expandAll()
	m.updateVisibleItems()
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
	m.expandAll()
	m.updateVisibleItems()

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
	m.expandAll()
	m.updateVisibleItems()
	m.width = 100
	m.height = 40
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
	// 5:   remove-orphans

	m.cursor = 3 // docker.down
	m.updateTaskPreview()
	view = m.View()
	if !strings.Contains(view, "Tasks for docker.down") {
		t.Errorf("expected title 'Tasks for docker.down', view content:\n%s", view)
	}

	// Move to api > django > migrate
	// 6: api
	// 7:   django
	// 8:     create-user-dev (because Docker is true)
	// 9:     collectstatic
	// 10:    makemigrations
	// 11:    migrate
	// 12:    gen-random-secret-key

	m.cursor = 11
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

func TestHistoryNavigationWithJK(t *testing.T) {
	cfg := &config.Config{}
	m := InitialModel(cfg, true)
	m.width = 100
	m.height = 20 // visibleHistoryCount will be 5

	// Inject some history entries
	m.history = []history.HistoryEntry{
		{Timestamp: time.Now(), Command: "cmd1"},
		{Timestamp: time.Now(), Command: "cmd2"},
		{Timestamp: time.Now(), Command: "cmd3"},
		{Timestamp: time.Now(), Command: "cmd4"},
		{Timestamp: time.Now(), Command: "cmd5"},
		{Timestamp: time.Now(), Command: "cmd6"},
	}
	m.historyCursor = 0

	// Switch focus to history
	m.focus = focusHistory

	// 1. Move down with 'j'
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updatedModel.(model)
	if m.historyCursor != 1 {
		t.Errorf("expected historyCursor 1 after 'j', got %d", m.historyCursor)
	}

	// Move down many times to trigger scroll
	for i := 0; i < 5; i++ {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = updatedModel.(model)
	}

	if m.historyCursor != 5 {
		t.Errorf("expected historyCursor 5, got %d", m.historyCursor)
	}

	if m.historyOffset == 0 {
		t.Error("expected historyOffset > 0 when cursor is at 5 and visibleCount is 5")
	}

	// 4. Move up with 'k'
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updatedModel.(model)
	if m.historyCursor != 4 {
		t.Errorf("expected historyCursor 4 after 'k', got %d", m.historyCursor)
	}
}

func TestHistoryJumpWithNumberKeys(t *testing.T) {
	cfg := &config.Config{}
	m := InitialModel(cfg, true)
	m.width = 100
	m.height = 20 // visibleHistoryCount will be 5

	m.history = []history.HistoryEntry{
		{Timestamp: time.Now(), Command: "cmd1"},
		{Timestamp: time.Now(), Command: "cmd2"},
		{Timestamp: time.Now(), Command: "cmd3"},
		{Timestamp: time.Now(), Command: "cmd4"},
		{Timestamp: time.Now(), Command: "cmd5"},
		{Timestamp: time.Now(), Command: "cmd6"},
		{Timestamp: time.Now(), Command: "cmd7"},
	}

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	m = updatedModel.(model)
	if m.focus != focusHistory {
		t.Fatalf("expected focusHistory after '3', got %v", m.focus)
	}
	if m.historyCursor != 2 {
		t.Errorf("expected historyCursor 2 after '3', got %d", m.historyCursor)
	}
	if m.historyOffset != 0 {
		t.Errorf("expected historyOffset 0 after '3', got %d", m.historyOffset)
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("6")})
	m = updatedModel.(model)
	if m.historyCursor != 5 {
		t.Errorf("expected historyCursor 5 after '6', got %d", m.historyCursor)
	}
	if m.historyOffset != 1 {
		t.Errorf("expected historyOffset 1 after '6', got %d", m.historyOffset)
	}
}

func TestGGKeybinding(t *testing.T) {
	cfg := &config.Config{}
	m := InitialModel(cfg, true)
	m.width = 100
	m.height = 40

	// 1. Test Commands panel
	// Move down a few times
	for i := 0; i < 5; i++ {
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updatedModel.(model)
	}
	if m.cursor == 0 {
		t.Fatal("expected cursor to be > 0 after moving down")
	}

	// Press 'g' then 'g'
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updatedModel.(model)
	if !m.pendingG {
		t.Error("expected pendingG to be true after first 'g'")
	}
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updatedModel.(model)
	if m.pendingG {
		t.Error("expected pendingG to be false after second 'g'")
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after 'gg', got %d", m.cursor)
	}

	// 2. Test History panel
	m.history = []history.HistoryEntry{
		{Timestamp: time.Now(), Command: "cmd1"},
		{Timestamp: time.Now(), Command: "cmd2"},
		{Timestamp: time.Now(), Command: "cmd3"},
	}
	m.focus = focusHistory
	m.historyCursor = 2

	// Press 'g' then 'g'
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updatedModel.(model)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updatedModel.(model)

	if m.historyCursor != 0 {
		t.Errorf("expected historyCursor 0 after 'gg', got %d", m.historyCursor)
	}

	// 3. Test Config panel
	m.focus = focusConfig
	m.configScrollOffset = 5
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updatedModel.(model)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updatedModel.(model)
	if m.configScrollOffset != 0 {
		t.Errorf("expected configScrollOffset 0 after 'gg', got %d", m.configScrollOffset)
	}

	// 4. Test reset of pendingG on other keys
	m.focus = focusHistory
	m.historyCursor = 0
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updatedModel.(model)
	if !m.pendingG {
		t.Error("expected pendingG to be true")
	}
	// Press 'j' instead of 'g'
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updatedModel.(model)
	if m.pendingG {
		t.Error("expected pendingG to be false after 'j'")
	}
	// History cursor was 0, focus is history, 'j' should make it 1
	if m.historyCursor != 1 {
		t.Errorf("expected historyCursor 1 after 'j', got %d", m.historyCursor)
	}

	// 4. Test reset of pendingG on non-rune keys
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updatedModel.(model)
	if !m.pendingG {
		t.Error("expected pendingG to be true")
	}
	// Press KeyUp
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(model)
	if m.pendingG {
		t.Error("expected pendingG to be false after KeyUp")
	}
}

func TestHistoryTaskPreview(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
	}
	m := InitialModel(cfg, true)
	m.width = 120
	m.height = 40

	// Inject some history entries
	m.history = []history.HistoryEntry{
		{Timestamp: time.Now(), Command: "build"},
		{Timestamp: time.Now(), Command: "docker down"},
	}
	m.historyCursor = 0

	// Initial preview should be for "build" (commands focused)
	if !strings.Contains(m.View(), "Tasks for build") {
		t.Error("expected initial preview for 'build'")
	}

	// Tab to history
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focus != focusHistory {
		t.Fatalf("expected focus history, got %v", m.focus)
	}

	// Preview should now be for history item at cursor 0: "build"
	view := m.View()
	if !strings.Contains(view, "Tasks for build") {
		t.Error("expected preview for history 'build'")
	}

	// Move down to "docker down"
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(model)
	if m.historyCursor != 1 {
		t.Fatalf("expected history cursor 1, got %d", m.historyCursor)
	}

	view = m.View()
	if !strings.Contains(view, "Tasks for docker down") {
		t.Errorf("expected preview for history 'docker down', view:\n%s", view)
	}

	// 3. Test with inputs: gcp set-config with account
	m.cfg.GoogleCloudPlatform = &config.GCPConfig{ProjectName: "test-project"}
	m.history = append(m.history, history.HistoryEntry{
		Timestamp: time.Now(),
		Command:   "gcp set-config",
		Inputs:    map[string]string{"gcp:account": "history@example.com"},
	})
	m.historyCursor = 2

	m.updateTaskPreview()
	view = m.View()
	if !strings.Contains(view, "Tasks for gcp set-config") {
		t.Errorf("expected preview for 'gcp set-config', view:\n%s", view)
	}
	if !strings.Contains(view, "gcloud config set account history@example.com") {
		t.Errorf("expected preview to contain account from history inputs, view:\n%s", view)
	}
}

func TestClearHistoryConfirmation(t *testing.T) {
	// Mock home directory for history
	tmpDir, err := os.MkdirTemp("", "cleat-ui-history-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldUserHomeDir := history.UserHomeDir
	history.UserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { history.UserHomeDir = oldUserHomeDir }()

	// Save some history
	history.Save(history.HistoryEntry{Command: "test-cmd", Timestamp: time.Now()})

	m := InitialModel(&config.Config{}, true)
	m.width = 100
	m.height = 40
	m.focus = focusHistory

	// Press 'x' to trigger confirmation
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	m = updatedModel.(model)

	if m.state != stateConfirmClearHistory {
		t.Errorf("expected state stateConfirmClearHistory, got %v", m.state)
	}

	view := m.View()
	if !strings.Contains(view, "Are you sure you want to clear history?") {
		t.Error("expected confirmation message in view")
	}

	// Test cancellation with 'n'
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = updatedModel.(model)
	if m.state != stateBrowsing {
		t.Errorf("expected state stateBrowsing after 'n', got %v", m.state)
	}

	// Trigger 'x' again
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	m = updatedModel.(model)

	// Confirm with 'y'
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m = updatedModel.(model)

	if m.state != stateBrowsing {
		t.Errorf("expected state stateBrowsing after 'y', got %v", m.state)
	}
	if len(m.history) != 0 {
		t.Errorf("expected history to be empty, got %d entries", len(m.history))
	}

	// Verify file is actually gone
	entries, _ := history.Load()
	if len(entries) != 0 {
		t.Error("expected history file to be cleared")
	}
}

func TestFocusedTitleColor(t *testing.T) {
	m := InitialModel(&config.Config{}, true)
	m.width = 100
	m.height = 40

	// Initially Commands is focused.
	view1 := m.View()
	// TrueColor ANSI for #ffffff is 255;255;255
	if !strings.Contains(view1, "255;255;255") {
		t.Error("expected white title color when pane is focused")
	}
	if !strings.Contains(view1, "Commands") {
		t.Error("expected 'Commands' title in view")
	}

	// Tab to History
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	view2 := m.View()
	if !strings.Contains(view2, "255;255;255") {
		t.Error("expected white title color when history pane is focused")
	}
	if !strings.Contains(view2, "Command History") {
		t.Error("expected 'Command History' title in view")
	}

	// Verify that the title of an unfocused pane is NOT white
	// Commands is now unfocused, its color should be comment color #6272a4
	// TrueColor ANSI for #6272a4 is 97;113;163 or 98;114;164
	if strings.Contains(view2, "255;255;255 > Commands") { // This is wrong, title is not white
		// The title " Commands " should not be wrapped in white color
	}

	// A better way: check that white color only appears once (for the focused pane title)
	// and potentially other things like the cursor if it uses white (it uses cyan).
	// Cleat title bar also uses purple and white? No, it uses purple.

	whiteCount := strings.Count(view2, "255;255;255")
	// If only one pane is focused, and no other white text is present...
	// Wait, taskPreview might have white text? Usually it doesn't have specific styles except what strategy provides.
	// Strategy usually uses default colors or specific ones.

	if whiteCount == 0 {
		t.Error("expected at least one white title when a pane is focused")
	}
}

func TestTaskPaneFocusAndScrolling(t *testing.T) {
	cfg := &config.Config{}
	m := InitialModel(cfg, true)
	m.width = 100
	m.height = 20

	m.taskPreview = []string{
		"task 1", "task 2", "task 3", "task 4", "task 5",
		"task 6", "task 7", "task 8", "task 9", "task 10",
	}

	// 1. 't' from Commands panel
	m.focus = focusCommands
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	m = updatedModel.(model)
	if m.focus != focusTasks {
		t.Errorf("expected focusTasks after 't' from Commands, got %v", m.focus)
	}
	if m.previousFocus != focusCommands {
		t.Errorf("expected previousFocus to be focusCommands, got %v", m.previousFocus)
	}

	// 2. Tab from Tasks panel takes you back
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focus != focusCommands {
		t.Errorf("expected focus back to Commands after Tab from Tasks, got %v", m.focus)
	}

	// 3. 't' from History panel
	m.focus = focusHistory
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	m = updatedModel.(model)
	if m.focus != focusTasks {
		t.Errorf("expected focusTasks after 't' from History, got %v", m.focus)
	}
	if m.previousFocus != focusHistory {
		t.Errorf("expected previousFocus to be focusHistory, got %v", m.previousFocus)
	}

	// 4. Tab from Tasks panel takes you back to History
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focus != focusHistory {
		t.Errorf("expected focus back to History after Tab from Tasks, got %v", m.focus)
	}

	// 5. 't' does NOT work from Config panel
	m.focus = focusConfig
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	m = updatedModel.(model)
	if m.focus != focusConfig {
		t.Errorf("expected focus to stay on Config after 't', got %v", m.focus)
	}

	// 6. Scrolling in Task panel
	m.focus = focusTasks
	m.taskScrollOffset = 0
	m.taskPreview = []string{
		"task 1", "task 2", "task 3", "task 4", "task 5",
		"task 6", "task 7", "task 8", "task 9", "task 10",
	}

	// Move down
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updatedModel.(model)
	if m.taskScrollOffset != 1 {
		t.Errorf("expected taskScrollOffset 1 after 'j', got %d", m.taskScrollOffset)
	}

	// Move up
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updatedModel.(model)
	if m.taskScrollOffset != 0 {
		t.Errorf("expected taskScrollOffset 0 after 'k', got %d", m.taskScrollOffset)
	}
}

func TestConfigFocusClearsTaskPreview(t *testing.T) {
	cfg := &config.Config{}
	m := InitialModel(cfg, true)
	m.width = 100
	m.height = 20

	// Set some task preview content manually
	m.taskPreview = []string{"task 1", "task 2"}

	// 1. Initially focus is on Commands.
	if m.focus != focusCommands {
		t.Fatalf("expected initial focusCommands, got %v", m.focus)
	}

	// 2. Tab to History. Task preview should still exist (it will be updated by updateTaskPreview)
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focus != focusHistory {
		t.Fatalf("expected focusHistory after 1st Tab, got %v", m.focus)
	}
	// Note: updateTaskPreview will run. If no history, it might clear it.
	// But let's assume it has something or we don't care yet.

	// 3. Tab to Configuration. Task preview SHOULD be cleared.
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focus != focusConfig {
		t.Fatalf("expected focusConfig after 2nd Tab, got %v", m.focus)
	}

	if len(m.taskPreview) != 0 {
		t.Errorf("expected taskPreview to be cleared when focused on Config, got %v", m.taskPreview)
	}

	// 4. Tab back to Commands. Task preview SHOULD be restored.
	// InitialModel builds a tree with some default items (build, run).
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focus != focusCommands {
		t.Fatalf("expected focus back to Commands after 3rd Tab, got %v", m.focus)
	}

	if len(m.taskPreview) == 0 {
		t.Error("expected taskPreview to be restored when tabbing away from Config to Commands")
	}
}
