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

// MockExecutor for use in other package tests
type MockExecutor struct {
	Commands []struct {
		Name string
		Args []string
	}
	Error error
}

func (m *MockExecutor) Run(name string, args ...string) error {
	m.Commands = append(m.Commands, struct {
		Name string
		Args []string
	}{Name: name, Args: args})
	return m.Error
}
