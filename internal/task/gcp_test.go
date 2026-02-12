package task

import (
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/session"
)

type mockExecutor struct {
	executor.ShellExecutor
	commands [][]string
}

func (m *mockExecutor) Run(name string, args ...string) error {
	m.commands = append(m.commands, append([]string{name}, args...))
	return nil
}

func (m *mockExecutor) RunWithDir(dir string, name string, args ...string) error {
	m.commands = append(m.commands, append([]string{name}, args...))
	return nil
}

func TestGCPInit(t *testing.T) {
	cfg := &config.Config{
		GoogleCloudPlatform: &config.GCPConfig{
			ProjectName: "test-project",
			Account:     "test@example.com",
		},
	}
	mock := &mockExecutor{}
	sess := session.NewSession(cfg, mock)

	task := NewGCPInit()
	if !task.ShouldRun(sess) {
		t.Fatal("ShouldRun should be true")
	}

	err := task.Run(sess)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	expected := [][]string{
		{"gcloud", "config", "set", "project", "test-project"},
		{"gcloud", "config", "set", "account", "test@example.com"},
	}

	if len(mock.commands) != len(expected) {
		t.Fatalf("expected %d commands, got %d", len(expected), len(mock.commands))
	}

	for i, cmd := range mock.commands {
		for j, arg := range cmd {
			if arg != expected[i][j] {
				t.Errorf("expected arg %d of command %d to be %s, got %s", j, i, expected[i][j], arg)
			}
		}
	}
}

func TestGCPCreateProject(t *testing.T) {
	cfg := &config.Config{
		GoogleCloudPlatform: &config.GCPConfig{
			ProjectName: "test-project",
		},
	}
	mock := &mockExecutor{}
	sess := session.NewSession(cfg, mock)

	task := NewGCPCreateProject()
	if !task.ShouldRun(sess) {
		t.Fatal("ShouldRun should be true")
	}

	err := task.Run(sess)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	expected := [][]string{
		{"gcloud", "projects", "create", "test-project"},
	}

	if len(mock.commands) != len(expected) {
		t.Fatalf("expected %d commands, got %d", len(expected), len(mock.commands))
	}

	for i, cmd := range mock.commands {
		for j, arg := range cmd {
			if arg != expected[i][j] {
				t.Errorf("expected arg %d of command %d to be %s, got %s", j, i, expected[i][j], arg)
			}
		}
	}
}
