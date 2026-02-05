package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/history"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/madewithfuture/cleat/internal/task"
)

type focus int

const (
	focusCommands focus = iota
	focusHistory
	focusConfig
	focusTasks
)

type uiState int

const (
	stateBrowsing uiState = iota
	stateInputCollection
	stateConfirmClearHistory
	stateCreatingWorkflow
	stateWorkflowNameInput
	stateWorkflowLocationSelection
	stateShowingConfig
)

// CommandItem represents a node in the command tree
type CommandItem struct {
	Label    string
	Command  string
	Children []CommandItem
	Expanded bool
}

type visibleItem struct {
	item  *CommandItem
	level int
	path  string
}

// editorFinishedMsg is sent when the editor process exits
type editorFinishedMsg struct{ err error }

// model holds all the TUI state
type model struct {
	cfg                     *config.Config
	cfgFound                bool
	quitting                bool
	width                   int
	height                  int
	tree                    []CommandItem
	visibleItems            []visibleItem
	cursor                  int
	scrollOffset            int
	configScrollOffset      int
	focus                   focus
	selectedCommand         string
	collectedInputs         map[string]string
	taskPreview             []string
	taskScrollOffset        int
	history                 []history.HistoryEntry
	historyCursor           int
	historyOffset           int
	previousFocus           focus
	workflows               []config.Workflow
	selectedWorkflowIndices []int
	showHelp                bool
	filtering               bool
	filterText              string
	state                   uiState
	requirements            []task.InputRequirement
	requirementIdx          int
	textInput               textinput.Model
	pendingG                bool
	workflowLocationIdx     int
	version                 string
}

