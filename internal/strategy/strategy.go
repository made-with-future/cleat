package strategy

import (
	"fmt"
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
	"github.com/madewithfuture/cleat/internal/task"
)

// Strategy defines how to execute a command
type Strategy interface {
	// Name returns the command name (e.g., "build", "run")
	Name() string

	// Tasks returns all tasks this strategy may execute
	Tasks() []task.Task

	// Execute runs the strategy with dependency resolution
	Execute(sess *session.Session) error

	// ResolveTasks returns the list of tasks to be executed in order
	ResolveTasks(sess *session.Session) ([]task.Task, error)
}

// CommandProvider handles the mapping of command strings to strategies
type CommandProvider interface {
	// CanHandle returns true if this provider can resolve the given command
	CanHandle(command string) bool
	// GetStrategy returns the appropriate strategy for the command
	GetStrategy(command string, sess *session.Session) Strategy
}

// RegistryProvider handles strategies registered via the global Registry
type RegistryProvider struct{}

func (p *RegistryProvider) CanHandle(command string) bool {
	_, ok := Registry[command]
	return ok
}

func (p *RegistryProvider) GetStrategy(command string, sess *session.Session) Strategy {
	constructor, ok := Registry[command]
	if !ok {
		return nil
	}
	var cfg *config.Config
	if sess != nil {
		cfg = sess.Config
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

func (s *BaseStrategy) ResolveTasks(sess *session.Session) ([]task.Task, error) {
	return s.buildExecutionPlan(sess)
}

// Execute runs tasks in dependency order
func (s *BaseStrategy) Execute(sess *session.Session) error {
	// Build execution plan respecting dependencies
	plan, err := s.buildExecutionPlan(sess)
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
		for _, req := range t.Requirements(sess) {
			requirements[req.Key] = req
		}
	}

	// Prompt for missing inputs
	for key, req := range requirements {
		if _, ok := sess.Inputs[key]; !ok {
			val, err := sess.Exec.Prompt(req.Prompt, req.Default)
			if err != nil {
				return fmt.Errorf("failed to get input for %s: %w", key, err)
			}
			sess.Inputs[key] = val
		}
	}

	// Execute tasks
	for _, t := range plan {
		if err := t.Run(sess); err != nil {
			return fmt.Errorf("task '%s' failed: %w", t.Name(), err)
		}
	}

	fmt.Printf("==> %s completed successfully\n", s.name)
	return nil
}

// buildExecutionPlan returns tasks in dependency order, filtering by ShouldRun
func (s *BaseStrategy) buildExecutionPlan(sess *session.Session) ([]task.Task, error) {
	// Build lookup map
	taskMap := make(map[string]task.Task)
	for _, t := range s.tasks {
		taskMap[t.Name()] = t
	}

	// Filter to tasks that should run
	var candidates []task.Task
	for _, t := range s.tasks {
		if t.ShouldRun(sess) {
			candidates = append(candidates, t)
		}
	}

	// Topological sort for dependency order
	return topologicalSort(candidates, taskMap, sess)
}

// topologicalSort orders tasks respecting dependencies
func topologicalSort(tasks []task.Task, allTasks map[string]task.Task, sess *session.Session) ([]task.Task, error) {
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
			if depTask, exists := allTasks[dep]; exists && depTask.ShouldRun(sess) {
				inDegree[name]++
				dependents[dep] = append(dependents[dep], name)
				// Ensure dependency is in our needed set
				needed[dep] = true
			}
		}
	}

	// Add any dependencies we discovered that weren't in original candidates
	for name := range needed {
		if t, exists := allTasks[name]; exists && t.ShouldRun(sess) {
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
func ResolveCommandTasks(command string, sess *session.Session) ([]task.Task, error) {
	s := GetStrategyForCommand(command, sess)
	if s == nil {
		return nil, fmt.Errorf("unknown command: %s", command)
	}
	return s.ResolveTasks(sess)
}

func GetStrategyForCommand(command string, sess *session.Session) Strategy {
	for _, p := range GetProviders() {
		if p.CanHandle(command) {
			if s := p.GetStrategy(command, sess); s != nil {
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

func (p *NpmProvider) GetStrategy(command string, sess *session.Session) Strategy {
	if sess == nil {
		return nil
	}

	if strings.HasPrefix(command, "npm install") {
		return GetNpmInstallStrategy(command, sess.Config)
	}

	if strings.HasPrefix(command, "npm run ") {
		fullScript := strings.TrimPrefix(command, "npm run ")

		// 1. Try to match as svcName:script first
		if colonIdx := strings.Index(fullScript, ":"); colonIdx != -1 {
			svcName := fullScript[:colonIdx]
			script := fullScript[colonIdx+1:]

			for i := range sess.Config.Services {
				if sess.Config.Services[i].Name == svcName {
					svc := &sess.Config.Services[i]
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
		for i := range sess.Config.Services {
			svc := &sess.Config.Services[i]
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
		for i := range sess.Config.Services {
			svc := &sess.Config.Services[i]
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

func (p *DockerProvider) GetStrategy(command string, sess *session.Session) Strategy {
	if sess == nil {
		return nil
	}

	if command == "docker down" {
		return NewDockerDownStrategy(sess.Config)
	}
	if command == "docker rebuild" {
		return NewDockerRebuildStrategy(sess.Config)
	}
	if command == "docker remove-orphans" {
		return NewDockerRemoveOrphansStrategy(sess.Config)
	}

	parts := strings.Split(command, ":")
	if len(parts) == 2 {
		baseCmd := parts[0]
		svcName := parts[1]
		var targetSvc *config.ServiceConfig
		for i := range sess.Config.Services {
			if sess.Config.Services[i].Name == svcName {
				targetSvc = &sess.Config.Services[i]
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

func (p *DjangoProvider) GetStrategy(command string, sess *session.Session) Strategy {
	if sess == nil {
		return nil
	}

	if command == "django runserver" {
		return NewDjangoRunServerStrategyGlobal(sess.Config)
	}
	if command == "django migrate" {
		return NewDjangoMigrateStrategyGlobal(sess.Config)
	}
	if command == "django makemigrations" {
		return NewDjangoMakeMigrationsStrategyGlobal(sess.Config)
	}
	if command == "django collectstatic" {
		return NewDjangoCollectStaticStrategyGlobal(sess.Config)
	}
	if command == "django create-user-dev" {
		return NewDjangoCreateUserDevStrategyGlobal(sess.Config)
	}
	if command == "django gen-random-secret-key" {
		return NewDjangoGenRandomSecretKeyStrategyGlobal(sess.Config)
	}

	parts := strings.Split(command, ":")
	if len(parts) == 2 {
		baseCmd := parts[0]
		svcName := parts[1]
		var targetSvc *config.ServiceConfig
		for i := range sess.Config.Services {
			if sess.Config.Services[i].Name == svcName {
				targetSvc = &sess.Config.Services[i]
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

func (p *GcpProvider) GetStrategy(command string, sess *session.Session) Strategy {
	if sess == nil {
		return nil
	}

	if strings.HasPrefix(command, "gcp app-engine deploy") {
		parts := strings.Split(command, ":")
		if len(parts) == 2 {
			svcName := parts[1]
			for i := range sess.Config.Services {
				if sess.Config.Services[i].Name == svcName {
					if sess.Config.Services[i].AppYaml != "" {
						return NewGCPAppEngineDeployStrategy(sess.Config.Services[i].AppYaml)
					}
				}
			}
		} else {
			if sess.Config.AppYaml != "" {
				return NewGCPAppEngineDeployStrategy(sess.Config.AppYaml)
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

func (p *TerraformProvider) GetStrategy(command string, sess *session.Session) Strategy {
	if sess == nil {
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
