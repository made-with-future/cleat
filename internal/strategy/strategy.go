package strategy

import (
	"fmt"
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/logger"
	"github.com/madewithfuture/cleat/internal/session"
	"github.com/madewithfuture/cleat/internal/task"
)

// Strategy defines how to execute a command
type Strategy interface {
	// Name returns the command name (e.g., "build", "run")
	Name() string

	// Tasks returns all tasks this strategy may execute.
	// NOTE: For dynamic strategies like workflows, this may return nil.
	// Callers should prefer ResolveTasks(session) when a session is available
	// to get the actual list of tasks to be executed.
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
	logger.Info("executing strategy", map[string]interface{}{"strategy": s.name})

	// Build execution plan respecting dependencies
	plan, err := s.buildExecutionPlan(sess)
	if err != nil {
		logger.Error("failed to build execution plan", err, map[string]interface{}{"strategy": s.name})
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
		logger.Debug("running task", map[string]interface{}{"task": t.Name()})
		if err := t.Run(sess); err != nil {
			logger.Error("task execution failed", err, map[string]interface{}{"task": t.Name(), "strategy": s.name})
			return fmt.Errorf("task '%s' failed: %w", t.Name(), err)
		}
	}

	logger.Info("strategy completed successfully", map[string]interface{}{"strategy": s.name})
	task.PrintStep(fmt.Sprintf("%s completed successfully", s.name))
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
	logger.Debug("resolving command to strategy", map[string]interface{}{"command": command})
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
		&WorkflowProvider{},
		&NpmProvider{},
		&DockerProvider{},
		&DjangoProvider{},
		&GcpProvider{},
		&TerraformProvider{},
		&RegistryProvider{},
		&PassthroughProvider{},
	}
}

// PassthroughProvider handles commands that don't match any other provider
// It assumes the command is a direct shell command
type PassthroughProvider struct{}

func (p *PassthroughProvider) CanHandle(command string) bool {
	// Don't handle internal workflow commands that failed to match a known workflow
	if strings.HasPrefix(command, "workflow:") {
		return false
	}
	return true // Can handle any other command not caught by others
}

func (p *PassthroughProvider) GetStrategy(command string, sess *session.Session) Strategy {
	return NewPassthroughStrategy(command)
}

// PassthroughStrategy directly executes a shell command
type PassthroughStrategy struct {
	command string
}

func NewPassthroughStrategy(command string) *PassthroughStrategy {
	return &PassthroughStrategy{command: command}
}

func (s *PassthroughStrategy) Name() string {
	return "passthrough:" + s.command
}

func (s *PassthroughStrategy) Tasks() []task.Task {
	// A passthrough strategy generates a single shell task
	return []task.Task{task.NewShellTask(s.command)}
}

func (s *PassthroughStrategy) ResolveTasks(sess *session.Session) ([]task.Task, error) {
	// For a passthrough strategy, the task list is simply the shell task itself.
	// No complex dependencies or sub-resolutions here.
	return s.Tasks(), nil
}

func (s *PassthroughStrategy) Execute(sess *session.Session) error {
	logger.Info("executing passthrough strategy", map[string]interface{}{"command": s.command})

	// Delegate to the single task
	shellTask := task.NewShellTask(s.command)
	if err := shellTask.Run(sess); err != nil {
		return fmt.Errorf("passthrough command '%s' failed: %w", s.command, err)
	}
	task.PrintStep(fmt.Sprintf("Command '%s' completed successfully", s.command))
	return nil
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
	return strings.HasPrefix(command, "docker ") || strings.HasPrefix(command, "docker:")
}

func (p *DockerProvider) GetStrategy(command string, sess *session.Session) Strategy {
	if sess == nil {
		return nil
	}

	if command == "docker down" || command == "docker:down" {
		return NewDockerDownStrategy(sess.Config)
	}
	if command == "docker rebuild" || command == "docker:rebuild" {
		return NewDockerRebuildStrategy(sess.Config)
	}
	if command == "docker remove-orphans" || command == "docker:remove-orphans" {
		return NewDockerRemoveOrphansStrategy(sess.Config)
	}
	if command == "docker up" || command == "docker:up" {
		return NewDockerUpStrategy(sess.Config)
	}

	// Handle service-specific commands: "docker <cmd>:<svc>" or "docker:<cmd>:<svc>"
	fullCmd := command
	if strings.HasPrefix(fullCmd, "docker ") {
		// Convert "docker down:web" to "docker:down:web" for easier splitting
		fullCmd = "docker:" + strings.TrimPrefix(fullCmd, "docker ")
	}

	parts := strings.Split(fullCmd, ":")
	if len(parts) == 3 {
		// docker:<baseCmd>:<svcName>
		baseCmd := parts[1]
		svcName := parts[2]
		var targetSvc *config.ServiceConfig
		for i := range sess.Config.Services {
			if sess.Config.Services[i].Name == svcName {
				targetSvc = &sess.Config.Services[i]
				break
			}
		}

		if targetSvc != nil {
			switch baseCmd {
			case "down":
				return NewDockerDownStrategyForService(targetSvc)
			case "rebuild":
				return NewDockerRebuildStrategyForService(targetSvc)
			case "remove-orphans":
				return NewDockerRemoveOrphansStrategyForService(targetSvc)
			case "up":
				return NewDockerUpStrategyForService(targetSvc)
			}
		}
	} else if len(parts) == 2 {
		// Cases like "docker:up" (handled above) or if it was "docker down:web" but now it's "docker:down:web"
		// Wait, if it was "docker down:web", strings.Split(normalized, ":") gives ["docker", "down", "web"]?
		// No, strings.Replace("docker down:web", " ", ":", 1) gives "docker:down:web".
		// Let's re-verify my normalization.
	}
	return nil
}