// InitialModel creates a new model with the given config
func InitialModel(cfg *config.Config, cfgFound bool, version string) model {
	ti := textinput.New()
	ti.Focus()

	m := model{
		cfg:                     cfg,
		cfgFound:                cfgFound,
		version:                 version,
		focus:                   focusCommands,
		state:                   stateBrowsing,
		collectedInputs:         make(map[string]string),
		selectedWorkflowIndices: []int{},
		textInput:               ti,
	}
	m.history, _ = history.Load()
	m.workflows, _ = history.LoadWorkflows(cfg)
	m.tree = buildCommandTree(cfg, m.workflows)
	m.updateVisibleItems()
	m.updateTaskPreview()
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

// updateTaskPreview generates the task preview for the currently selected command
func (m *model) updateTaskPreview() {
	if m.focus == focusTasks {
		return
	}
	m.taskScrollOffset = 0
	if m.focus == focusConfig {
		m.taskPreview = nil
		return
	}
	var command string
	var inputs map[string]string

	if m.focus == focusHistory {
		if len(m.history) == 0 {
			m.taskPreview = nil
			return
		}
		entry := m.history[m.historyCursor]
		command = entry.Command
		inputs = entry.Inputs
	} else {
		if len(m.visibleItems) == 0 {
			m.taskPreview = nil
			return
		}

		item := m.visibleItems[m.cursor]
		if item.item.Command == "" {
			m.taskPreview = []string{"(expand to see commands)"}
			return
		}
		command = item.item.Command
	}

	type taskWithConfig struct {
		t       task.Task
		cfg     *config.Config
		cmdName string
	}
	var tasksToPreview []taskWithConfig

	if strings.HasPrefix(command, "workflow:") {
		name := strings.TrimPrefix(command, "workflow:")
		var workflow *config.Workflow
		for i := range m.workflows {
			if m.workflows[i].Name == name {
				workflow = &m.workflows[i]
				break
			}
		}

		if workflow != nil {
			for _, workflowCmd := range workflow.Commands {
				cfgForCmd := m.cfg
				tasks, err := strategy.ResolveCommandTasks(workflowCmd, cfgForCmd)
				if err == nil {
					for _, t := range tasks {
						tasksToPreview = append(tasksToPreview, taskWithConfig{t, cfgForCmd, workflowCmd})
					}
				}
			}
		}
	} else {
		// Use saved inputs for history items if available
		cfg := m.cfg
		if len(inputs) > 0 {
			// Create a temporary config with the saved inputs merged in
			tempCfg := *m.cfg
			tempCfg.Inputs = make(map[string]string)
			for k, v := range m.cfg.Inputs {
				tempCfg.Inputs[k] = v
			}
			for k, v := range inputs {
				tempCfg.Inputs[k] = v
			}
			cfg = &tempCfg
		}

		tasks, err := strategy.ResolveCommandTasks(command, cfg)
		if err != nil {
			m.taskPreview = []string{fmt.Sprintf("Error: %v", err)}
			return
		}
		for _, t := range tasks {
			tasksToPreview = append(tasksToPreview, taskWithConfig{t, cfg, ""})
		}
	}

	if len(tasksToPreview) == 0 {
		m.taskPreview = []string{"No tasks will run"}
		return
	}

	var preview []string
	commentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6272a4"))

	// Calculate available width for wrapping
	availableWidth := 0
	if m.width > 0 {
		availableWidth = m.width - 2 - 2 - 2 // -2 for borders, -2 for main UI padding, -2 for content padding
		if availableWidth < 0 {
			availableWidth = 0
		}
	}

	lastCmd := ""
	orange := lipgloss.Color("#ffb86c")

	for _, tc := range tasksToPreview {
		taskWidth := availableWidth
		indent := ""
		if tc.cmdName != "" {
			indent = "  "
			taskWidth -= 2
			if taskWidth < 0 {
				taskWidth = 0
			}
		}

		// Command header for workflows
		if tc.cmdName != "" && tc.cmdName != lastCmd {
			if len(preview) > 0 {
				preview = append(preview, "")
			}
			preview = append(preview, lipgloss.NewStyle().Foreground(orange).Render("→ "+tc.cmdName))
			lastCmd = tc.cmdName
		}

		// Name
		nameLines := wrapLines(strings.Fields(tc.t.Name()), taskWidth, indent+"• ", indent+"  ", lipgloss.NewStyle())
		preview = append(preview, nameLines...)

		// Description
		if tc.t.Description() != "" {
			descLines := wrapLines(strings.Fields(tc.t.Description()), taskWidth, indent+"  ", indent+"  ", lipgloss.NewStyle())
			preview = append(preview, descLines...)
		}

		// Commands
		for _, cmd := range tc.t.Commands(tc.cfg) {
			cmdLines := wrapLines(cmd, taskWidth, indent+"    $ ", indent+"        ", commentStyle)
			preview = append(preview, cmdLines...)
		}
	}
	m.taskPreview = preview
}

// paneWidths calculates the width of left and right panes
func (m model) paneWidths() (int, int) {
	const gap = 2
	available := m.width - gap
	if available < 0 {
		return 0, 0
	}
	if m.width > 110 && m.width < 150 {
		left := (available * 40) / 100
		return left, available - left
	}
	if m.width >= 150 {
		left := (available * 35) / 100
		return left, available - left
	}
	left := available / 2
	return left, available - left
}

// updateVisibleItems rebuilds the visible items list based on current filter and expansion state
func (m *model) updateVisibleItems() {
	m.visibleItems = []visibleItem{}
	if m.filterText != "" {
		for i := range m.tree {
			m.flattenFiltered(&m.tree[i], 0, "")
		}
	} else {
		for i := range m.tree {
			m.flatten(&m.tree[i], 0, "")
		}
	}
}

func (m *model) flatten(item *CommandItem, level int, parentPath string) {
	path := item.Label
	if parentPath != "" {
		path = parentPath + "." + item.Label
	}
	m.visibleItems = append(m.visibleItems, visibleItem{item: item, level: level, path: path})
	if item.Expanded && len(item.Children) > 0 {
		for i := range item.Children {
			m.flatten(&item.Children[i], level+1, path)
		}
	}
}

func (m *model) flattenFiltered(item *CommandItem, level int, parentPath string) {
	path := item.Label
	if parentPath != "" {
		path = parentPath + "." + item.Label
	}

	selfMatches := matches(item, m.filterText)
	descendantMatches := anyDescendantMatches(item, m.filterText)

	if selfMatches || descendantMatches {
		m.visibleItems = append(m.visibleItems, visibleItem{item: item, level: level, path: path})
		for i := range item.Children {
			child := &item.Children[i]
			if selfMatches || matches(child, m.filterText) || anyDescendantMatches(child, m.filterText) {
				m.flattenFiltered(child, level+1, path)
			}
		}
	}
}

// expandAll expands all items in the tree
func (m *model) expandAll() {
	for i := range m.tree {
		m.tree[i].setExpandedRecursive(true)
	}
}

// collapseAll collapses all items in the tree
func (m *model) collapseAll() {
	for i := range m.tree {
		m.tree[i].setExpandedRecursive(false)
	}
}

func (item *CommandItem) setExpandedRecursive(expanded bool) {
	if len(item.Children) > 0 {
		item.Expanded = expanded
		for i := range item.Children {
			item.Children[i].setExpandedRecursive(expanded)
		}
	}
}

// visibleCommandCount returns how many commands can fit in the pane
func (m model) visibleCommandCount() int {
	if m.height == 0 {
		return len(m.visibleItems)
	}
	titleLines := 2
	helpLines := 3
	paneHeight := (m.height - helpLines - titleLines) / 2
	// Subtract: 2 for borders, 0 for title (now on border), 0 for blank line (now removed), 1 for potential scroll indicator
	availableLines := paneHeight - 2 - 0 - 0 - 1
	if availableLines < 1 {
		availableLines = 1
	}
	return availableLines
}

func (m model) visibleConfigCount() int {
	if m.height == 0 {
		return 0
	}
	// Modal takes most of the height
	modalHeight := m.height - 10
	if modalHeight < 5 {
		modalHeight = 5
	}

	// Subtract: 2 for borders, 0 for title (now on border), 0 for blank line (removed), 1 for action hint, 1 for potential scroll indicator
	availableLines := modalHeight - 2 - 0 - 0 - 1 - 1
	if availableLines < 1 {
		availableLines = 1
	}
	return availableLines
}

func (m model) visibleHistoryCount() int {
	if m.height == 0 {
		return len(m.history)
	}
	titleLines := 2
	helpLines := 3
	paneHeight := (m.height - helpLines - titleLines) / 2
	// Subtract: 2 for borders, 1 for potential scroll indicator, 1 for padding/alignment consistency
	availableLines := paneHeight - 2 - 1 - 1
	if availableLines < 1 {
		availableLines = 1
	}
	return availableLines
}

func (m model) visibleTasksCount() int {
	if m.height == 0 {
		return len(m.taskPreview)
	}
	titleLines := 2
	helpLines := 3
	paneHeight := (m.height - helpLines - titleLines) - ((m.height - helpLines - titleLines) / 2)
	// Subtract: 2 for borders, 1 for potential scroll indicator
	availableLines := paneHeight - 2 - 1
	if availableLines < 1 {
		availableLines = 1
	}
	return availableLines
}
