package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// View renders the current UI state
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
			Foreground(themeRed).
			Render(fmt.Sprintf("Terminal too small (%dx%d). Minimum: %dx%d", m.width, m.height, minWidth, minHeight))
	}

	base := m.renderMainUI()

	if m.state == stateInputCollection {
		return m.overlay(base, m.renderInputModal())
	}

	if m.state == stateConfirmClearHistory {
		return m.overlay(base, m.renderConfirmModal())
	}

	if m.state == stateWorkflowNameInput {
		return m.overlay(base, m.renderWorkflowNameModal())
	}

	if m.state == stateWorkflowLocationSelection {
		return m.overlay(base, m.renderWorkflowLocationModal())
	}

	if m.state == stateShowingConfig {
		return m.overlay(base, m.renderConfigModal())
	}

	// Show help overlay if active
	if m.showHelp {
		return m.overlay(base, m.renderHelpOverlay())
	}

	return base
}

// renderMainUI renders the main TUI with all panes
func (m model) renderMainUI() string {
	// Theme colors using terminal ANSI colors
	purple := themePurple
	cyan := themeCyan
	comment := themeComment
	green := themeGreen
	red := themeRed
	orange := themeOrange

	// Build title bar
	title := " Cleat "
	if m.version != "" {
		title = fmt.Sprintf(" Cleat %s ", m.version)
	}
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
	taskColor := comment

	switch m.focus {
	case focusCommands:
		commandsColor = purple
	case focusHistory:
		historyColor = purple
	case focusTasks:
		taskColor = purple
	}

	// Calculate dimensions
	gap := 2
	titleLines := 2
	helpLines := 3
	leftPaneWidth, rightPaneWidth := m.paneWidths()
	paneHeight := m.height - helpLines - titleLines

	// Height split
	topPaneHeight := paneHeight / 2
	bottomPaneHeight := paneHeight - topPaneHeight

	// Build pane content
	leftLines := m.buildCommandsContent(comment, purple, cyan)
	taskLines := m.buildTaskPreviewContent()
	historyLines := m.buildHistoryContent(comment, cyan, green, red, orange, rightPaneWidth)

	// Draw boxes
	commandsBox := drawBox(leftLines, leftPaneWidth, topPaneHeight, commandsColor, "Commands", m.focus == focusCommands)
	historyBox := drawBox(historyLines, rightPaneWidth, topPaneHeight, historyColor, "Command History", m.focus == focusHistory)

	taskTitle := m.buildTaskTitle()
	taskBox := drawBox(taskLines, m.width, bottomPaneHeight, taskColor, taskTitle, m.focus == focusTasks)

	// Join Row 1 boxes side-by-side
	commandsBoxLines := strings.Split(commandsBox, "\n")
	historyBoxLines := strings.Split(historyBox, "\n")

	var combined strings.Builder
	for i := 0; i < topPaneHeight; i++ {
		left := ""
		right := ""
		if i < len(commandsBoxLines) {
			left = commandsBoxLines[i]
		}
		if i < len(historyBoxLines) {
			right = historyBoxLines[i]
		}
		combined.WriteString(left)
		combined.WriteString(strings.Repeat(" ", gap))
		combined.WriteString(right)
		combined.WriteString("\n")
	}

	// Append Row 2 (Tasks)
	combined.WriteString(taskBox)

	// Help line
	helpText := lipgloss.NewStyle().Foreground(comment).Render("↑/↓: navigate • enter: select/toggle • tab: switch pane • c: config • ?: help • q: quit")
	if m.state == stateCreatingWorkflow {
		helpText = lipgloss.NewStyle().Foreground(comment).Render("↑/↓: navigate • space/enter: select • c: confirm • esc: cancel")
	}
	if m.state == stateShowingConfig {
		helpText = lipgloss.NewStyle().Foreground(comment).Render("↑/↓: scroll • enter: edit • c/esc/q: close")
	}
	if !m.cfgFound {
		warning := lipgloss.NewStyle().Foreground(themeRed).Render("(no cleat.yaml)")
		separator := lipgloss.NewStyle().Foreground(comment).Render(" • ")
		helpText = warning + separator + helpText
	}

	helpText = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(helpText)

	return titleBar + "\n" + combined.String() + "\n\n" + helpText
}

