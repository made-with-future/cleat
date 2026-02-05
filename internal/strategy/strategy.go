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
}

// CommandProvider handles the mapping of command strings to strategies
type CommandProvider interface {
	// CanHandle returns true if this provider can resolve the given command
	CanHandle(command string) bool
	// GetStrategy returns the appropriate strategy for the command
	GetStrategy(command string, cfg *config.Config) Strategy
}

// RegistryProvider handles strategies registered via the global Registry
type RegistryProvider struct{}

func (p *RegistryProvider) CanHandle(command string) bool {
	_, ok := Registry[command]
	return ok
}

func (p *RegistryProvider) GetStrategy(command string, cfg *config.Config) Strategy {
	constructor, ok := Registry[command]
	if !ok {
		return nil
	}
	return constructor(cfg)
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
	name  string
	tasks []task.Task
	mode  ExecutionMode
}

func NewBaseStrategy(name string, tasks []task.Task) *BaseStrategy {
	return &BaseStrategy{
		name:  name,
		tasks: tasks,
		mode:  Serial,
	}
}

func (s *BaseStrategy) Name() string       { return s.name }
func (s *BaseStrategy) Tasks() []task.Task { return s.tasks }

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

	// Collect requirements from all tasks
	requirements := make(map[string]task.InputRequirement)
	for _, t := range plan {
		for _, req := range t.Requirements(cfg) {
			requirements[req.Key] = req
		}
	}

	// Prompt for missing inputs
	for key, req := range requirements {
		if _, ok := cfg.Inputs[key]; !ok {
			val, err := exec.Prompt(req.Prompt, req.Default)
			if err != nil {
				return fmt.Errorf("failed to get input for %s: %w", key, err)
			}
			cfg.Inputs[key] = val
		}
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

func GetStrategyForCommand(command string, cfg *config.Config) Strategy {
	for _, p := range GetProviders() {
		if p.CanHandle(command) {
			if s := p.GetStrategy(command, cfg); s != nil {
				return s
			}
		}
	}
	return nil
}

// GetProviders returns the prioritized list of command providers
func GetProviders() []CommandProvider {
	return []CommandProvider{
		&NpmProvider{},
		&DockerProvider{},
		&DjangoProvider{},
		&GcpProvider{},
		&TerraformProvider{},
		&RegistryProvider{},
	}
}

// NpmProvider handles NPM related commands
type NpmProvider struct{}

func (p *NpmProvider) CanHandle(command string) bool {
	return strings.HasPrefix(command, "npm install") || strings.HasPrefix(command, "npm run ")
}

func (p *NpmProvider) GetStrategy(command string, cfg *config.Config) Strategy {
	if cfg == nil {
		return nil
	}

	if strings.HasPrefix(command, "npm install") {
		return GetNpmInstallStrategy(command, cfg)
	}

	if strings.HasPrefix(command, "npm run ") {
		fullScript := strings.TrimPrefix(command, "npm run ")

		// 1. Try to match as svcName:script first
		if colonIdx := strings.Index(fullScript, ":"); colonIdx != -1 {
			svcName := fullScript[:colonIdx]
			script := fullScript[colonIdx+1:]

			for i := range cfg.Services {
				if cfg.Services[i].Name == svcName {
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
					for j := range svc.Modules {
						if svc.Modules[j].Npm != nil {
							return NewNpmScriptStrategy(svc, svc.Modules[j].Npm, script)
						}
					}
				}
			}
		}

		// 2. No service prefix match, search for the script name in all NPM modules
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			for j := range svc.Modules {
				mod := &svc.Modules[j]
				if mod.Npm != nil {
					for _, s := range mod.Npm.Scripts {
						if s == fullScript {
							return NewNpmScriptStrategy(svc, mod.Npm, fullScript)
						}
					}
				}
			}
		}

		// 3. Last resort: use the first service that has an NPM module
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			for j := range svc.Modules {
				if svc.Modules[j].Npm != nil {
					return NewNpmScriptStrategy(svc, svc.Modules[j].Npm, fullScript)
				}
			}
		}
	}

	return nil
}

