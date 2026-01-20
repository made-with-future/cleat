package strategy

import (
	"fmt"
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/task"
)

// Strategy defines how to execute a command
type Strategy interface {
	// Name returns the command name (e.g., "build", "run")
	Name() string

	// Tasks returns all tasks this strategy may execute
	Tasks() []task.Task

	// Execute runs the strategy with dependency resolution
	Execute(cfg *config.Config, exec executor.Executor) error

	// ResolveTasks returns the list of tasks to be executed in order
	ResolveTasks(cfg *config.Config) ([]task.Task, error)

	// ReturnToUI returns true if the TUI should be restored after execution
	ReturnToUI() bool
}

// ExecutionMode determines how tasks are run
type ExecutionMode int

const (
	// Serial runs tasks one at a time
	Serial ExecutionMode = iota
	// Parallel runs independent tasks concurrently (future)
	// Parallel
)

// BaseStrategy provides common execution logic
type BaseStrategy struct {
	name       string
	tasks      []task.Task
	mode       ExecutionMode
	returnToUI bool
}

func NewBaseStrategy(name string, tasks []task.Task) *BaseStrategy {
	return &BaseStrategy{
		name:       name,
		tasks:      tasks,
		mode:       Serial,
		returnToUI: false,
	}
}

func (s *BaseStrategy) Name() string       { return s.name }
func (s *BaseStrategy) Tasks() []task.Task { return s.tasks }
func (s *BaseStrategy) ReturnToUI() bool   { return s.returnToUI }

func (s *BaseStrategy) SetReturnToUI(v bool) *BaseStrategy {
	s.returnToUI = v
	return s
}

func (s *BaseStrategy) ResolveTasks(cfg *config.Config) ([]task.Task, error) {
	return s.buildExecutionPlan(cfg)
}

// Execute runs tasks in dependency order
func (s *BaseStrategy) Execute(cfg *config.Config, exec executor.Executor) error {
	// Build execution plan respecting dependencies
	plan, err := s.buildExecutionPlan(cfg)
	if err != nil {
		return err
	}

	if len(plan) == 0 {
		fmt.Printf("No tasks to run for '%s' based on current configuration\n", s.name)
		return nil
	}

	// Execute tasks
	for _, t := range plan {
		if err := t.Run(cfg, exec); err != nil {
			return fmt.Errorf("task '%s' failed: %w", t.Name(), err)
		}
	}

	fmt.Printf("==> %s completed successfully\n", s.name)
	return nil
}

// buildExecutionPlan returns tasks in dependency order, filtering by ShouldRun
func (s *BaseStrategy) buildExecutionPlan(cfg *config.Config) ([]task.Task, error) {
	// Build lookup map
	taskMap := make(map[string]task.Task)
	for _, t := range s.tasks {
		taskMap[t.Name()] = t
	}

	// Filter to tasks that should run
	var candidates []task.Task
	for _, t := range s.tasks {
		if t.ShouldRun(cfg) {
			candidates = append(candidates, t)
		}
	}

	// Topological sort for dependency order
	return topologicalSort(candidates, taskMap, cfg)
}

// topologicalSort orders tasks respecting dependencies
func topologicalSort(tasks []task.Task, allTasks map[string]task.Task, cfg *config.Config) ([]task.Task, error) {
	// Track which tasks we need to run
	needed := make(map[string]bool)
	for _, t := range tasks {
		needed[t.Name()] = true
	}

	// Build in-degree map and adjacency list
	inDegree := make(map[string]int)
	dependents := make(map[string][]string) // dep -> tasks that depend on it

	for _, t := range tasks {
		name := t.Name()
		if _, exists := inDegree[name]; !exists {
			inDegree[name] = 0
		}

		for _, dep := range t.Dependencies() {
			// Only count dependency if the dep task exists AND should run
			if depTask, exists := allTasks[dep]; exists && depTask.ShouldRun(cfg) {
				inDegree[name]++
				dependents[dep] = append(dependents[dep], name)
				// Ensure dependency is in our needed set
				needed[dep] = true
			}
		}
	}

	// Add any dependencies we discovered that weren't in original candidates
	for name := range needed {
		if t, exists := allTasks[name]; exists && t.ShouldRun(cfg) {
			found := false
			for _, existing := range tasks {
				if existing.Name() == name {
					found = true
					break
				}
			}
			if !found {
				tasks = append(tasks, t)
				if _, exists := inDegree[name]; !exists {
					inDegree[name] = 0
				}
			}
		}
	}

	// Kahn's algorithm
	var queue []task.Task
	for _, t := range tasks {
		if inDegree[t.Name()] == 0 {
			queue = append(queue, t)
		}
	}

	var result []task.Task
	for len(queue) > 0 {
		t := queue[0]
		queue = queue[1:]
		result = append(result, t)

		for _, depName := range dependents[t.Name()] {
			inDegree[depName]--
			if inDegree[depName] == 0 {
				for _, candidate := range tasks {
					if candidate.Name() == depName {
						queue = append(queue, candidate)
						break
					}
				}
			}
		}
	}

	// Check for cycles
	if len(result) != len(tasks) {
		return nil, fmt.Errorf("circular dependency detected in tasks")
	}

	return result, nil
}

