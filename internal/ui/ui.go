package ui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/madewithfuture/cleat/internal/config"
)

// Start launches the TUI and returns the selected command and collected inputs
func Start() (string, map[string]string, error) {
	cfg, err := config.LoadDefaultConfig()
	cfgFound := true
	if err != nil {
		if os.IsNotExist(err) {
			cfg = &config.Config{SourcePath: "cleat.yaml"}
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
