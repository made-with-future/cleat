package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/history"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/madewithfuture/cleat/internal/task"
)

const defaultConfigTemplate = `# Cleat configuration
# See https://github.com/madewithfuture/cleat for documentation

version: 1
docker: true
services:
  - name: backend
    dir: .
    modules:
      - python:
          django: true
          django_service: backend
  - name: frontend
    dir: ./frontend
    modules:
      - npm:
          service: backend-node
          scripts:
            - build
`

type focus int

const (
	focusCommands focus = iota
	focusHistory
	focusConfig
)

type uiState int

const (
	stateBrowsing uiState = iota
	stateInputCollection
	stateConfirmClearHistory
)

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

type model struct {
	cfg                *config.Config
	cfgFound           bool
	quitting           bool
	width              int
	height             int
	tree               []CommandItem
	visibleItems       []visibleItem
	cursor             int
	scrollOffset       int
	configScrollOffset int
	focus              focus
	selectedCommand    string
	collectedInputs    map[string]string
	taskPreview        []string
	history            []history.HistoryEntry
	historyCursor      int
	historyOffset      int
	showHelp           bool
	filtering          bool
	filterText         string
	state              uiState
	requirements       []task.InputRequirement
	requirementIdx     int
	textInput          textinput.Model
	pendingG           bool
}

func InitialModel(cfg *config.Config, cfgFound bool) model {
	ti := textinput.New()
	ti.Focus()

	m := model{
		cfg:             cfg,
		cfgFound:        cfgFound,
		tree:            buildCommandTree(cfg),
		focus:           focusCommands,
		state:           stateBrowsing,
		collectedInputs: make(map[string]string),
		textInput:       ti,
	}
	m.history, _ = history.Load()
	m.updateVisibleItems()
	m.updateTaskPreview()
	return m
}

func (m *model) updateTaskPreview() {
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

	if len(tasks) == 0 {
		m.taskPreview = []string{"No tasks will run"}
		return
	}

	var preview []string
	commentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6272a4"))

	// Calculate available width for wrapping
	availableWidth := 0
	if m.width > 0 {
		_, rightPaneWidth := m.paneWidths()
		availableWidth = rightPaneWidth - 3 - 2 // -3 for borders/padding, -2 for right padding
	}

	for _, t := range tasks {
		// Name
		nameLines := wrapLines(strings.Fields(t.Name()), availableWidth, "• ", "  ", lipgloss.NewStyle())
		preview = append(preview, nameLines...)

		// Description
		if t.Description() != "" {
			descLines := wrapLines(strings.Fields(t.Description()), availableWidth, "  ", "  ", lipgloss.NewStyle())
			preview = append(preview, descLines...)
		}

		// Commands
		for _, cmd := range t.Commands(cfg) {
			cmdLines := wrapLines(cmd, availableWidth, "    $ ", "        ", commentStyle)
			preview = append(preview, cmdLines...)
		}
	}
	m.taskPreview = preview
}

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

// wrapLines wraps a slice of strings (e.g. command arguments or words) to fit within width
func wrapLines(args []string, width int, firstPrefix, restPrefix string, style lipgloss.Style) []string {
	if width <= 0 || width <= len(firstPrefix) || width <= len(restPrefix) {
		// No wrapping if width is too small or unknown
		return []string{style.Render(firstPrefix + strings.Join(args, " "))}
	}

	var lines []string
	var currentLine strings.Builder

	prefix := firstPrefix
	currentWidth := len(prefix)
	currentLine.WriteString(prefix)

	for _, arg := range args {
		argWidth := lipgloss.Width(arg)
		space := 0
		if currentLine.Len() > len(prefix) {
			space = 1
		}

		if currentWidth+space+argWidth > width {
			if currentLine.Len() > len(prefix) {
				// Current arg doesn't fit, finish line and start new one
				lines = append(lines, style.Render(currentLine.String()))
				prefix = restPrefix
				currentLine.Reset()
				currentLine.WriteString(prefix)
				currentLine.WriteString(arg)
				currentWidth = len(prefix) + argWidth
			} else {
				// Single arg is already too wide, just add it and it will be truncated later
				currentLine.WriteString(arg)
				lines = append(lines, style.Render(currentLine.String()))
				prefix = restPrefix
				currentLine.Reset()
				currentLine.WriteString(prefix)
				currentWidth = len(prefix)
			}
		} else {
			if space > 0 {
				currentLine.WriteString(" ")
				currentWidth += 1
			}
			currentLine.WriteString(arg)
			currentWidth += argWidth
		}
	}

	if currentLine.Len() > len(prefix) || (len(lines) == 0 && currentLine.Len() > 0) {
		lines = append(lines, style.Render(currentLine.String()))
	}

	return lines
}

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

