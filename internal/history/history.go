package history

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/madewithfuture/cleat/internal/config"
)

type HistoryEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Command   string            `json:"command"`
	Inputs    map[string]string `json:"inputs,omitempty"`
	Success   bool              `json:"success"`
}

var UserHomeDir = os.UserHomeDir

const (
	maxHistorySize = 50
)

func getHistoryFilePath() (string, error) {
	home, err := UserHomeDir()
	if err != nil {
		return "", err
	}

	root := config.FindProjectRoot()
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(absRoot))
	// Use project directory name + hash for better recognizability
	// Using 8 bytes (16 hex chars) to reduce collision risk while keeping filenames reasonable
	projectDirName := filepath.Base(absRoot)
	if projectDirName == "/" || projectDirName == "." || projectDirName == "" {
		projectDirName = "root"
	}

	id := fmt.Sprintf("%s-%x", projectDirName, hash[:8])

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

func Clear() error {
	historyFile, err := getHistoryFilePath()
	if err != nil {
		return err
	}
	return os.Remove(historyFile)
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