// DockerProvider handles service-specific docker commands
type DockerProvider struct{}

func (p *DockerProvider) CanHandle(command string) bool {
	return strings.HasPrefix(command, "docker ")
}

func (p *DockerProvider) GetStrategy(command string, cfg *config.Config) Strategy {
	if cfg == nil {
		return nil
	}

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
			case "docker down":
				return NewDockerDownStrategyForService(targetSvc)
			case "docker rebuild":
				return NewDockerRebuildStrategyForService(targetSvc)
			case "docker remove-orphans":
				return NewDockerRemoveOrphansStrategyForService(targetSvc)
			}
		}
	}
	return nil
}

// DjangoProvider handles service-specific django commands
type DjangoProvider struct{}

func (p *DjangoProvider) CanHandle(command string) bool {
	return strings.HasPrefix(command, "django ")
}

func (p *DjangoProvider) GetStrategy(command string, cfg *config.Config) Strategy {
	if cfg == nil {
		return nil
	}

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
			case "django runserver":
				return NewDjangoRunServerStrategy(targetSvc)
			case "django migrate":
				return NewDjangoMigrateStrategy(targetSvc)
			case "django makemigrations":
				return NewDjangoMakeMigrationsStrategy(targetSvc)
			case "django collectstatic":
				return NewDjangoCollectStaticStrategy(targetSvc)
			case "django create-user-dev":
				return NewDjangoCreateUserDevStrategy(targetSvc)
			case "django gen-random-secret-key":
				return NewDjangoGenRandomSecretKeyStrategy(targetSvc)
			}
		}
	}
	return nil
}

// GcpProvider handles Google Cloud Platform commands
type GcpProvider struct{}

func (p *GcpProvider) CanHandle(command string) bool {
	return strings.HasPrefix(command, "gcp app-engine deploy") || strings.HasPrefix(command, "gcp app-engine promote")
}

func (p *GcpProvider) GetStrategy(command string, cfg *config.Config) Strategy {
	if cfg == nil {
		return nil
	}

	if strings.HasPrefix(command, "gcp app-engine deploy") {
		parts := strings.Split(command, ":")
		if len(parts) == 2 {
			svcName := parts[1]
			for i := range cfg.Services {
				if cfg.Services[i].Name == svcName {
					if cfg.Services[i].AppYaml != "" {
						return NewGCPAppEngineDeployStrategy(cfg.Services[i].AppYaml)
					}
				}
			}
		} else {
			if cfg.AppYaml != "" {
				return NewGCPAppEngineDeployStrategy(cfg.AppYaml)
			}
		}
	}

	if strings.HasPrefix(command, "gcp app-engine promote") {
		parts := strings.Split(command, ":")
		if len(parts) == 2 {
			return NewGCPAppEnginePromoteStrategy(parts[1])
		} else {
			return NewGCPAppEnginePromoteStrategy("")
		}
	}

	return nil
}

// TerraformProvider handles Terraform related commands
type TerraformProvider struct{}

func (p *TerraformProvider) CanHandle(command string) bool {
	return strings.HasPrefix(command, "terraform ")
}

func (p *TerraformProvider) GetStrategy(command string, cfg *config.Config) Strategy {
	if cfg == nil {
		return nil
	}

	parts := strings.Split(command, ":")
	baseCmd := parts[0]
	var env string
	if len(parts) == 2 {
		env = parts[1]
	}
	action := strings.TrimPrefix(baseCmd, "terraform ")
	var args []string
	switch action {
	case "init-upgrade":
		action = "init"
		args = []string{"-upgrade"}
	case "apply-refresh":
		action = "apply"
		args = []string{"-refresh-only"}
	}
	return NewTerraformStrategy(env, action, args)
}
