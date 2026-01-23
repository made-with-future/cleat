package history

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type HistoryEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Command   string            `json:"command"`
	Inputs    map[string]string `json:"inputs,omitempty"`
}

var userHomeDir = os.UserHomeDir

const (
	maxHistorySize = 50
)

func getHistoryFilePath() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(absCwd))
	// Use project directory name + short hash for better recognizability
	projectDirName := filepath.Base(absCwd)
	if projectDirName == "/" || projectDirName == "." || projectDirName == "" {
		projectDirName = "root"
	}

	id := fmt.Sprintf("%s-%x", projectDirName, hash[:4])

	return filepath.Join(home, ".cleat", id+".history.json"), nil
}

func Save(entry HistoryEntry) error {
	entries, _ := Load()

	// Prepend new entry
	entries = append([]HistoryEntry{entry}, entries...)

	// Limit to maxHistorySize
	if len(entries) > maxHistorySize {
		entries = entries[:maxHistorySize]
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	historyFile, err := getHistoryFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(historyFile), 0755); err != nil {
		return err
	}

	return os.WriteFile(historyFile, data, 0644)
}

func Load() ([]HistoryEntry, error) {
	historyFile, err := getHistoryFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(historyFile)
	if err != nil {
		return nil, err
	}

	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}
