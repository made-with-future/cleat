package ui

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/history"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/madewithfuture/cleat/internal/task"
)

// Update handles all keyboard and window events
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.state == stateConfirmClearHistory {
		return m.handleConfirmClearHistory(msg)
	}

	if m.state == stateInputCollection {
		return m.handleInputCollection(msg)
	}

	if m.state == stateWorkflowNameInput {
		return m.handleWorkflowNameInput(msg)
	}

	if m.state == stateWorkflowLocationSelection {
		return m.handleWorkflowLocationSelection(msg)
	}

	if m.state == stateCreatingWorkflow {
		return m.handleCreatingWorkflow(msg)
	}

	switch msg := msg.(type) {
	case editorFinishedMsg:
		return m.handleEditorFinished(msg)
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTaskPreview()
	}
	return m, nil
}

func (m model) handleConfirmClearHistory(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "enter":
			history.Clear()
			m.history, _ = history.Load()
			m.historyCursor = 0
			m.historyOffset = 0
			m.state = stateBrowsing
			m.updateTaskPreview()
			return m, nil
		case "n", "esc":
			m.state = stateBrowsing
			return m, nil
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) handleInputCollection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.collectedInputs[m.requirements[m.requirementIdx].Key] = m.textInput.Value()
			m.requirementIdx++
			if m.requirementIdx >= len(m.requirements) {
				m.quitting = true
				return m, tea.Quit
			}
			m.textInput.Prompt = m.requirements[m.requirementIdx].Prompt + ": "
			m.textInput.SetValue(m.requirements[m.requirementIdx].Default)
			m.textInput.CursorEnd()
			return m, nil
		case tea.KeyEsc:
			m.state = stateBrowsing
			return m, nil
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) handleWorkflowNameInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			name := m.textInput.Value()
			if name != "" {
				m.state = stateWorkflowLocationSelection
				m.workflowLocationIdx = 0 // 0: Project, 1: User
				return m, nil
			}
			m.state = stateBrowsing
			m.selectedWorkflowIndices = []int{}
			return m, nil
		case tea.KeyEsc:
			m.state = stateBrowsing
			return m, nil
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) handleWorkflowLocationSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			name := m.textInput.Value()
			var workflowSteps []string
			for _, idx := range m.selectedWorkflowIndices {
				if idx < len(m.history) {
					entry := m.history[idx]
					workflowSteps = append(workflowSteps, entry.Command)
				}
			}

			if len(workflowSteps) > 0 {
				workflow := config.Workflow{
					Name:     name,
					Commands: workflowSteps,
				}
				if m.workflowLocationIdx == 0 {
					history.SaveWorkflowToProject(workflow)
				} else {
					history.SaveWorkflowToUser(workflow)
				}
				m.workflows, _ = history.LoadWorkflows(m.cfg)
				m.tree = buildCommandTree(m.cfg, m.workflows)
				m.updateVisibleItems()
			}
			m.state = stateBrowsing
			m.selectedWorkflowIndices = []int{}
			return m, nil
		case tea.KeyEsc:
			m.state = stateBrowsing
			m.selectedWorkflowIndices = []int{}
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.workflowLocationIdx > 0 {
				m.workflowLocationIdx--
			}
			return m, nil
		case "down", "j":
			if m.workflowLocationIdx < 1 {
				m.workflowLocationIdx++
			}
			return m, nil
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) handleCreatingWorkflow(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			if len(m.history) > 0 {
				foundIdx := -1
				for i, idx := range m.selectedWorkflowIndices {
					if idx == m.historyCursor {
						foundIdx = i
						break
					}
				}

				if foundIdx != -1 {
					// Remove from selection
					m.selectedWorkflowIndices = append(m.selectedWorkflowIndices[:foundIdx], m.selectedWorkflowIndices[foundIdx+1:]...)
				} else {
					// Add to selection
					m.selectedWorkflowIndices = append(m.selectedWorkflowIndices, m.historyCursor)
				}
			}
			return m, nil
		case "c":
			if len(m.selectedWorkflowIndices) > 0 {
				m.state = stateWorkflowNameInput
				m.textInput.Prompt = "Workflow Name: "
				m.textInput.SetValue("")
				m.textInput.Focus()
				return m, nil
			}
		case "esc":
			m.state = stateBrowsing
			m.selectedWorkflowIndices = []int{}
			return m, nil
		case "up", "k":
			m.handleUpKey()
			return m, nil
		case "down", "j":
			m.handleDownKey()
			return m, nil
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) handleEditorFinished(msg editorFinishedMsg) (tea.Model, tea.Cmd) {
	// Reload config after editor closes
	cfg, err := config.LoadConfig(m.cfg.SourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			m.cfg = &config.Config{}
			m.cfgFound = false
		}
		// If other error, keep existing config
	} else {
		m.cfg = cfg
		m.cfgFound = true
		// Rebuild commands tree with new npm scripts and workflows
		m.workflows, _ = history.LoadWorkflows(cfg)
		m.tree = buildCommandTree(cfg, m.workflows)
		m.updateVisibleItems()
		m.updateTaskPreview()
		m.configScrollOffset = 0
	}
	return m, nil
}

