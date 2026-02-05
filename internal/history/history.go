package history

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/logger"
	"gopkg.in/yaml.v3"
)

type HistoryEntry struct {
	Timestamp     time.Time         `json:"timestamp"`
	Command       string            `json:"command"`
	Inputs        map[string]string `json:"inputs,omitempty"`
	Success       bool              `json:"success"`
	WorkflowRunID string            `json:"workflow_run_id,omitempty"`
}

var UserHomeDir = os.UserHomeDir

const (
	maxHistorySize = 50
)

func getFilePath(suffix string) (string, error) {
	home, err := UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}

	id := config.GetProjectID()
	return filepath.Join(home, ".cleat", id+"."+suffix), nil
}

func getHistoryFilePath() (string, error) {
	return getFilePath("history.yaml")
}

func Save(entry HistoryEntry) error {
	entries, err := Load()
	if err != nil {
		logger.Warn("failed to load existing history before saving", map[string]interface{}{"error": err.Error()})
		// Initialize empty entries if load failed
		entries = []HistoryEntry{}
	}

	// Prepend new entry
	entries = append([]HistoryEntry{entry}, entries...)

	// Limit to maxHistorySize
	if len(entries) > maxHistorySize {
		entries = entries[:maxHistorySize]
	}

	data, err := yaml.Marshal(entries)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	historyFile, err := getHistoryFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(historyFile), 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	if err := os.WriteFile(historyFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}
	return nil
}

func Clear() error {
	historyFile, err := getHistoryFilePath()
	if err != nil {
		return err
	}
	if err := os.Remove(historyFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear history: %w", err)
	}
	return nil
}

func Load() ([]HistoryEntry, error) {
	historyFile, err := getHistoryFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var entries []HistoryEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return entries, nil
}