func matches(item *CommandItem, text string) bool {
	if text == "" {
		return true
	}
	text = strings.ToLower(text)
	return strings.Contains(strings.ToLower(item.Label), text) ||
		strings.Contains(strings.ToLower(item.Command), text)
}

func anyDescendantMatches(item *CommandItem, text string) bool {
	for i := range item.Children {
		if matches(&item.Children[i], text) || anyDescendantMatches(&item.Children[i], text) {
			return true
		}
	}
	return false
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

func (m *model) expandAll() {
	for i := range m.tree {
		m.tree[i].setExpandedRecursive(true)
	}
}

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

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.state == stateConfirmClearHistory {
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

	if m.state == stateInputCollection {
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

	switch msg := msg.(type) {
	case editorFinishedMsg:
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
			// Rebuild commands tree with new npm scripts
			m.tree = buildCommandTree(cfg)
			m.updateVisibleItems()
			m.updateTaskPreview()
			m.configScrollOffset = 0
		}
		return m, nil

	case tea.KeyMsg:
		if msg.Type != tea.KeyRunes || string(msg.Runes) != "g" {
			m.pendingG = false
		}
		if m.filtering {
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
			}
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
			m.focus = (m.focus + 1) % 3
			m.updateTaskPreview()
		case tea.KeyShiftTab:
			m.focus = (m.focus - 1 + 3) % 3
			m.updateTaskPreview()
		case tea.KeyUp:
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
			}
		case tea.KeyDown:
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
			}
		case tea.KeyEnter:
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
		case tea.KeyRunes:
			if m.filtering {
				m.filterText += string(msg.Runes)
				m.updateVisibleItems()
				m.cursor = 0
				m.scrollOffset = 0
				m.updateTaskPreview()
				return m, nil
			}

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
				}
			case "j":
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
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTaskPreview()
	}
	return m, nil
}

