package strategy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/madewithfuture/cleat/internal/logger"
	"github.com/madewithfuture/cleat/internal/session"
	"github.com/madewithfuture/cleat/internal/task"
)

// WorkflowProvider handles named workflows from the configuration
type WorkflowProvider struct{}

func (p *WorkflowProvider) CanHandle(command string) bool {
	return strings.HasPrefix(command, "workflow:")
}

func (p *WorkflowProvider) GetStrategy(command string, sess *session.Session) Strategy {
	if sess == nil {
		return nil
	}

	wfName := strings.TrimPrefix(command, "workflow:")
	for _, wf := range sess.Config.Workflows {
		if wf.Name == wfName {
			return NewWorkflowStrategy(wf.Name, wf.Commands)
		}
	}

	return nil
}

// WorkflowStrategy executes a sequence of commands
type WorkflowStrategy struct {
	name     string
	commands []string
}

func NewWorkflowStrategy(name string, commands []string) *WorkflowStrategy {
	return &WorkflowStrategy{
		name:     name,
		commands: commands,
	}
}

func (s *WorkflowStrategy) Name() string {
	return "workflow:" + s.name
}

// Tasks returns nil for WorkflowStrategy as tasks are resolved dynamically.
// Callers MUST use ResolveTasks(session) to get the task list.
func (s *WorkflowStrategy) Tasks() []task.Task {
	return nil
}

func (s *WorkflowStrategy) ResolveTasks(sess *session.Session) ([]task.Task, error) {
	// 1. Detect cycles
	for _, name := range sess.WorkflowStack {
		if name == s.name {
			return nil, fmt.Errorf("cycle detected: %s -> %s", strings.Join(sess.WorkflowStack, " -> "), s.name)
		}
	}

	// 2. Enforce depth limit (failsafe)
	if len(sess.WorkflowStack) > 50 { // Hard limit of 50 nested workflows
		return nil, fmt.Errorf("max workflow nesting depth (50) exceeded")
	}

	// 3. Push to stack
	sess.WorkflowStack = append(sess.WorkflowStack, s.name)
	defer func() {
		// Pop from stack
		if len(sess.WorkflowStack) > 0 {
			sess.WorkflowStack = sess.WorkflowStack[:len(sess.WorkflowStack)-1]
		}
	}()

	var allTasks []task.Task
	for _, cmd := range s.commands {
		tasks, err := ResolveCommandTasks(cmd, sess)
		if err != nil {
			return nil, fmt.Errorf("workflow '%s' failed to resolve command '%s': %w", s.name, cmd, err)
		}
		allTasks = append(allTasks, tasks...)
	}
	return allTasks, nil
}

func (s *WorkflowStrategy) Execute(sess *session.Session) error {
	logger.Info("executing workflow strategy", map[string]interface{}{"workflow": s.name})

	// 1. Resolve all tasks to collect requirements upfront
	allTasks, err := s.ResolveTasks(sess)
	if err != nil {
		return err
	}

	// 2. Collect requirements from all tasks
	requirements := make(map[string]task.InputRequirement)
	for _, t := range allTasks {
		for _, req := range t.Requirements(sess) {
			requirements[req.Key] = req
		}
	}

	// 3. Prompt for missing inputs
	var keys []string
	for k := range requirements {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		req := requirements[key]
		if _, ok := sess.Inputs[key]; !ok {
			val, err := sess.Exec.Prompt(req.Prompt, req.Default)
			if err != nil {
				return fmt.Errorf("failed to get input for %s: %w", key, err)
			}
			sess.Inputs[key] = val
		}
	}

	// 4. Execute tasks sequentially
	for _, t := range allTasks {
		logger.Debug("running workflow task", map[string]interface{}{"workflow": s.name, "task": t.Name()})
		if err := t.Run(sess); err != nil {
			return fmt.Errorf("workflow '%s' task '%s' failed: %w", s.name, t.Name(), err)
		}
	}

	fmt.Printf("==> Workflow '%s' completed successfully\n", s.name)
	return nil
}
