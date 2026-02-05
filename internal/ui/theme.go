package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Theme colors using terminal ANSI colors (0-15)
	// These will respect the user's terminal theme.
	themePurple  = lipgloss.Color("5")  // Magenta
	themeCyan    = lipgloss.Color("6")  // Cyan
	themeComment = lipgloss.Color("8")  // Gray (Bright Black)
	themeGreen   = lipgloss.Color("2")  // Green
	themeRed     = lipgloss.Color("1")  // Red
	themeOrange  = lipgloss.Color("3")  // Yellow
	themeFG      = lipgloss.Color("7")  // White
	themeWhite   = lipgloss.Color("15") // Bright White
)
