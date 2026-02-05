package ui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/logger"
)

// Start launches the TUI and returns the selected command and collected inputs
func Start(version string) (string, map[string]string, error) {
	cfg, err := config.LoadDefaultConfig()
	cfgFound := true
	var initialErr error

	if err != nil {
		if os.IsNotExist(err) {
			cfg = &config.Config{SourcePath: "cleat.yaml"}
			cfgFound = false
		} else {
			// Configuration exists but failed to load (e.g., syntax error)
			logger.Error("failed to load config during TUI startup", err, nil)
			cfg = &config.Config{SourcePath: "cleat.yaml"}
			cfgFound = false
			initialErr = err
		}
	}

	exec := executor.Default
	m := InitialModel(cfg, cfgFound, version, exec)
	if initialErr != nil {
		m.fatalError = initialErr
	}

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