// buildCommandsContent builds the commands pane content
func (m model) buildCommandsContent(comment, purple, cyan lipgloss.Color) []string {
	var leftLines []string

	if m.filtering {
		filterStyle := lipgloss.NewStyle().Foreground(purple)
		leftLines = append(leftLines, "  "+filterStyle.Render("/"+m.filterText+"█"))
	} else if m.filterText != "" {
		filterStyle := lipgloss.NewStyle().Foreground(comment)
		leftLines = append(leftLines, "  "+filterStyle.Render("/"+m.filterText))
	}

	visibleCount := m.visibleCommandCount()
	hasMoreAbove := m.scrollOffset > 0
	hasMoreBelow := m.scrollOffset+visibleCount < len(m.visibleItems)

	// Show scroll up indicator
	if hasMoreAbove {
		leftLines = append(leftLines, "    "+lipgloss.NewStyle().Foreground(comment).Render("▲ more"))
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
			leftLines = append(leftLines, "  "+lipgloss.NewStyle().Foreground(cursorColor).Render("> "+label))
		} else {
			leftLines = append(leftLines, "    "+label)
		}
	}

	// Show scroll down indicator
	if hasMoreBelow {
		leftLines = append(leftLines, "    "+lipgloss.NewStyle().Foreground(comment).Render("▼ more"))
	}

	return leftLines
}

// buildTaskPreviewContent builds the task preview pane content
func (m model) buildTaskPreviewContent() []string {
	var taskLines []string
	comment := themeComment

	visibleTasksCount := m.visibleTasksCount()
	hasMoreAbove := m.taskScrollOffset > 0
	hasMoreBelow := m.taskScrollOffset+visibleTasksCount < len(m.taskPreview)

	// Show scroll up indicator
	if hasMoreAbove {
		taskLines = append(taskLines, "    "+lipgloss.NewStyle().Foreground(comment).Render("▲ more"))
	}

	// Render visible tasks
	endIdx := m.taskScrollOffset + visibleTasksCount
	if endIdx > len(m.taskPreview) {
		endIdx = len(m.taskPreview)
	}

	for i := m.taskScrollOffset; i < endIdx; i++ {
		taskLines = append(taskLines, "  "+m.taskPreview[i])
	}

	// Show scroll down indicator
	if hasMoreBelow {
		taskLines = append(taskLines, "    "+lipgloss.NewStyle().Foreground(comment).Render("▼ more"))
	}

	return taskLines
}

// buildConfigContent builds the configuration pane content
func (m model) buildConfigContent(comment, purple lipgloss.Color) []string {
	var configLines []string

	allConfigLines := m.buildConfigLines()
	visibleConfigCount := m.visibleConfigCount()
	hasMoreConfigAbove := m.configScrollOffset > 0
	hasMoreConfigBelow := m.configScrollOffset+visibleConfigCount < len(allConfigLines)

	// Show scroll up indicator
	if hasMoreConfigAbove {
		configLines = append(configLines, "    "+lipgloss.NewStyle().Foreground(comment).Render("▲ more"))
	}

	// Render visible config lines
	endConfigIdx := m.configScrollOffset + visibleConfigCount
	if endConfigIdx > len(allConfigLines) {
		endConfigIdx = len(allConfigLines)
	}
	for i := m.configScrollOffset; i < endConfigIdx; i++ {
		configLines = append(configLines, "  "+allConfigLines[i])
	}

	// Show scroll down indicator
	if hasMoreConfigBelow {
		configLines = append(configLines, "    "+lipgloss.NewStyle().Foreground(comment).Render("▼ more"))
	}

	configLines = append(configLines, "")
	// Action hint
	if m.focus == focusConfig {
		actionText := "Press Enter to edit"
		if !m.cfgFound {
			actionText = "Press Enter to create"
		}
		configLines = append(configLines, "  "+lipgloss.NewStyle().Foreground(purple).Render(actionText))
	}

	return configLines
}

