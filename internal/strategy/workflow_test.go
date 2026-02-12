package strategy

import (
	"errors"
	"strings"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
	"github.com/madewithfuture/cleat/internal/task"
)

// refined mockWorkflowExecutor
type mockWorkflowExecutor struct {
	executedCommands []string // stores "cmd arg1 arg2"
	prompts          []string // stores prompt messages
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
	m.prompts = append(m.prompts, message)
	// Simple simulation: return a value if "Prompt" matches, else default
	if strings.Contains(message, "Enter value") {
		return "user-input", nil
	}
	return defaultValue, nil
}

// ... existing tests ...

func TestWorkflowPromptOrder(t *testing.T) {
	// Register tasks with inputs Z, A, M
	reqTaskZ := &mockTaskWithReqs{name: "Z", reqs: []task.InputRequirement{{Key: "z", Prompt: "Prompt Z"}}}
	reqTaskA := &mockTaskWithReqs{name: "A", reqs: []task.InputRequirement{{Key: "a", Prompt: "Prompt A"}}}
	reqTaskM := &mockTaskWithReqs{name: "M", reqs: []task.InputRequirement{{Key: "m", Prompt: "Prompt M"}}}

	Register("cmd-z", func(cfg *config.Config) Strategy { return NewBaseStrategy("cmd-z", []task.Task{reqTaskZ}) })
	Register("cmd-a", func(cfg *config.Config) Strategy { return NewBaseStrategy("cmd-a", []task.Task{reqTaskA}) })
	Register("cmd-m", func(cfg *config.Config) Strategy { return NewBaseStrategy("cmd-m", []task.Task{reqTaskM}) })
	defer func() { delete(Registry, "cmd-z"); delete(Registry, "cmd-a"); delete(Registry, "cmd-m") }()

	mockExec := &mockWorkflowExecutor{}
	cfg := &config.Config{
		Workflows: []config.Workflow{
			{Name: "order-test", Commands: []string{"cmd-z", "cmd-a", "cmd-m"}},
		},
	}
	sess := session.NewSession(cfg, mockExec)
	sess.Inputs = make(map[string]string) // Ensure empty so it prompts

	provider := &WorkflowProvider{}
	strat := provider.GetStrategy("workflow:order-test", sess)
	strat.Execute(sess)

	// Check order. Should be A, M, Z
	if len(mockExec.prompts) != 3 {
		t.Fatalf("Expected 3 prompts, got %d", len(mockExec.prompts))
	}
	if mockExec.prompts[0] != "Prompt A" {
		t.Errorf("Expected first prompt 'Prompt A', got '%s'", mockExec.prompts[0])
	}
	if mockExec.prompts[1] != "Prompt M" {
		t.Errorf("Expected second prompt 'Prompt M', got '%s'", mockExec.prompts[1])
	}
	if mockExec.prompts[2] != "Prompt Z" {
		t.Errorf("Expected third prompt 'Prompt Z', got '%s'", mockExec.prompts[2])
	}
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

func TestWorkflowErrorMessage(t *testing.T) {
	mockExec := &mockWorkflowExecutor{
		failOnCommand: "step1",
	}
	cfg := &config.Config{
		Workflows: []config.Workflow{
			{
				Name:     "err-test",
				Commands: []string{"echo step1"},
			},
		},
	}
	sess := session.NewSession(cfg, mockExec)

	provider := &WorkflowProvider{}
	strat := provider.GetStrategy("workflow:err-test", sess)

	err := strat.Execute(sess)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify error message structure
	// Expected: "workflow 'err-test' task 'shell:echo' failed: ..."
	expectedPart := "workflow 'err-test' task"
	if !strings.Contains(err.Error(), expectedPart) {
		t.Errorf("Error message '%s' does not contain expected context '%s'", err.Error(), expectedPart)
	}
}

// mockTaskWithReqs
type mockTaskWithReqs struct {
	name string
	reqs []task.InputRequirement
}

func (t *mockTaskWithReqs) Name() string                              { return t.name }
func (t *mockTaskWithReqs) Description() string                       { return "mock task" }
func (t *mockTaskWithReqs) Dependencies() []string                    { return nil }
func (t *mockTaskWithReqs) ShouldRun(sess *session.Session) bool      { return true }
func (t *mockTaskWithReqs) Run(sess *session.Session) error           { return nil }
func (t *mockTaskWithReqs) Commands(sess *session.Session) [][]string { return nil }
func (t *mockTaskWithReqs) Requirements(sess *session.Session) []task.InputRequirement {
	return t.reqs
}

func TestWorkflowInputPrompting(t *testing.T) {
	reqTask := &mockTaskWithReqs{
		name: "req-task",
		reqs: []task.InputRequirement{
			{Key: "test:input", Prompt: "Enter value", Default: "default"},
		},
	}

	Register("test-req", func(cfg *config.Config) Strategy {
		return NewBaseStrategy("test-req", []task.Task{reqTask})
	})
	// Cleanup registry after test
	defer func() {
		delete(Registry, "test-req")
	}()

	mockExec := &mockWorkflowExecutor{}
	cfg := &config.Config{
		Workflows: []config.Workflow{
			{
				Name:     "prompt-test",
				Commands: []string{"test-req"},
			},
		},
	}
	sess := session.NewSession(cfg, mockExec)

	provider := &WorkflowProvider{}
	strat := provider.GetStrategy("workflow:prompt-test", sess)

	// Clear inputs to force prompt
	sess.Inputs = make(map[string]string)

	err := strat.Execute(sess)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Verify input was collected
	if val, ok := sess.Inputs["test:input"]; !ok || val != "user-input" {
		t.Errorf("Expected input 'user-input', got '%s' (ok=%v)", val, ok)
	}
}

func TestUnknownWorkflowError(t *testing.T) {
	mockExec := &mockWorkflowExecutor{}
	cfg := &config.Config{
		Workflows: []config.Workflow{},
	}
	sess := session.NewSession(cfg, mockExec)

	// Verify GetStrategyForCommand returns nil for unknown workflow
	strat := GetStrategyForCommand("workflow:unknown", sess)
	if strat != nil {
		t.Errorf("Expected nil strategy for unknown workflow, got %v", strat)
	}
}

func TestWorkflowRecursionLoop(t *testing.T) {
	mockExec := &mockWorkflowExecutor{}
	cfg := &config.Config{
		Workflows: []config.Workflow{
			{
				Name:     "loop-a",
				Commands: []string{"workflow:loop-a"},
			},
			{
				Name:     "loop-b-1",
				Commands: []string{"workflow:loop-b-2"},
			},
			{
				Name:     "loop-b-2",
				Commands: []string{"workflow:loop-b-1"},
			},
		},
	}
	sess := session.NewSession(cfg, mockExec)

	provider := &WorkflowProvider{}

	t.Run("self-reference", func(t *testing.T) {
		strat := provider.GetStrategy("workflow:loop-a", sess)
		err := strat.Execute(sess)
		if err == nil || !strings.Contains(err.Error(), "cycle detected") {
			t.Errorf("Expected cycle detected error, got %v", err)
		}
	})

	t.Run("mutual-recursion", func(t *testing.T) {
		strat := provider.GetStrategy("workflow:loop-b-1", sess)
		err := strat.Execute(sess)
		if err == nil || !strings.Contains(err.Error(), "cycle detected") {
			t.Errorf("Expected cycle detected error, got %v", err)
		}
	})
}

type SpyStrategy struct {
	BaseStrategy
	resolveCount int
}

func (s *SpyStrategy) ResolveTasks(sess *session.Session) ([]task.Task, error) {
	s.resolveCount++
	return s.BaseStrategy.ResolveTasks(sess)
}

func TestWorkflowDoubleResolution(t *testing.T) {
	spy := &SpyStrategy{
		BaseStrategy: *NewBaseStrategy("spy", nil),
	}

	// Register spy
	Register("spy-cmd", func(cfg *config.Config) Strategy {
		return spy
	})
	defer func() {
		delete(Registry, "spy-cmd")
	}()

	mockExec := &mockWorkflowExecutor{}
	cfg := &config.Config{
		Workflows: []config.Workflow{
			{
				Name:     "spy-wf",
				Commands: []string{"spy-cmd"},
			},
		},
	}
	sess := session.NewSession(cfg, mockExec)

	provider := &WorkflowProvider{}
	strat := provider.GetStrategy("workflow:spy-wf", sess)

	err := strat.Execute(sess)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Currently expected to fail (count will be 2)
	if spy.resolveCount != 1 {
		t.Fatalf("Expected 1 resolution, got %d", spy.resolveCount)
	}
}