func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type != tea.KeyRunes || string(msg.Runes) != "g" {
		m.pendingG = false
	}

	if m.filtering {
		return m.handleFilteringKeys(msg)
	}

	// If help is showing, any key dismisses it
	if m.showHelp {
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		default:
			m.showHelp = false
			return m, nil
		}
	}

	switch msg.Type {
	case tea.KeyCtrlC:
		m.quitting = true
		return m, tea.Quit
	case tea.KeyEsc:
		m.quitting = true
		return m, tea.Quit
	case tea.KeyTab:
		if m.focus == focusTasks {
			m.focus = m.previousFocus
		} else {
			m.focus = (m.focus + 1) % 3
		}
		m.updateTaskPreview()
	case tea.KeyShiftTab:
		if m.focus == focusTasks {
			m.focus = m.previousFocus
		} else {
			m.focus = (m.focus - 1 + 3) % 3
		}
		m.updateTaskPreview()
	case tea.KeyUp:
		m.handleUpKey()
	case tea.KeyDown:
		m.handleDownKey()
	case tea.KeyEnter:
		return m.handleEnterKey()
	case tea.KeyRunes:
		return m.handleRuneKeys(msg)
	}
	return m, nil
}

func (m model) handleFilteringKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.filtering = false
		m.filterText = ""
		m.updateVisibleItems()
		m.cursor = 0
		m.scrollOffset = 0
		m.updateTaskPreview()
		return m, nil
	case tea.KeyEnter:
		m.filtering = false
		// Continue to handle the Enter key for selection
		return m.handleEnterKey()
	case tea.KeyBackspace:
		if len(m.filterText) > 0 {
			m.filterText = m.filterText[:len(m.filterText)-1]
			m.updateVisibleItems()
			m.cursor = 0
			m.scrollOffset = 0
			m.updateTaskPreview()
		} else {
			m.filtering = false
			m.updateVisibleItems()
			m.updateTaskPreview()
		}
		return m, nil
	case tea.KeyRunes:
		m.filterText += string(msg.Runes)
		m.updateVisibleItems()
		m.cursor = 0
		m.scrollOffset = 0
		m.updateTaskPreview()
		return m, nil
	}
	return m, nil
}

func (m *model) handleUpKey() {
	if m.focus == focusCommands && m.cursor > 0 {
		m.cursor--
		if m.cursor < m.scrollOffset {
			m.scrollOffset = m.cursor
		}
		m.updateTaskPreview()
	} else if m.focus == focusHistory && m.historyCursor > 0 {
		m.historyCursor--
		if m.historyCursor < m.historyOffset {
			m.historyOffset = m.historyCursor
		}
		m.updateTaskPreview()
	} else if m.focus == focusConfig && m.configScrollOffset > 0 {
		m.configScrollOffset--
	} else if m.focus == focusTasks && m.taskScrollOffset > 0 {
		m.taskScrollOffset--
	}
}

func (m *model) handleDownKey() {
	if m.focus == focusCommands && m.cursor < len(m.visibleItems)-1 {
		m.cursor++
		visibleCount := m.visibleCommandCount()
		if m.cursor >= m.scrollOffset+visibleCount {
			m.scrollOffset = m.cursor - visibleCount + 1
		}
		m.updateTaskPreview()
	} else if m.focus == focusHistory && m.historyCursor < len(m.history)-1 {
		m.historyCursor++
		visibleCount := m.visibleHistoryCount()
		if m.historyCursor >= m.historyOffset+visibleCount {
			m.historyOffset = m.historyCursor - visibleCount + 1
		}
		m.updateTaskPreview()
	} else if m.focus == focusConfig {
		lines := m.buildConfigLines()
		visibleCount := m.visibleConfigCount()
		if m.configScrollOffset < len(lines)-visibleCount {
			m.configScrollOffset++
		}
	} else if m.focus == focusTasks {
		visibleCount := m.visibleTasksCount()
		if m.taskScrollOffset < len(m.taskPreview)-visibleCount {
			m.taskScrollOffset++
		}
	}
}