// buildHistoryContent builds the history pane content
func (m model) buildHistoryContent(comment, cyan, green, red, orange lipgloss.Color, rightPaneWidth int) []string {
	var historyLines []string
	visibleHistoryCount := m.visibleHistoryCount()
	hasMoreHistoryAbove := m.historyOffset > 0
	hasMoreHistoryBelow := m.historyOffset+visibleHistoryCount < len(m.history)

	if hasMoreHistoryAbove {
		historyLines = append(historyLines, "    "+lipgloss.NewStyle().Foreground(comment).Render("▲ more"))
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

		// Add selection mark if creating workflow
		if m.state == stateCreatingWorkflow {
			mark := "[ ]"
			for j, idx := range m.selectedWorkflowIndices {
				if idx == i {
					mark = fmt.Sprintf("[%d]", j+1)
					break
				}
			}
			icon = mark + " " + icon
		}

		renderedIcon := lipgloss.NewStyle().Foreground(iconColor).Render(icon)

		// Workflow marker
		marker := " "
		if entry.WorkflowRunID != "" {
			marker = lipgloss.NewStyle().Foreground(orange).Render("┃")
		}

		// Format: Command (aligned left) ... Date Time (aligned right)
		ts := entry.Timestamp.Format("2006-01-02 15:04")

		// Calculate widths for dynamic layout
		prefixWidth := 4 // "  > " or "    "
		iconWidth := 2   // icon + space
		if m.state == stateCreatingWorkflow {
			iconWidth = 6 // "[1] " + icon + " "
		}
		paddingWidth := 4 // spaces before marker (3) + marker (1)

		// Total non-content width including borders (2), prefix (4), icon (2/6), padding/marker (4), and some extra right margin (2)
		contentWidth := rightPaneWidth - 2 - prefixWidth - iconWidth - paddingWidth - 2
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

		label = label + "   " + marker

		if i == m.historyCursor {
			cursorColor := cyan
			if m.focus != focusHistory {
				cursorColor = comment
			}
			historyLines = append(historyLines, "  "+lipgloss.NewStyle().Foreground(cursorColor).Render("> "+label))
		} else {
			historyLines = append(historyLines, "    "+label)
		}
	}
	if hasMoreHistoryBelow {
		historyLines = append(historyLines, "    "+lipgloss.NewStyle().Foreground(comment).Render("▼ more"))
	}

	return historyLines
}

// buildTaskTitle builds the title for the task preview pane
func (m model) buildTaskTitle() string {
	taskTitle := "Tasks to run"
	if m.focus == focusHistory {
		if len(m.history) > 0 {
			entry := m.history[m.historyCursor]
			taskTitle = fmt.Sprintf("Tasks for %s", entry.Command)
		}
	} else if len(m.visibleItems) > 0 {
		vItem := m.visibleItems[m.cursor]
		if vItem.item.Command != "" {
			taskTitle = fmt.Sprintf("Tasks for %s", strings.TrimSpace(vItem.path))
		}
	}
	return taskTitle
}