// DjangoProvider handles service-specific django commands
type DjangoProvider struct{}

func (p *DjangoProvider) CanHandle(command string) bool {
	return strings.HasPrefix(command, "django ") || strings.HasPrefix(command, "django:")
}

func (p *DjangoProvider) GetStrategy(command string, sess *session.Session) Strategy {
	if sess == nil {
		return nil
	}

	// Normalize: "django migrate" -> "django:migrate"
	normalized := strings.Replace(command, " ", ":", 1)
	parts := strings.Split(normalized, ":")

	if len(parts) == 2 {
		baseCmd := parts[1]
		switch baseCmd {
		case "runserver":
			return NewDjangoRunServerStrategyGlobal(sess.Config)
		case "migrate":
			return NewDjangoMigrateStrategyGlobal(sess.Config)
		case "makemigrations":
			return NewDjangoMakeMigrationsStrategyGlobal(sess.Config)
		case "collectstatic":
			return NewDjangoCollectStaticStrategyGlobal(sess.Config)
		case "create-user-dev":
			return NewDjangoCreateUserDevStrategyGlobal(sess.Config)
		case "gen-random-secret-key":
			return NewDjangoGenRandomSecretKeyStrategyGlobal(sess.Config)
		}
	}

	if len(parts) == 3 {
		baseCmd := parts[1]
		svcName := parts[2]
		var targetSvc *config.ServiceConfig
		for i := range sess.Config.Services {
			if sess.Config.Services[i].Name == svcName {
				targetSvc = &sess.Config.Services[i]
				break
			}
		}

		if targetSvc != nil {
			switch baseCmd {
			case "runserver":
				return NewDjangoRunServerStrategy(targetSvc)
			case "migrate":
				return NewDjangoMigrateStrategy(targetSvc)
			case "makemigrations":
				return NewDjangoMakeMigrationsStrategy(targetSvc)
			case "collectstatic":
				return NewDjangoCollectStaticStrategy(targetSvc)
			case "create-user-dev":
				return NewDjangoCreateUserDevStrategy(targetSvc)
			case "gen-random-secret-key":
				return NewDjangoGenRandomSecretKeyStrategy(targetSvc)
			}
		}
	}
	return nil
}

// GcpProvider handles Google Cloud Platform commands
type GcpProvider struct{}

func (p *GcpProvider) CanHandle(command string) bool {
	return strings.HasPrefix(command, "gcp ") || strings.HasPrefix(command, "gcp:")
}

func (p *GcpProvider) GetStrategy(command string, sess *session.Session) Strategy {
	if sess == nil {
		return nil
	}

	// Normalize: "gcp app-engine deploy" -> "gcp:app-engine:deploy"
	// Actually GCP commands might have multiple spaces.
	// Let's look at existing logic.

	if strings.HasPrefix(command, "gcp app-engine deploy") || strings.HasPrefix(command, "gcp:app-engine:deploy") {
		// handle deploy
		fullCmd := strings.Replace(command, " ", ":", -1)
		parts := strings.Split(fullCmd, ":")
		// gcp:app-engine:deploy[:svc]
		if len(parts) == 4 {
			svcName := parts[3]
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

	if strings.HasPrefix(command, "gcp app-engine promote") || strings.HasPrefix(command, "gcp:app-engine:promote") {
		fullCmd := strings.Replace(command, " ", ":", -1)
		parts := strings.Split(fullCmd, ":")
		// gcp:app-engine:promote[:svc]
		if len(parts) == 4 {
			return NewGCPAppEnginePromoteStrategy(parts[3])
		} else {
			return NewGCPAppEnginePromoteStrategy("")
		}
	}

	if command == "gcp console" || command == "gcp:console" {
		return NewGCPConsoleStrategy()
	}

	if command == "gcp init" || command == "gcp:init" {
		return NewGCPInitStrategy()
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
