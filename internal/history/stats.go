package history

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/madewithfuture/cleat/internal/logger"
	"gopkg.in/yaml.v3"
)

type CommandStat struct {
	Count int64 `yaml:"count"`
}

type ProjectStats struct {
	Commands map[string]CommandStat `yaml:"commands"`
}

func getStatsFilePath() (string, error) {
	return getFilePath("stats.yaml")
}

func UpdateStats(command string) error {
	stats, err := LoadStats()
	if err != nil {
		logger.Warn("failed to load existing stats before update", map[string]interface{}{"error": err.Error()})
		// If loading fails, start fresh
		stats = ProjectStats{
			Commands: make(map[string]CommandStat),
		}
	}

	if stats.Commands == nil {
		stats.Commands = make(map[string]CommandStat)
	}

	cmdStat := stats.Commands[command]
	cmdStat.Count++
	stats.Commands[command] = cmdStat

	return SaveStats(stats)
}

func SaveStats(stats ProjectStats) error {
	data, err := yaml.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	statsFile, err := getStatsFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(statsFile), 0755); err != nil {
		return fmt.Errorf("failed to create stats directory: %w", err)
	}

	if err := os.WriteFile(statsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write stats file: %w", err)
	}
	return nil
}

func LoadStats() (ProjectStats, error) {
	statsFile, err := getStatsFilePath()
	if err != nil {
		return ProjectStats{}, err
	}

	data, err := os.ReadFile(statsFile)
	if err != nil {
		// If file doesn't exist, return empty stats
		if os.IsNotExist(err) {
			return ProjectStats{Commands: make(map[string]CommandStat)}, nil
		}
		return ProjectStats{}, fmt.Errorf("failed to read stats file: %w", err)
	}

	var stats ProjectStats
	if err := yaml.Unmarshal(data, &stats); err != nil {
		return ProjectStats{}, fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	return stats, nil
}

type CommandCount struct {
	Command string
	Count   int64
}

func GetTopCommands(limit int) ([]CommandCount, error) {
	stats, err := LoadStats()
	if err != nil {
		return nil, err
	}

	var sorted []CommandCount
	for cmd, stat := range stats.Commands {
		sorted = append(sorted, CommandCount{Command: cmd, Count: stat.Count})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Count > sorted[j].Count
	})

	if len(sorted) > limit {
		sorted = sorted[:limit]
	}

	return sorted, nil
}
