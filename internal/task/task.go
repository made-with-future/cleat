package task

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
)

// Task represents an atomic unit of work
type Task interface {
	// Name returns a unique identifier for this task
	Name() string

	// Description returns a human-readable description
	Description() string

	// Dependencies returns task names that must run before this task
	Dependencies() []string

	// ShouldRun determines if this task applies given the config
	ShouldRun(cfg *config.Config) bool

	// Run executes the task
	Run(cfg *config.Config, exec executor.Executor) error

	// Commands returns the actual CLI commands that will be executed
	Commands(cfg *config.Config) [][]string

	// Requirements returns the input requirements for this task
	Requirements(cfg *config.Config) []InputRequirement
}

// InputRequirement represents a piece of information needed from the user
type InputRequirement struct {
	Key     string
	Prompt  string
	Default string
}

// BaseTask provides common defaults for tasks
type BaseTask struct {
	TaskName        string
	TaskDescription string
	TaskDeps        []string
}

func (t *BaseTask) Name() string           { return t.TaskName }
func (t *BaseTask) Description() string    { return t.TaskDescription }
func (t *BaseTask) Dependencies() []string { return t.TaskDeps }
func (t *BaseTask) Requirements(cfg *config.Config) []InputRequirement {
	return nil
}

// ShouldUseOp checks if 1Password CLI is available and if any .env file in .envs/ contains "op://"
func ShouldUseOp(baseDir string) bool {
	if _, err := exec.LookPath("op"); err != nil {
		return false
	}

	envsDir := filepath.Join(baseDir, ".envs")
	entries, err := os.ReadDir(envsDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".env") {
			content, err := os.ReadFile(filepath.Join(envsDir, entry.Name()))
			if err == nil && strings.Contains(string(content), "op://") {
				return true
			}
		}
	}
	return false
}