// buildConfigLines generates the config preview lines
func (m model) buildConfigLines() []string {
	var configLines []string

	if !m.cfgFound {
		configLines = append(configLines, " "+lipgloss.NewStyle().Foreground(themeRed).Italic(true).Render("No cleat.yaml found"))
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

// renderInputModal renders the input collection modal
func (m model) renderInputModal() string {
	purple := themePurple
	fg := themeFG

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
		lipgloss.NewStyle().Foreground(themeComment).Render("  Enter: continue • Esc: cancel"),
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

// renderConfirmModal renders the confirmation dialog
func (m model) renderWorkflowNameModal() string {
	width := 60
	height := 10

	purple := themePurple
	comment := themeComment

	title := " Create Workflow "
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)

	content := "\n Enter a name for your workflow:\n\n"
	content += " " + m.textInput.View() + "\n\n"
	content += lipgloss.NewStyle().Foreground(comment).Render(" enter: confirm • esc: cancel")

	lines := strings.Split(content, "\n")
	var centeredLines []string
	for _, line := range lines {
		padding := (width - lipgloss.Width(line)) / 2
		if padding < 0 {
			padding = 0
		}
		centeredLines = append(centeredLines, strings.Repeat(" ", padding)+line)
	}

	modal := drawBox(centeredLines, width, height, purple, titleStyle.Render(title), true)
	return modal
}

func (m model) renderConfirmModal() string {
	purple := themePurple
	fg := themeFG
	red := themeRed

	title := lipgloss.NewStyle().Bold(true).Foreground(red).Render("Clear History")

	content := []string{
		title,
		"",
		lipgloss.NewStyle().Foreground(fg).Render("Are you sure you want to clear history?"),
		"",
		lipgloss.NewStyle().Foreground(themeComment).Render("  y: confirm • n/Esc: cancel"),
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Foreground(fg).
		Padding(0, 2)

	box := boxStyle.Render(strings.Join(content, "\n"))
	return box
}

// renderHelpOverlay renders a centered help modal
func (m model) renderWorkflowLocationModal() string {
	purple := themePurple
	comment := themeComment
	cyan := themeCyan

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(purple).MarginBottom(1)
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Padding(1, 4).
		Width(40)

	options := []string{"Project (cleat.workflows.yaml)", "User (~/.cleat/...workflows.yaml)"}
	var renderedOptions []string

	for i, opt := range options {
		if i == m.workflowLocationIdx {
			renderedOptions = append(renderedOptions, lipgloss.NewStyle().Foreground(cyan).Render("> "+opt))
		} else {
			renderedOptions = append(renderedOptions, "  "+opt)
		}
	}

	content := titleStyle.Render("Save Workflow to...") + "\n\n" +
		strings.Join(renderedOptions, "\n") + "\n\n" +
		lipgloss.NewStyle().Foreground(comment).Render("↑/↓: navigate • enter: select • esc: cancel")

	return modalStyle.Render(content)
}

func (m model) renderConfigModal() string {
	purple := themePurple
	comment := themeComment

	width := 60
	if width > m.width-10 {
		width = m.width - 10
	}
	if width < 60 {
		width = 60
	}
	if width > m.width {
		width = m.width
	}
	height := m.height - 10
	if height < 10 {
		height = 10
	}
	if height > m.height {
		height = m.height
	}

	configLines := m.buildConfigContent(comment, purple)

	title := " Configuration "
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)

	modal := drawBox(configLines, width, height, purple, titleStyle.Render(title), true)
	return modal
}

func (m model) renderHelpOverlay() string {
	purple := themePurple
	comment := themeComment
	fg := themeFG

	title := lipgloss.NewStyle().Bold(true).Foreground(purple).Render("Keyboard Shortcuts")

	helpItems := []string{
		"",
		title,
		"",
		"  ↑/k        Move up",
		"  ↓/j        Move down",
		"  e          Expand all",
		"  C          Collapse all",
		"  /          Filter commands",
		"  c          Show configuration",
		"  Enter      Select/Toggle / Edit config (in config modal)",
		"  1-9        Jump to history item",
		"  t          Jump to task panel",
		"  x          Clear history (history pane)",
		"  w          Create workflow from history (history pane)",
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
			renderedTitle = lipgloss.NewStyle().Foreground(themeWhite).Bold(true).Render(displayTitle)
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

// overlay overlays foreground content on background at center position
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
