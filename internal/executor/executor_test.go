package executor

import (
	"testing"
)

func TestShellExecutorInterface(t *testing.T) {
	// Verify ShellExecutor implements Executor interface
	var _ Executor = &ShellExecutor{}
}

func TestDefaultExecutor(t *testing.T) {
	if Default == nil {
		t.Error("expected Default executor to be non-nil")
	}
}

func TestShellExecutorRun(t *testing.T) {
	exec := &ShellExecutor{}
	// Test that Run calls RunWithDir with empty dir
	// We'll use echo which is available on all platforms
	err := exec.Run("echo", "test")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestShellExecutorRunWithDir(t *testing.T) {
	exec := &ShellExecutor{}
	// Test with a valid directory
	err := exec.RunWithDir(".", "echo", "test")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// MockExecutor for use in other package tests
type MockExecutor struct {
	Commands []struct {
		Dir  string
		Name string
		Args []string
	}
	Prompts       []string
	PromptInputs  []string
	PromptCounter int
	Error         error
}

func (m *MockExecutor) Run(name string, args ...string) error {
	return m.RunWithDir("", name, args...)
}

func (m *MockExecutor) RunWithDir(dir string, name string, args ...string) error {
	m.Commands = append(m.Commands, struct {
		Dir  string
		Name string
		Args []string
	}{Dir: dir, Name: name, Args: args})
	return m.Error
}

func (m *MockExecutor) Prompt(message string, defaultValue string) (string, error) {
	m.Prompts = append(m.Prompts, message)
	if m.Error != nil {
		return "", m.Error
	}
	if m.PromptCounter < len(m.PromptInputs) {
		result := m.PromptInputs[m.PromptCounter]
		m.PromptCounter++
		if result == "" {
			return defaultValue, nil
		}
		return result, nil
	}
	return defaultValue, nil
}

func TestMockExecutor(t *testing.T) {
	mock := &MockExecutor{}

	// Test Run
	err := mock.Run("test", "arg1", "arg2")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(mock.Commands) != 1 {
		t.Errorf("expected 1 command, got %d", len(mock.Commands))
	}
	if mock.Commands[0].Name != "test" {
		t.Errorf("expected command name 'test', got %q", mock.Commands[0].Name)
	}

	// Test RunWithDir
	err = mock.RunWithDir("/tmp", "test2", "arg1")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(mock.Commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(mock.Commands))
	}
	if mock.Commands[1].Dir != "/tmp" {
		t.Errorf("expected dir '/tmp', got %q", mock.Commands[1].Dir)
	}

	// Test Prompt with default
	mock.PromptInputs = []string{""}
	result, err := mock.Prompt("Enter value", "default")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "default" {
		t.Errorf("expected 'default', got %q", result)
	}

	// Test Prompt with input
	mock.PromptCounter = 0
	mock.PromptInputs = []string{"custom"}
	result, err = mock.Prompt("Enter value", "default")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "custom" {
		t.Errorf("expected 'custom', got %q", result)
	}
}
