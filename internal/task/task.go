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
			if FileUsesOp(filepath.Join(envsDir, entry.Name())) {
				return true
			}
		}
	}
	return false
}

// FileUsesOp checks if 1Password CLI is available and if the specified .env file contains "op://"
func FileUsesOp(envFile string) bool {
	if _, err := exec.LookPath("op"); err != nil {
		return false
	}

	content, err := os.ReadFile(envFile)
	if err == nil && strings.Contains(string(content), "op://") {
		return true
	}
	return false
}

// DetectEnvFile returns the paths to the .env file to use for tasks.
// execPath: path relative to searchDir, suitable for use with executor.RunWithDir(searchDir, ...).
// absPath: absolute path or relative to CWD, suitable for use with os.Open.
// displayPath: user-friendly path for logging.
func DetectEnvFile(searchDir string) (execPath string, absPath string, displayPath string) {
	// 1. Check if we're in a Terraform subdirectory and it has .env
	cwd, err := os.Getwd()
	if err == nil {
		// Check if cwd is or is under a .iac directory
		isTF := false
		parts := strings.Split(cwd, string(filepath.Separator))
		for _, part := range parts {
			if part == ".iac" {
				isTF = true
				break
			}
		}

		if isTF {
			envPath := filepath.Join(cwd, ".env")
			if _, err := os.Stat(envPath); err == nil {
				// Return path relative to searchDir so it works with RunWithDir(searchDir)
				relPath, err := filepath.Rel(searchDir, envPath)
				if err == nil {
					return relPath, envPath, ".env"
				}
				return envPath, envPath, ".env"
			}
		}
	}

	// 2. Fallback to .envs/dev.env in searchDir
	devEnvPath := filepath.Join(searchDir, ".envs/dev.env")
	if _, err := os.Stat(devEnvPath); err == nil {
		// The path relative to searchDir is just .envs/dev.env
		displayPath = ".envs/dev.env"
		if searchDir == "." || searchDir == "" {
			displayPath = "./.envs/dev.env"
		}
		return filepath.Join(".envs", "dev.env"), devEnvPath, displayPath
	}

	return "", "", ""
}
