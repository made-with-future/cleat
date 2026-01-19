package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/madewithfuture/cleat/internal/config"
)

const configPath = "cleat.yaml"

const defaultConfigTemplate = `# Cleat configuration
# See https://github.com/madewithfuture/cleat for documentation

docker: true
django: false
django_service: backend
npm:
  service: backend-node
  scripts:
    - build
`

type focus int

const (
	focusCommands focus = iota
	focusConfig
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
}

// editorFinishedMsg is sent when the editor process exits
type editorFinishedMsg struct{ err error }

type model struct {
	cfg             *config.Config
	cfgFound        bool
	quitting        bool
	width           int
	height          int
	tree            []CommandItem
	visibleItems    []visibleItem
	cursor          int
	scrollOffset    int
	focus           focus
	selectedCommand string
	showHelp        bool
}

func InitialModel(cfg *config.Config, cfgFound bool) model {
	m := model{
		cfg:      cfg,
		cfgFound: cfgFound,
		tree:     buildCommandTree(cfg),
		focus:    focusCommands,
	}
	m.updateVisibleItems()
	return m
}

func (m *model) updateVisibleItems() {
	m.visibleItems = []visibleItem{}
	for i := range m.tree {
		m.flatten(&m.tree[i], 0)
	}
}

func (m *model) flatten(item *CommandItem, level int) {
	m.visibleItems = append(m.visibleItems, visibleItem{item: item, level: level})
	if item.Expanded && len(item.Children) > 0 {
		for i := range item.Children {
			m.flatten(&item.Children[i], level+1)
		}
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		}
		return m, nil

	case tea.KeyMsg:
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
			}
		case tea.KeyDown:
			if m.focus == focusCommands && m.cursor < len(m.visibleItems)-1 {
				m.cursor++
				visibleCount := m.visibleCommandCount()
				if m.cursor >= m.scrollOffset+visibleCount {
					m.scrollOffset = m.cursor - visibleCount + 1
				}
			}
		case tea.KeyEnter:
			if m.focus == focusCommands {
				item := m.visibleItems[m.cursor]
				if len(item.item.Children) > 0 {
					item.item.Expanded = !item.item.Expanded
					m.updateVisibleItems()
					if m.cursor >= len(m.visibleItems) {
						m.cursor = len(m.visibleItems) - 1
					}
				} else {
					m.selectedCommand = item.item.Command
					m.quitting = true
					return m, tea.Quit
				}
			} else if m.focus == focusConfig {
				return m, m.openEditor()
			}
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "q":
				m.quitting = true
				return m, tea.Quit
			case "?":
				m.showHelp = true
			case "k":
				if m.focus == focusCommands && m.cursor > 0 {
					m.cursor--
					if m.cursor < m.scrollOffset {
						m.scrollOffset = m.cursor
					}
				}
			case "j":
				if m.focus == focusCommands && m.cursor < len(m.visibleItems)-1 {
					m.cursor++
					visibleCount := m.visibleCommandCount()
					if m.cursor >= m.scrollOffset+visibleCount {
						m.scrollOffset = m.cursor - visibleCount + 1
					}
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
			},
			Expanded: true,
		})
	}

	if len(cfg.Npm.Scripts) > 0 {
		npmItem := CommandItem{
			Label:    "npm",
			Expanded: true,
		}
		for _, script := range cfg.Npm.Scripts {
			npmItem.Children = append(npmItem.Children, CommandItem{
				Label:   fmt.Sprintf("run %s", script),
				Command: fmt.Sprintf("npm run %s", script),
			})
		}
		tree = append(tree, npmItem)
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
	// Subtract: 2 for borders, 1 for title, 1 for blank line after title, 1 for potential scroll indicator
	availableLines := paneHeight - 2 - 1 - 1 - 1
	if availableLines < 1 {
		availableLines = 1
	}
	return availableLines
}

// drawBox draws a box with rounded corners around content
func drawBox(lines []string, width, height int, borderColor lipgloss.Color) string {
	colorStyle := lipgloss.NewStyle().Foreground(borderColor)

	innerWidth := width - 2 // Account for left and right borders

	var result strings.Builder

	// Top border
	result.WriteString(colorStyle.Render("╭" + strings.Repeat("─", innerWidth) + "╮"))
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

func (m model) View() string {
	if m.quitting {
		return ""
	}

	if m.width == 0 || m.height == 0 {
		return "Initializing..."
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

	// Build left pane content (with padding)
	var leftLines []string
	leftLines = append(leftLines, " "+lipgloss.NewStyle().Bold(true).Foreground(leftColor).Render("Commands"))
	leftLines = append(leftLines, "")

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

	// Build right pane content (with padding)
	var rightLines []string
	rightLines = append(rightLines, " "+lipgloss.NewStyle().Bold(true).Foreground(rightColor).Render("Configuration"))
	rightLines = append(rightLines, "")
	if !m.cfgFound {
		rightLines = append(rightLines, " "+lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")).Italic(true).Render("No cleat.yaml found"))
		rightLines = append(rightLines, "")
	}
	rightLines = append(rightLines, fmt.Sprintf(" Docker: %v", m.cfg.Docker))
	rightLines = append(rightLines, fmt.Sprintf(" Django: %v", m.cfg.Django))
	if m.cfg.DjangoService != "" {
		rightLines = append(rightLines, fmt.Sprintf("   Service: %s", m.cfg.DjangoService))
	}
	rightLines = append(rightLines, fmt.Sprintf(" NPM: %v", len(m.cfg.Npm.Scripts) > 0))
	if m.cfg.Npm.Service != "" {
		rightLines = append(rightLines, fmt.Sprintf("   Service: %s", m.cfg.Npm.Service))
	}
	rightLines = append(rightLines, "")
	// Action hint
	if m.focus == focusConfig {
		actionText := "Press Enter to edit"
		if !m.cfgFound {
			actionText = "Press Enter to create"
		}
		rightLines = append(rightLines, " "+lipgloss.NewStyle().Foreground(purple).Render(actionText))
	}

	// Draw boxes
	leftBox := drawBox(leftLines, paneWidth, paneHeight, leftColor)
	rightBox := drawBox(rightLines, paneWidth, paneHeight, rightColor)

	// Join boxes horizontally
	leftBoxLines := strings.Split(leftBox, "\n")
	rightBoxLines := strings.Split(rightBox, "\n")

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

func Start() (string, error) {
	cfg, err := config.LoadConfig("cleat.yaml")
	cfgFound := true
	if err != nil {
		if os.IsNotExist(err) {
			cfg = &config.Config{}
			cfgFound = false
		} else {
			return "", err
		}
	}

	m := InitialModel(cfg, cfgFound)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	if fm, ok := finalModel.(model); ok {
		return fm.selectedCommand, nil
	}

	return "", nil
}
