package history

import (
	"os"
	"path/filepath"
	"sort"

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
		return err
	}

	statsFile, err := getStatsFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(statsFile), 0755); err != nil {
		return err
	}

	return os.WriteFile(statsFile, data, 0644)
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
		return ProjectStats{}, err
	}

	var stats ProjectStats
	if err := yaml.Unmarshal(data, &stats); err != nil {
		return ProjectStats{}, err
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
