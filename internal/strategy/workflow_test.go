package strategy

import (
	"errors"
	"strings"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
)

// refined mockWorkflowExecutor
type mockWorkflowExecutor struct {
	executedCommands []string // stores "cmd arg1 arg2"
	failOnCommand    string   // command substring to trigger failure
}

func (m *mockWorkflowExecutor) Run(name string, args ...string) error {
	fullCmd := name
	if len(args) > 0 {
		fullCmd += " " + strings.Join(args, " ")
	}
	m.executedCommands = append(m.executedCommands, fullCmd)

	if m.failOnCommand != "" && strings.Contains(fullCmd, m.failOnCommand) {
		return errors.New("command execution failed")
	}
	return nil
}

func (m *mockWorkflowExecutor) RunWithDir(dir string, name string, args ...string) error {
	return m.Run(name, args...)
}

func (m *mockWorkflowExecutor) Prompt(message string, defaultValue string) (string, error) {
	return defaultValue, nil
}

func TestWorkflowSequentialExecution(t *testing.T) {
	mockExec := &mockWorkflowExecutor{}
	cfg := &config.Config{
		Workflows: []config.Workflow{
			{
				Name:     "seq-test",
				Commands: []string{"echo step1", "echo step2"},
			},
		},
	}
	sess := session.NewSession(cfg, mockExec)

	provider := &WorkflowProvider{}
	strat := provider.GetStrategy("workflow:seq-test", sess)
	if strat == nil {
		t.Fatal("Expected strategy to be returned")
	}

	err := strat.Execute(sess)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	// Verify order
	if len(mockExec.executedCommands) != 2 {
		t.Fatalf("Expected 2 commands, got %d", len(mockExec.executedCommands))
	}
	if mockExec.executedCommands[0] != "echo step1" {
		t.Errorf("Expected first command 'echo step1', got '%s'", mockExec.executedCommands[0])
	}
	if mockExec.executedCommands[1] != "echo step2" {
		t.Errorf("Expected second command 'echo step2', got '%s'", mockExec.executedCommands[1])
	}
}

func TestWorkflowFailFast(t *testing.T) {
	mockExec := &mockWorkflowExecutor{
		failOnCommand: "step1",
	}
	cfg := &config.Config{
		Workflows: []config.Workflow{
			{
				Name:     "fail-test",
				Commands: []string{"echo step1", "echo step2"},
			},
		},
	}
	sess := session.NewSession(cfg, mockExec)

	provider := &WorkflowProvider{}
	strat := provider.GetStrategy("workflow:fail-test", sess)
	
	err := strat.Execute(sess)
	if err == nil {
		t.Fatal("Expected workflow to fail, but it succeeded")
	}

	// Should have executed step1 and stopped
	if len(mockExec.executedCommands) != 1 {
		t.Errorf("Expected 1 command executed, got %d", len(mockExec.executedCommands))
	}
	if !strings.Contains(mockExec.executedCommands[0], "step1") {
		t.Errorf("Expected 'echo step1' to run, got '%s'", mockExec.executedCommands[0])
	}
}
