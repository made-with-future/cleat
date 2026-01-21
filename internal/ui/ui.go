package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/madewithfuture/cleat/internal/task"
)

const configPath = "cleat.yaml"

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
	focusConfig
)

type uiState int

const (
	stateBrowsing uiState = iota
	stateInputCollection
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
	showHelp           bool
	filtering          bool
	filterText         string
	state              uiState
	requirements       []task.InputRequirement
	requirementIdx     int
	textInput          textinput.Model
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
	m.updateVisibleItems()
	m.updateTaskPreview()
	return m
}

func (m *model) updateTaskPreview() {
	if len(m.visibleItems) == 0 {
		m.taskPreview = nil
		return
	}

	item := m.visibleItems[m.cursor]
	if item.item.Command == "" {
		m.taskPreview = []string{"(expand to see commands)"}
		return
	}

	tasks, err := strategy.ResolveCommandTasks(item.item.Command, m.cfg)
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
		gap := 2
		paneWidth := (m.width - gap) / 2
		availableWidth = paneWidth - 3
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
		for _, cmd := range t.Commands(m.cfg) {
			cmdLines := wrapLines(cmd, availableWidth, "    $ ", "        ", commentStyle)
			preview = append(preview, cmdLines...)
		}
	}
	m.taskPreview = preview
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

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		cfg, err := config.LoadConfig(configPath)
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
			m.focus = (m.focus + 1) % 2
		case tea.KeyShiftTab:
			m.focus = (m.focus - 1 + 2) % 2
		case tea.KeyUp:
			if m.focus == focusCommands && m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
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
			case "k":
				if m.focus == focusCommands && m.cursor > 0 {
					m.cursor--
					if m.cursor < m.scrollOffset {
						m.scrollOffset = m.cursor
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
		if err := os.WriteFile(configPath, []byte(defaultConfigTemplate), 0644); err != nil {
			// If we can't write, just try to open anyway
			_ = err
		}
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // Fallback
	}

	c := exec.Command(editor, configPath)
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
			Expanded: true,
		})
	}

	if cfg.GoogleCloudPlatform != nil {
		tree = append(tree, CommandItem{
			Label: "gcp",
			Children: []CommandItem{
				{Label: "activate", Command: "gcp activate"},
				{Label: "adc-login", Command: "gcp adc-login"},
				{Label: "init", Command: "gcp init"},
				{Label: "set-config", Command: "gcp set-config"},
			},
			Expanded: true,
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
					Expanded: true,
				})
			}
			tree = append(tree, CommandItem{
				Label:    "terraform",
				Children: tfChildren,
				Expanded: true,
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
				Expanded: true,
			})
		}
	}

	for i := range cfg.Services {
		svc := &cfg.Services[i]
		svcItem := CommandItem{
			Label:    svc.Name,
			Expanded: true,
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
				djangoChildren = append(djangoChildren, CommandItem{Label: "migrate", Command: fmt.Sprintf("django migrate:%s", svc.Name)})
				djangoChildren = append(djangoChildren, CommandItem{Label: "gen-random-secret-key", Command: fmt.Sprintf("django gen-random-secret-key:%s", svc.Name)})

				svcItem.Children = append(svcItem.Children, CommandItem{
					Label:    "django",
					Children: djangoChildren,
					Expanded: true,
				})
			}

			// NPM
			if mod.Npm != nil && len(mod.Npm.Scripts) > 0 {
				npmItem := CommandItem{
					Label:    "npm",
					Expanded: true,
				}
				for _, script := range mod.Npm.Scripts {
					npmItem.Children = append(npmItem.Children, CommandItem{
						Label:   fmt.Sprintf("run %s", script),
						Command: fmt.Sprintf("npm run %s:%s", svc.Name, script),
					})
				}
				svcItem.Children = append(svcItem.Children, npmItem)
			}
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
	paneHeight := m.height - helpLines - titleLines
	// Subtract: 2 for borders, 0 for title (now on border), 1 for blank line after title (or filter bar), 1 for potential scroll indicator
	availableLines := paneHeight - 2 - 0 - 1 - 1
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
	paneHeight := m.height - helpLines - titleLines
	configPaneHeight := paneHeight - (paneHeight / 2)

	// Subtract: 2 for borders, 0 for title (now on border), 1 for blank line, 1 for action hint, 1 for potential scroll indicator
	availableLines := configPaneHeight - 2 - 0 - 1 - 1 - 1
	if availableLines < 1 {
		availableLines = 1
	}
	return availableLines
}