// openEditor creates config if needed and opens it in $EDITOR
func (m model) openEditor() tea.Cmd {
	// Create default config if it doesn't exist
	if !m.cfgFound {
		if err := os.WriteFile(m.cfg.SourcePath, []byte(defaultConfigTemplate), 0644); err != nil {
			// If we can't write, just try to open anyway
			_ = err
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

// buildCommandTree creates the commands tree from config
func buildCommandTree(cfg *config.Config) []CommandItem {
	var tree []CommandItem
	tree = append(tree, CommandItem{Label: "build", Command: "build"})
	tree = append(tree, CommandItem{Label: "run", Command: "run"})

	if cfg.Docker {
		tree = append(tree, CommandItem{
			Label: "docker",
			Children: []CommandItem{
				{Label: "down", Command: "docker down"},
				{Label: "rebuild", Command: "docker rebuild"},
				{Label: "remove-orphans", Command: "docker remove-orphans"},
			},
		})
	}

	if cfg.GoogleCloudPlatform != nil {
		gcpChildren := []CommandItem{
			{Label: "activate", Command: "gcp activate"},
			{Label: "adc-login", Command: "gcp adc-login"},
			{Label: "init", Command: "gcp init"},
			{Label: "set-config", Command: "gcp set-config"},
		}
		if cfg.AppYaml != "" {
			gcpChildren = append(gcpChildren, CommandItem{Label: "deploy", Command: "gcp app-engine deploy"})
			gcpChildren = append(gcpChildren, CommandItem{Label: "promote", Command: "gcp app-engine promote"})
		}
		gcpChildren = append(gcpChildren, CommandItem{Label: "console", Command: "gcp console"})
		tree = append(tree, CommandItem{
			Label:    "gcp",
			Children: gcpChildren,
		})
	}

	if cfg.Terraform != nil {
		if cfg.Terraform.UseFolders && len(cfg.Terraform.Envs) > 0 {
			var tfChildren []CommandItem
			for _, env := range cfg.Terraform.Envs {
				tfChildren = append(tfChildren, CommandItem{
					Label: env,
					Children: []CommandItem{
						{Label: "init", Command: "terraform init:" + env},
						{Label: "init-upgrade", Command: "terraform init-upgrade:" + env},
						{Label: "plan", Command: "terraform plan:" + env},
						{Label: "apply", Command: "terraform apply:" + env},
						{Label: "apply-refresh", Command: "terraform apply-refresh:" + env},
					},
				})
			}
			tree = append(tree, CommandItem{
				Label:    "terraform",
				Children: tfChildren,
			})
		} else {
			tree = append(tree, CommandItem{
				Label: "terraform",
				Children: []CommandItem{
					{Label: "init", Command: "terraform init"},
					{Label: "init-upgrade", Command: "terraform init-upgrade"},
					{Label: "plan", Command: "terraform plan"},
					{Label: "apply", Command: "terraform apply"},
					{Label: "apply-refresh", Command: "terraform apply-refresh"},
				},
			})
		}
	}

	for i := range cfg.Services {
		svc := &cfg.Services[i]
		svcItem := CommandItem{
			Label: svc.Name,
		}

		for j := range svc.Modules {
			mod := &svc.Modules[j]

			// Python/Django
			if mod.Python != nil && mod.Python.Django {
				var djangoChildren []CommandItem
				if cfg.Docker {
					djangoChildren = append(djangoChildren, CommandItem{Label: "create-user-dev", Command: fmt.Sprintf("django create-user-dev:%s", svc.Name)})
				}
				djangoChildren = append(djangoChildren, CommandItem{Label: "collectstatic", Command: fmt.Sprintf("django collectstatic:%s", svc.Name)})
				djangoChildren = append(djangoChildren, CommandItem{Label: "makemigrations", Command: fmt.Sprintf("django makemigrations:%s", svc.Name)})
				djangoChildren = append(djangoChildren, CommandItem{Label: "migrate", Command: fmt.Sprintf("django migrate:%s", svc.Name)})
				djangoChildren = append(djangoChildren, CommandItem{Label: "gen-random-secret-key", Command: fmt.Sprintf("django gen-random-secret-key:%s", svc.Name)})

				svcItem.Children = append(svcItem.Children, CommandItem{
					Label:    "django",
					Children: djangoChildren,
				})
			}

			// NPM
			if mod.Npm != nil {
				npmItem := CommandItem{
					Label: "npm",
				}
				npmItem.Children = append(npmItem.Children, CommandItem{
					Label:   "install",
					Command: fmt.Sprintf("npm install:%s", svc.Name),
				})
				for _, script := range mod.Npm.Scripts {
					npmItem.Children = append(npmItem.Children, CommandItem{
						Label:   fmt.Sprintf("run %s", script),
						Command: fmt.Sprintf("npm run %s:%s", svc.Name, script),
					})
				}
				svcItem.Children = append(svcItem.Children, npmItem)
			}
		}

		if svc.AppYaml != "" {
			svcItem.Children = append(svcItem.Children, CommandItem{Label: "deploy", Command: fmt.Sprintf("gcp app-engine deploy:%s", svc.Name)})
			svcItem.Children = append(svcItem.Children, CommandItem{Label: "promote", Command: fmt.Sprintf("gcp app-engine promote:%s", svc.Name)})
		}

		if len(svcItem.Children) > 0 {
			tree = append(tree, svcItem)
		}
	}

	return tree
}

// visibleCommandCount returns how many commands can fit in the pane
func (m model) visibleCommandCount() int {
	if m.height == 0 {
		return len(m.visibleItems)
	}
	titleLines := 1
	helpLines := 2
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
	titleLines := 1
	helpLines := 2
	paneHeight := (m.height - helpLines - titleLines) - ((m.height - helpLines - titleLines) / 2)

	// Subtract: 2 for borders, 0 for title (now on border), 0 for blank line (removed), 1 for action hint, 1 for potential scroll indicator
	availableLines := paneHeight - 2 - 0 - 0 - 1 - 1
	if availableLines < 1 {
		availableLines = 1
	}
	return availableLines
}

// drawBox draws a box with rounded corners around content, with an optional title in the top border
func drawBox(lines []string, width, height int, borderColor lipgloss.Color, title string, titleFocused bool) string {
	colorStyle := lipgloss.NewStyle().Foreground(borderColor)

	innerWidth := width - 2 // Account for left and right borders

	var result strings.Builder

	// Top border
	if title != "" {
		displayTitle := " " + title + " "
		titleWidth := lipgloss.Width(displayTitle)
		if titleWidth > innerWidth-2 {
			// Truncate if too long
			displayTitle = " " + ansi.Truncate(strings.TrimSpace(title), innerWidth-4, "…") + " "
			titleWidth = lipgloss.Width(displayTitle)
		}

		var renderedTitle string
		if titleFocused {
			renderedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true).Render(displayTitle)
		} else {
			renderedTitle = colorStyle.Render(displayTitle)
		}

		dashes := innerWidth - titleWidth - 1
		if dashes < 0 {
			dashes = 0
		}
		result.WriteString(colorStyle.Render("╭─") + renderedTitle + colorStyle.Render(strings.Repeat("─", dashes)+"╮"))
	} else {
		result.WriteString(colorStyle.Render("╭" + strings.Repeat("─", innerWidth) + "╮"))
	}
	result.WriteString("\n")

	// Content lines
	contentLines := height - 2 // Account for top and bottom borders
	for i := 0; i < contentLines; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}

		// Truncate or pad line to fit (ANSI-aware)
		visibleWidth := lipgloss.Width(line)
		if visibleWidth > innerWidth {
			line = ansi.Truncate(line, innerWidth, "")
			visibleWidth = lipgloss.Width(line)
		}
		if visibleWidth < innerWidth {
			line = line + strings.Repeat(" ", innerWidth-visibleWidth)
		}

		result.WriteString(colorStyle.Render("│"))
		result.WriteString(line)
		result.WriteString(colorStyle.Render("│"))
		result.WriteString("\n")
	}

	// Bottom border
	result.WriteString(colorStyle.Render("╰" + strings.Repeat("─", innerWidth) + "╯"))

	return result.String()
}

