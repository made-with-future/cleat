package ui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/logger"
)

type programRunner interface {
	Run() (tea.Model, error)
}

var runnerFactory = func(m tea.Model) programRunner {
	return tea.NewProgram(m, tea.WithAltScreen())
}

// Start launches the TUI and returns the selected command and collected inputs
func Start(version string, configPath string) (string, map[string]string, error) {
	cfg, err := config.LoadConfigWithDefault(configPath)
	cfgFound := true
	var initialErr error

	if err != nil {
		// This should only happen if there's a syntax error or similar,
		// as LoadConfigWithDefault handles missing files.
		logger.Error("failed to load config during TUI startup", err, nil)
		cfg = &config.Config{SourcePath: configPath}
		cfgFound = false
		initialErr = err
	} else {
		// Check if the file actually exists to set cfgFound correctly
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			cfgFound = false
		}
	}

	exec := executor.Default
	m := InitialModel(cfg, cfgFound, version, exec)
	if initialErr != nil {
		m.fatalError = initialErr
	}

	p := runnerFactory(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", nil, err
	}

	if fm, ok := finalModel.(model); ok {
		return fm.selectedCommand, fm.collectedInputs, nil
	}

	return "", nil, nil
}
