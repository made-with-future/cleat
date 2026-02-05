package strategy

import (
	"fmt"
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

func (s *WorkflowStrategy) Tasks() []task.Task {
	// This is a bit tricky because WorkflowStrategy doesn't own the tasks directly,
	// it delegates to other strategies. For the sake of the interface, we'll
	// return an empty slice here as ResolveTasks will be used for execution.
	return nil
}

func (s *WorkflowStrategy) ResolveTasks(sess *session.Session) ([]task.Task, error) {
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

	for _, cmd := range s.commands {
		logger.Debug("workflow step", map[string]interface{}{"workflow": s.name, "command": cmd})
		strat := GetStrategyForCommand(cmd, sess)
		if strat == nil {
			return fmt.Errorf("workflow '%s' step '%s' failed: unknown command", s.name, cmd)
		}

		if err := strat.Execute(sess); err != nil {
			return fmt.Errorf("workflow '%s' step '%s' failed: %w", s.name, cmd, err)
		}
	}

	fmt.Printf("==> Workflow '%s' completed successfully\n", s.name)
	return nil
}