func (m model) handleEnterKey() (tea.Model, tea.Cmd) {
	if m.focus == focusCommands && len(m.visibleItems) > 0 {
		item := m.visibleItems[m.cursor]
		if len(item.item.Children) > 0 {
			item.item.Expanded = !item.item.Expanded
			m.updateVisibleItems()
			if m.cursor >= len(m.visibleItems) {
				m.cursor = len(m.visibleItems) - 1
			}
			m.updateTaskPreview()
		} else {
			m.selectedCommand = item.item.Command
			s := strategy.GetStrategyForCommand(m.selectedCommand, m.cfg)
			if s != nil {
				plan, _ := s.ResolveTasks(m.cfg)
				var reqs []task.InputRequirement
				seen := make(map[string]bool)
				for _, t := range plan {
					for _, r := range t.Requirements(m.cfg) {
						if !seen[r.Key] {
							reqs = append(reqs, r)
							seen[r.Key] = true
						}
					}
				}
				if len(reqs) > 0 {
					m.state = stateInputCollection
					m.requirements = reqs
					m.requirementIdx = 0
					m.textInput.Prompt = reqs[0].Prompt + ": "
					m.textInput.SetValue(reqs[0].Default)
					m.textInput.CursorEnd()
					return m, nil
				}
			}
			m.quitting = true
			return m, tea.Quit
		}
	} else if m.focus == focusHistory && len(m.history) > 0 {
		entry := m.history[m.historyCursor]
		m.selectedCommand = entry.Command
		m.collectedInputs = make(map[string]string)
		for k, v := range entry.Inputs {
			m.collectedInputs[k] = v
		}
		m.quitting = true
		return m, tea.Quit
	} else if m.focus == focusConfig {
		return m, m.openEditor()
	}
	return m, nil
}

func (m model) handleRuneKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(msg.Runes) == 1 {
		r := msg.Runes[0]
		if r >= '1' && r <= '9' && len(m.history) > 0 {
			target := int(r - '1')
			if target < len(m.history) {
				m.focus = focusHistory
				m.historyCursor = target
				visibleCount := m.visibleHistoryCount()
				if m.historyCursor < m.historyOffset {
					m.historyOffset = m.historyCursor
				} else if m.historyCursor >= m.historyOffset+visibleCount {
					m.historyOffset = m.historyCursor - visibleCount + 1
				}
				m.updateTaskPreview()
			}
			return m, nil
		}
	}

	switch string(msg.Runes) {
	case "q":
		m.quitting = true
		return m, tea.Quit
	case "/":
		m.filtering = true
		m.filterText = ""
		m.updateVisibleItems()
		m.cursor = 0
		m.scrollOffset = 0
		m.updateTaskPreview()
		return m, nil
	case "?":
		m.showHelp = true
	case "x":
		if m.focus == focusHistory && len(m.history) > 0 {
			m.state = stateConfirmClearHistory
			return m, nil
		}
	case "w":
		if m.focus == focusHistory && len(m.history) > 0 {
			m.state = stateCreatingWorkflow
			m.selectedWorkflowIndices = []int{}
			return m, nil
		}
	case "e":
		if m.focus == focusCommands {
			m.expandAll()
			m.updateVisibleItems()
			m.updateTaskPreview()
			return m, nil
		}
	case "c":
		if m.focus == focusCommands {
			m.collapseAll()
			m.updateVisibleItems()
			m.cursor = 0
			m.scrollOffset = 0
			m.updateTaskPreview()
			return m, nil
		}
	case "t":
		if m.focus == focusCommands || m.focus == focusHistory {
			m.previousFocus = m.focus
			m.focus = focusTasks
			m.taskScrollOffset = 0
			return m, nil
		}
	case "g":
		if m.pendingG {
			if m.focus == focusCommands {
				m.cursor = 0
				m.scrollOffset = 0
			} else if m.focus == focusHistory {
				m.historyCursor = 0
				m.historyOffset = 0
			} else if m.focus == focusConfig {
				m.configScrollOffset = 0
			}
			m.updateTaskPreview()
			m.pendingG = false
		} else {
			m.pendingG = true
		}
	case "k":
		m.handleUpKey()
	case "j":
		m.handleDownKey()
	}
	return m, nil
}

// openEditor creates config if needed and opens it in $EDITOR
func (m model) openEditor() tea.Cmd {
	// Create default config if it doesn't exist
	if !m.cfgFound {
		if err := os.WriteFile(m.cfg.SourcePath, []byte(defaultConfigTemplate), 0644); err != nil {
			// Log the error but continue - the editor will show file creation error
			fmt.Fprintf(os.Stderr, "Warning: failed to create default config: %v\n", err)
		}
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // Fallback
	}

	c := exec.Command(editor, m.cfg.SourcePath)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}