// drawBox draws a box with rounded corners around content, with an optional title in the top border
func drawBox(lines []string, width, height int, borderColor lipgloss.Color, title string) string {
	colorStyle := lipgloss.NewStyle().Foreground(borderColor)

	innerWidth := width - 2 // Account for left and right borders

	var result strings.Builder

	// Top border
	if title != "" {
		renderedTitle := " " + title + " "
		titleWidth := lipgloss.Width(renderedTitle)
		if titleWidth > innerWidth-2 {
			// Truncate if too long
			renderedTitle = " " + ansi.Truncate(strings.TrimSpace(title), innerWidth-4, "…") + " "
			titleWidth = lipgloss.Width(renderedTitle)
		}
		dashes := innerWidth - titleWidth - 1
		if dashes < 0 {
			dashes = 0
		}
		result.WriteString(colorStyle.Render("╭─" + renderedTitle + strings.Repeat("─", dashes) + "╮"))
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
		"  /          Filter commands",
		"  Enter      Select/Toggle / Edit config",
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

	// Center the box on screen
	boxWidth := lipgloss.Width(box)
	boxHeight := lipgloss.Height(box)

	horizontalPad := (m.width - boxWidth) / 2
	verticalPad := (m.height - boxHeight) / 2

	if horizontalPad < 0 {
		horizontalPad = 0
	}
	if verticalPad < 0 {
		verticalPad = 0
	}

	// Build centered output
	var result strings.Builder
	for i := 0; i < verticalPad; i++ {
		result.WriteString("\n")
	}

	for _, line := range strings.Split(box, "\n") {
		result.WriteString(strings.Repeat(" ", horizontalPad))
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
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
		if m.cfg.Terraform.UseFolders {
			configLines = append(configLines, "   use_folders: true")
			if len(m.cfg.Terraform.Envs) > 0 {
				configLines = append(configLines, "   envs:")
				for _, env := range m.cfg.Terraform.Envs {
					configLines = append(configLines, fmt.Sprintf("     - %s", env))
				}
			}
		} else {
			configLines = append(configLines, "   use_folders: false")
		}
	}

	for i := range m.cfg.Services {
		svc := &m.cfg.Services[i]
		configLines = append(configLines, fmt.Sprintf(" service: %s", svc.Name))
		if svc.Dir != "" {
			configLines = append(configLines, fmt.Sprintf("   dir: %s", svc.Dir))
		}
		if svc.Docker {
			configLines = append(configLines, fmt.Sprintf("   docker: %v", svc.Docker))
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

	// Center the box on screen
	boxWidth := lipgloss.Width(box)
	boxHeight := lipgloss.Height(box)

	horizontalPad := (m.width - boxWidth) / 2
	verticalPad := (m.height - boxHeight) / 2

	if horizontalPad < 0 {
		horizontalPad = 0
	}
	if verticalPad < 0 {
		verticalPad = 0
	}

	var result strings.Builder
	for i := 0; i < verticalPad; i++ {
		result.WriteString("\n")
	}

	for _, line := range strings.Split(box, "\n") {
		result.WriteString(strings.Repeat(" ", horizontalPad))
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	if m.state == stateInputCollection {
		return m.renderInputModal()
	}

	// Show help overlay if active
	if m.showHelp {
		return m.renderHelpOverlay()
	}

	// Minimum usable dimensions for side-by-side panes
	const minWidth = 60
	const minHeight = 20
	if m.width < minWidth || m.height < minHeight {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff5555")).
			Render(fmt.Sprintf("Terminal too small (%dx%d). Minimum: %dx%d", m.width, m.height, minWidth, minHeight))
	}

	// Dracula colors
	purple := lipgloss.Color("#bd93f9")
	comment := lipgloss.Color("#6272a4")

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
	leftColor := comment
	rightColor := comment
	if m.focus == focusCommands {
		leftColor = purple
	} else {
		rightColor = purple
	}

	// Calculate dimensions
	gap := 2
	titleLines := 1
	helpLines := 2
	paneWidth := (m.width - gap) / 2
	paneHeight := m.height - helpLines - titleLines

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
	} else {
		leftLines = append(leftLines, "")
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
			cursorColor := purple
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
	taskLines = append(taskLines, "")
	for _, line := range m.taskPreview {
		taskLines = append(taskLines, " "+line)
	}

	// Build right pane content (with padding)
	var configLines []string
	configLines = append(configLines, "")

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

	// Draw boxes
	leftBox := drawBox(leftLines, paneWidth, paneHeight, leftColor, "Commands")

	taskTitle := "Tasks to run"
	if len(m.visibleItems) > 0 {
		vItem := m.visibleItems[m.cursor]
		if vItem.item.Command != "" {
			taskTitle = fmt.Sprintf("Tasks for %s", strings.TrimSpace(vItem.path))
		}
	}
	taskBox := drawBox(taskLines, paneWidth, taskPaneHeight, comment, taskTitle)

	configBox := drawBox(configLines, paneWidth, configPaneHeight, rightColor, "Configuration")

	// Join boxes horizontally
	leftBoxLines := strings.Split(leftBox, "\n")
	taskBoxLines := strings.Split(taskBox, "\n")
	configBoxLines := strings.Split(configBox, "\n")
	rightBoxLines := append(taskBoxLines, configBoxLines...)

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
