package theme

import "github.com/charmbracelet/lipgloss"

var (
	// Theme colors using terminal ANSI colors (0-15)
	// These will respect the user's terminal theme.
	Purple  = lipgloss.Color("5")  // Magenta
	Cyan    = lipgloss.Color("6")  // Cyan
	Comment = lipgloss.Color("8")  // Gray (Bright Black)
	Green   = lipgloss.Color("2")  // Green
	Red     = lipgloss.Color("1")  // Red
	Orange  = lipgloss.Color("3")  // Yellow
	FG      = lipgloss.Color("7")  // White
	White   = lipgloss.Color("15") // Bright White
)
