package task

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/madewithfuture/cleat/internal/logger"
	"github.com/madewithfuture/cleat/internal/session"
)

// Task represents an atomic unit of work
type Task interface {
	// Name returns a unique identifier for this task
	Name() string

	// Description returns a human-readable description
	Description() string

	// Dependencies returns task names that must run before this task
	Dependencies() []string

	// ShouldRun determines if this task applies given the session context
	ShouldRun(sess *session.Session) bool

	// Run executes the task using session context
	Run(sess *session.Session) error

	// Commands returns the actual CLI commands that will be executed
	Commands(sess *session.Session) [][]string

	// Requirements returns the input requirements for this task
	Requirements(sess *session.Session) []InputRequirement
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
func (t *BaseTask) Requirements(sess *session.Session) []InputRequirement {
	return nil
}

// ShouldUseOp checks if 1Password CLI is available and if any .env file in .envs/ contains "op://"
func ShouldUseOp(baseDir string) bool {
	if _, err := exec.LookPath("op"); err != nil {
		logger.Debug("op CLI not found in PATH", nil)
		return false
	}

	envsDir := filepath.Join(baseDir, ".envs")
	entries, err := os.ReadDir(envsDir)
	if err != nil {
		if !os.IsNotExist(err) {
			logger.Warn("failed to read .envs directory", map[string]interface{}{"path": envsDir, "error": err.Error()})
		}
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
	if err != nil {
		if !os.IsNotExist(err) {
			logger.Warn("failed to read env file", map[string]interface{}{"path": envFile, "error": err.Error()})
		}
		return false
	}

	if strings.Contains(string(content), "op://") {
		logger.Debug("file uses 1Password (op:// detected)", map[string]interface{}{"file": envFile})
		return true
	}
	return false
}

// DetectEnvFile returns the paths to the .env file to use for tasks.
func DetectEnvFile(searchDir string) (execPath string, absPath string, displayPath string) {
	cwd, err := os.Getwd()
	if err == nil {
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
				relPath, err := filepath.Rel(searchDir, envPath)
				if err == nil {
					return relPath, envPath, ".env"
				}
				return envPath, envPath, ".env"
			}
		}
	}

	devEnvPath := filepath.Join(searchDir, ".envs/dev.env")
	if _, err := os.Stat(devEnvPath); err == nil {
		displayPath = ".envs/dev.env"
		if searchDir == "." || searchDir == "" {
			displayPath = "./.envs/dev.env"
		}
		return filepath.Join(".envs", "dev.env"), devEnvPath, displayPath
	}

	return "", "", ""
}