// Registry holds all available strategies
var Registry = make(map[string]func(*config.Config) Strategy)

// Register adds a strategy constructor to the registry
func Register(name string, constructor func(*config.Config) Strategy) {
	Registry[name] = constructor
}

// Get returns a strategy by name
func Get(name string, cfg *config.Config) (Strategy, bool) {
	constructor, ok := Registry[name]
	if !ok {
		return nil, false
	}
	return constructor(cfg), true
}

// ResolveCommandTasks returns the execution plan for a command string
func ResolveCommandTasks(command string, cfg *config.Config) ([]task.Task, error) {
	s := GetStrategyForCommand(command, cfg)
	if s == nil {
		return nil, fmt.Errorf("unknown command: %s", command)
	}
	return s.ResolveTasks(cfg)
}

// GetStrategyForCommand returns a strategy by its command string
func GetStrategyForCommand(command string, cfg *config.Config) Strategy {
	if cfg == nil {
		s, ok := Get(command, nil)
		if !ok {
			return nil
		}
		return s
	}

	if strings.HasPrefix(command, "npm run ") {
		// Command might be "npm run script" or "npm run service:script"
		parts := strings.Split(strings.TrimPrefix(command, "npm run "), ":")
		if len(parts) == 2 {
			svcName := parts[0]
			script := parts[1]
			for i := range cfg.Services {
				if cfg.Services[i].Name == svcName {
					svc := &cfg.Services[i]
					// Find the NPM module in this service
					for j := range svc.Modules {
						mod := &svc.Modules[j]
						if mod.Npm != nil {
							// If there are multiple NPM modules, this might be ambiguous,
							// but for now we'll take the first one or the one that has the script
							for _, s := range mod.Npm.Scripts {
								if s == script {
									return NewNpmScriptStrategy(svc, mod.Npm, script)
								}
							}
						}
					}
					// If no module has the script, try the first NPM module
					for j := range svc.Modules {
						if svc.Modules[j].Npm != nil {
							return NewNpmScriptStrategy(svc, svc.Modules[j].Npm, script)
						}
					}
				}
			}
		} else {
			script := parts[0]
			// Try to find a service and module that has this script
			for i := range cfg.Services {
				svc := &cfg.Services[i]
				for j := range svc.Modules {
					mod := &svc.Modules[j]
					if mod.Npm != nil {
						for _, s := range mod.Npm.Scripts {
							if s == script {
								return NewNpmScriptStrategy(svc, mod.Npm, script)
							}
						}
					}
				}
			}
		}
	}

	// Handle service-specific django commands from TUI: "django migrate:backend"
	if strings.HasPrefix(command, "django ") {
		parts := strings.Split(command, ":")
		if len(parts) == 2 {
			baseCmd := parts[0]
			svcName := parts[1]
			var targetSvc *config.ServiceConfig
			for i := range cfg.Services {
				if cfg.Services[i].Name == svcName {
					targetSvc = &cfg.Services[i]
					break
				}
			}

			if targetSvc != nil {
				switch baseCmd {
				case "django migrate":
					return NewDjangoMigrateStrategy(targetSvc)
				case "django collectstatic":
					return NewDjangoCollectStaticStrategy(targetSvc)
				case "django create-user-dev":
					return NewDjangoCreateUserDevStrategy(targetSvc)
				}
			}
		}
	}

	s, ok := Get(command, cfg)
	if !ok {
		return nil
	}
	return s
}
