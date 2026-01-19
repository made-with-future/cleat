package task

import (
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