func (m model) overlay(background, foreground string) string {
	bgLines := strings.Split(background, "\n")
	fgLines := strings.Split(foreground, "\n")

	fgWidth := lipgloss.Width(foreground)
	fgHeight := len(fgLines)

	x := (m.width - fgWidth) / 2
	y := (m.height - fgHeight) / 2

	for i := 0; i < fgHeight; i++ {
		if y+i >= 0 && y+i < len(bgLines) {
			bgLines[y+i] = m.overlayLine(bgLines[y+i], fgLines[i], x)
		}
	}

	return strings.Join(bgLines, "\n")
}

func (m model) overlayLine(bg, fg string, x int) string {
	bgWidth := lipgloss.Width(bg)
	fgWidth := lipgloss.Width(fg)

	if x < 0 {
		x = 0
	}
	if x >= bgWidth {
		// If background is shorter than x, pad it with spaces
		bg = bg + strings.Repeat(" ", x-bgWidth)
		return bg + fg
	}

	left := ansi.Truncate(bg, x, "")
	right := m.visibleTail(bg, x+fgWidth)

	return left + fg + right
}

func (m model) visibleTail(s string, skipWidth int) string {
	var currentStyle strings.Builder
	var result strings.Builder
	currW := 0
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			start := i
			i++
			if i < len(s) && s[i] == '[' { // CSI
				i++
				for i < len(s) && (s[i] >= 0x30 && s[i] <= 0x3f) {
					i++
				}
				for i < len(s) && (s[i] >= 0x20 && s[i] <= 0x2f) {
					i++
				}
				if i < len(s) && (s[i] >= 0x40 && s[i] <= 0x7e) {
					i++
				}
			}
			style := s[start:i]
			if currW < skipWidth {
				currentStyle.WriteString(style)
			} else {
				result.WriteString(style)
			}
			continue
		}

		r, width := utf8.DecodeRuneInString(s[i:])
		rw := lipgloss.Width(string(r))
		if currW >= skipWidth {
			result.WriteRune(r)
		}
		currW += rw
		i += width
	}
	return currentStyle.String() + result.String()
}

// renderHelpOverlay renders a centered help modal
func (m model) renderHelpOverlay() string {
	purple := lipgloss.Color("#bd93f9")
	comment := lipgloss.Color("#6272a4")
	fg := lipgloss.Color("#f8f8f2")

	title := lipgloss.NewStyle().Bold(true).Foreground(purple).Render("Keyboard Shortcuts")

	helpItems := []string{
		"",
		title,
		"",
		"  ↑/k        Move up",
		"  ↓/j        Move down",
		"  e          Expand all",
		"  c          Collapse all",
		"  /          Filter commands",
		"  Enter      Select/Toggle / Edit config",
		"  1-9        Jump to history item",
		"  x          Clear history (history pane)",
		"  Tab        Switch pane",
		"  Shift+Tab  Switch pane (reverse)",
		"  q/Esc      Quit",
		"  ?          Show this help",
		"",
		lipgloss.NewStyle().Foreground(comment).Render("  Press any key to close"),
		"",
	}

	content := strings.Join(helpItems, "\n")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Foreground(fg).
		Padding(0, 2)

	box := boxStyle.Render(content)
	return box
}

func (m model) buildConfigLines() []string {
	var configLines []string

	if !m.cfgFound {
		configLines = append(configLines, " "+lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")).Italic(true).Render("No cleat.yaml found"))
		configLines = append(configLines, "")
	}
	configLines = append(configLines, fmt.Sprintf(" version: %d", m.cfg.Version))
	configLines = append(configLines, fmt.Sprintf(" docker: %v", m.cfg.Docker))

	if len(m.cfg.Envs) > 0 {
		configLines = append(configLines, " envs:")
		for _, env := range m.cfg.Envs {
			configLines = append(configLines, fmt.Sprintf("   - %s", env))
		}
	}

	if m.cfg.GoogleCloudPlatform != nil {
		configLines = append(configLines, " google_cloud_platform:")
		if m.cfg.GoogleCloudPlatform.ProjectName != "" {
			configLines = append(configLines, fmt.Sprintf("   project_name: %s", m.cfg.GoogleCloudPlatform.ProjectName))
		}
	}

	if m.cfg.Terraform != nil {
		configLines = append(configLines, " terraform:")
		if len(m.cfg.Terraform.Envs) > 0 {
			configLines = append(configLines, "   envs:")
			for _, env := range m.cfg.Terraform.Envs {
				configLines = append(configLines, fmt.Sprintf("     - %s", env))
			}
		}
	}

	for i := range m.cfg.Services {
		svc := &m.cfg.Services[i]
		configLines = append(configLines, fmt.Sprintf(" service: %s", svc.Name))
		if svc.Dir != "" {
			configLines = append(configLines, fmt.Sprintf("   dir: %s", svc.Dir))
		}
		if svc.IsDocker() {
			configLines = append(configLines, fmt.Sprintf("   docker: %v", svc.IsDocker()))
		}
		for j := range svc.Modules {
			mod := &svc.Modules[j]
			if mod.Python != nil && mod.Python.Django {
				configLines = append(configLines, "   python:")
				configLines = append(configLines, fmt.Sprintf("     django: %v", mod.Python.Django))
				if mod.Python.DjangoService != "" {
					configLines = append(configLines, fmt.Sprintf("     django_service: %s", mod.Python.DjangoService))
				}
				if mod.Python.PackageManager != "" {
					configLines = append(configLines, fmt.Sprintf("     package_manager: %s", mod.Python.PackageManager))
				}
			}
			if mod.Npm != nil && len(mod.Npm.Scripts) > 0 {
				configLines = append(configLines, "   npm:")
				configLines = append(configLines, fmt.Sprintf("     service: %s", mod.Npm.Service))
			}
		}
	}

	return configLines
}

func (m model) renderInputModal() string {
	purple := lipgloss.Color("#bd93f9")
	fg := lipgloss.Color("#f8f8f2")

	title := lipgloss.NewStyle().Bold(true).Foreground(purple).Render("Input Required")

	stepInfo := fmt.Sprintf("Step %d of %d", m.requirementIdx+1, len(m.requirements))

	content := []string{
		"",
		title,
		"",
		stepInfo,
		"",
		m.textInput.View(),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6272a4")).Render("  Enter: continue • Esc: cancel"),
		"",
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Foreground(fg).
		Padding(0, 2)

	box := boxStyle.Render(strings.Join(content, "\n"))
	return box
}

func (m model) renderConfirmModal() string {
	purple := lipgloss.Color("#bd93f9")
	fg := lipgloss.Color("#f8f8f2")
	red := lipgloss.Color("#ff5555")

	title := lipgloss.NewStyle().Bold(true).Foreground(red).Render("Clear History")

	content := []string{
		title,
		"",
		lipgloss.NewStyle().Foreground(fg).Render("Are you sure you want to clear history?"),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6272a4")).Render("  y: confirm • n/Esc: cancel"),
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Foreground(fg).
		Padding(0, 2)

	box := boxStyle.Render(strings.Join(content, "\n"))
	return box
}

func (m model) visibleHistoryCount() int {
	// 2 (top/bottom borders) + 2 (more indicators) = 4
	// (Search/Padding line removed)
	// But it's split with Commands pane
	return (m.height-1-2)/2 - 3
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Minimum usable dimensions for side-by-side panes
	const minWidth = 60
	const minHeight = 20
	if m.width < minWidth || m.height < minHeight {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff5555")).
			Render(fmt.Sprintf("Terminal too small (%dx%d). Minimum: %dx%d", m.width, m.height, minWidth, minHeight))
	}

	base := m.renderMainUI()

	if m.state == stateInputCollection {
		return m.overlay(base, m.renderInputModal())
	}

	if m.state == stateConfirmClearHistory {
		return m.overlay(base, m.renderConfirmModal())
	}

	// Show help overlay if active
	if m.showHelp {
		return m.overlay(base, m.renderHelpOverlay())
	}

	return base
}

func (m model) renderMainUI() string {
	// Dracula colors
	purple := lipgloss.Color("#bd93f9")
	cyan := lipgloss.Color("#8be9fd")
	comment := lipgloss.Color("#6272a4")
	green := lipgloss.Color("#50fa7b")
	red := lipgloss.Color("#ff5555")

	// Build title bar
	title := " Cleat "
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
	borderStyle := lipgloss.NewStyle().Foreground(comment)

	titleRendered := titleStyle.Render(title)
	titleWidth := lipgloss.Width(titleRendered)
	totalPadding := m.width - titleWidth - 2 // -2 for the corner characters
	leftPadding := totalPadding / 2
	rightPadding := totalPadding - leftPadding

	titleBar := borderStyle.Render("╭") +
		borderStyle.Render(strings.Repeat("─", leftPadding)) +
		titleRendered +
		borderStyle.Render(strings.Repeat("─", rightPadding)) +
		borderStyle.Render("╮")

	// Determine border colors based on focus
	commandsColor := comment
	historyColor := comment
	configColor := comment

	switch m.focus {
	case focusCommands:
		commandsColor = purple
	case focusHistory:
		historyColor = purple
	case focusConfig:
		configColor = purple
	}

	// Calculate dimensions
	gap := 2
	titleLines := 1
	helpLines := 2
	leftPaneWidth, rightPaneWidth := m.paneWidths()
	paneHeight := m.height - helpLines - titleLines

	// Left panes height split
	commandsPaneHeight := paneHeight / 2
	historyPaneHeight := paneHeight - commandsPaneHeight

	// Right panes height split
	taskPaneHeight := paneHeight / 2
	configPaneHeight := paneHeight - taskPaneHeight

	// Build left pane content (with padding)
	var leftLines []string

	if m.filtering {
		filterStyle := lipgloss.NewStyle().Foreground(purple)
		leftLines = append(leftLines, " "+filterStyle.Render("/"+m.filterText+"█"))
	} else if m.filterText != "" {
		filterStyle := lipgloss.NewStyle().Foreground(comment)
		leftLines = append(leftLines, " "+filterStyle.Render("/"+m.filterText))
	}

	visibleCount := m.visibleCommandCount()
	hasMoreAbove := m.scrollOffset > 0
	hasMoreBelow := m.scrollOffset+visibleCount < len(m.visibleItems)

	// Show scroll up indicator
	if hasMoreAbove {
		leftLines = append(leftLines, " "+lipgloss.NewStyle().Foreground(comment).Render("▲ more"))
	}

	// Render visible commands
	endIdx := m.scrollOffset + visibleCount
	if endIdx > len(m.visibleItems) {
		endIdx = len(m.visibleItems)
	}
	for i := m.scrollOffset; i < endIdx; i++ {
		vItem := m.visibleItems[i]
		label := vItem.item.Label
		if len(vItem.item.Children) > 0 {
			marker := "▸ "
			if vItem.item.Expanded {
				marker = "▾ "
			}
			label = marker + label
		} else {
			label = "  " + label
		}

		// Indentation
		indent := strings.Repeat("  ", vItem.level)
		label = indent + label

		if i == m.cursor {
			cursorColor := cyan
			if m.focus != focusCommands {
				cursorColor = comment // Dim when not focused
			}
			leftLines = append(leftLines, " "+lipgloss.NewStyle().Foreground(cursorColor).Render("> "+label))
		} else {
			leftLines = append(leftLines, "   "+label)
		}
	}

	// Show scroll down indicator
	if hasMoreBelow {
		leftLines = append(leftLines, " "+lipgloss.NewStyle().Foreground(comment).Render("▼ more"))
	}

	// Build task preview pane content
	var taskLines []string
	for _, line := range m.taskPreview {
		taskLines = append(taskLines, " "+line)
	}

	// Build config pane content (with padding)
	var configLines []string

	allConfigLines := m.buildConfigLines()
	visibleConfigCount := m.visibleConfigCount()
	hasMoreConfigAbove := m.configScrollOffset > 0
	hasMoreConfigBelow := m.configScrollOffset+visibleConfigCount < len(allConfigLines)

	// Show scroll up indicator
	if hasMoreConfigAbove {
		configLines = append(configLines, " "+lipgloss.NewStyle().Foreground(comment).Render("▲ more"))
	}

	// Render visible config lines
	endConfigIdx := m.configScrollOffset + visibleConfigCount
	if endConfigIdx > len(allConfigLines) {
		endConfigIdx = len(allConfigLines)
	}
	for i := m.configScrollOffset; i < endConfigIdx; i++ {
		configLines = append(configLines, " "+allConfigLines[i])
	}

	// Show scroll down indicator
	if hasMoreConfigBelow {
		configLines = append(configLines, " "+lipgloss.NewStyle().Foreground(comment).Render("▼ more"))
	}

	configLines = append(configLines, "")
	// Action hint
	if m.focus == focusConfig {
		actionText := "Press Enter to edit"
		if !m.cfgFound {
			actionText = "Press Enter to create"
		}
		configLines = append(configLines, " "+lipgloss.NewStyle().Foreground(purple).Render(actionText))
	}

	// Build history pane content
	var historyLines []string
	visibleHistoryCount := m.visibleHistoryCount()
	hasMoreHistoryAbove := m.historyOffset > 0
	hasMoreHistoryBelow := m.historyOffset+visibleHistoryCount < len(m.history)

	if hasMoreHistoryAbove {
		historyLines = append(historyLines, " "+lipgloss.NewStyle().Foreground(comment).Render("▲ more"))
	}

	endHistoryIdx := m.historyOffset + visibleHistoryCount
	if endHistoryIdx > len(m.history) {
		endHistoryIdx = len(m.history)
	}
	for i := m.historyOffset; i < endHistoryIdx; i++ {
		entry := m.history[i]

		// Icon for success/failure
		icon := "✓"
		iconColor := green
		if !entry.Success {
			icon = "✘"
			iconColor = red
		}
		renderedIcon := lipgloss.NewStyle().Foreground(iconColor).Render(icon)

		// Format: Command (aligned left) ... Date Time (aligned right)
		ts := entry.Timestamp.Format("2006-01-02 15:04")
		contentWidth := rightPaneWidth - 2 - 3 - 2 - 2 // -2 for borders, -3 for prefix, -2 for right padding, -2 for icon
		if contentWidth < 0 {
			contentWidth = 0
		}

		tsWidth := lipgloss.Width(ts)
		var label string
		if contentWidth <= tsWidth {
			label = renderedIcon + " " + ansi.Truncate(ts, contentWidth, "")
		} else {
			cmdMaxWidth := contentWidth - tsWidth - 1 // at least one space
			displayCmd := entry.Command
			if lipgloss.Width(displayCmd) > cmdMaxWidth {
				displayCmd = ansi.Truncate(displayCmd, cmdMaxWidth, "…")
			}
			spaces := contentWidth - lipgloss.Width(displayCmd) - tsWidth
			if spaces < 0 {
				spaces = 0
			}
			label = renderedIcon + " " + displayCmd + strings.Repeat(" ", spaces) + ts
		}

		if i == m.historyCursor {
			cursorColor := cyan
			if m.focus != focusHistory {
				cursorColor = comment
			}
			historyLines = append(historyLines, " "+lipgloss.NewStyle().Foreground(cursorColor).Render("> "+label))
		} else {
			historyLines = append(historyLines, "   "+label)
		}
	}
	if hasMoreHistoryBelow {
		historyLines = append(historyLines, " "+lipgloss.NewStyle().Foreground(comment).Render("▼ more"))
	}

	// Draw boxes
	commandsBox := drawBox(leftLines, leftPaneWidth, commandsPaneHeight, commandsColor, "Commands", m.focus == focusCommands)
	configBox := drawBox(configLines, leftPaneWidth, configPaneHeight, configColor, "Configuration", m.focus == focusConfig)

	taskTitle := "Tasks to run"
	if m.focus == focusHistory {
		if len(m.history) > 0 {
			entry := m.history[m.historyCursor]
			// Try to find the path from command tree if possible, otherwise use command string
			taskTitle = fmt.Sprintf("Tasks for %s", entry.Command)
		}
	} else if len(m.visibleItems) > 0 {
		vItem := m.visibleItems[m.cursor]
		if vItem.item.Command != "" {
			taskTitle = fmt.Sprintf("Tasks for %s", strings.TrimSpace(vItem.path))
		}
	}
	taskBox := drawBox(taskLines, rightPaneWidth, taskPaneHeight, comment, taskTitle, false)

	historyBox := drawBox(historyLines, rightPaneWidth, historyPaneHeight, historyColor, "Command History", m.focus == focusHistory)

	// Join boxes vertically for left and right sides
	commandsBoxLines := strings.Split(commandsBox, "\n")
	configBoxLines := strings.Split(configBox, "\n")
	leftBoxLines := append(commandsBoxLines, configBoxLines...)

	taskBoxLines := strings.Split(taskBox, "\n")
	historyBoxLines := strings.Split(historyBox, "\n")
	rightBoxLines := append(taskBoxLines, historyBoxLines...)

	var combined strings.Builder
	maxLines := len(leftBoxLines)
	if len(rightBoxLines) > maxLines {
		maxLines = len(rightBoxLines)
	}

	for i := 0; i < maxLines; i++ {
		left := ""
		right := ""
		if i < len(leftBoxLines) {
			left = leftBoxLines[i]
		}
		if i < len(rightBoxLines) {
			right = rightBoxLines[i]
		}
		combined.WriteString(left)
		combined.WriteString(strings.Repeat(" ", gap))
		combined.WriteString(right)
		if i < maxLines-1 {
			combined.WriteString("\n")
		}
	}

	// Help line
	helpText := lipgloss.NewStyle().Foreground(comment).Render("↑/↓: navigate • enter: select/toggle • tab: switch pane • ?: help • q: quit")
	if !m.cfgFound {
		warning := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")).Render("(no cleat.yaml)")
		separator := lipgloss.NewStyle().Foreground(comment).Render(" • ")
		helpText = warning + separator + helpText
	}

	return titleBar + "\n" + combined.String() + "\n\n" + helpText
}

func Start() (string, map[string]string, error) {
	cfg, err := config.LoadConfig("cleat.yaml")
	cfgFound := true
	if err != nil {
		if os.IsNotExist(err) {
			cfg = &config.Config{}
			cfgFound = false
		} else {
			return "", nil, err
		}
	}

	m := InitialModel(cfg, cfgFound)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", nil, err
	}

	if fm, ok := finalModel.(model); ok {
		return fm.selectedCommand, fm.collectedInputs, nil
	}

	return "", nil, nil
}
