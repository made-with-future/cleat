package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/session"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/madewithfuture/cleat/internal/task"
)

// FixturePath returns the absolute path to a fixture directory
func FixturePath(name string) string {
	// Get the path relative to the testdata directory
	wd, _ := os.Getwd()
	// If we're already in testdata, use current dir, otherwise go up
	if filepath.Base(wd) == "testdata" {
		return filepath.Join(wd, "fixtures", name)
	}
	return filepath.Join(wd, "testdata", "fixtures", name)
}

// LoadFixture loads a fixture's configuration and returns it
func LoadFixture(t *testing.T, name string) *config.Config {
	t.Helper()
	fixturePath := FixturePath(name)

	// Change to fixture directory
	oldWd, _ := os.Getwd()
	err := os.Chdir(fixturePath)
	if err != nil {
		t.Fatalf("failed to change to fixture directory %s: %v", fixturePath, err)
	}
	defer os.Chdir(oldWd)

	// Load config
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		t.Fatalf("failed to load fixture config for %s: %v", name, err)
	}

	return cfg
}

// LoadFixtureFromPath loads a configuration from a specific path
func LoadFixtureFromPath(t *testing.T, fixturePath string) *config.Config {
	t.Helper()

	oldWd, _ := os.Getwd()
	err := os.Chdir(fixturePath)
	if err != nil {
		t.Fatalf("failed to change to fixture directory %s: %v", fixturePath, err)
	}
	defer os.Chdir(oldWd)

	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		t.Fatalf("failed to load config from %s: %v", fixturePath, err)
	}

	return cfg
}

// MockExecutor is a test executor that records commands without executing them
type MockExecutor struct {
	Commands        []ExecutedCommand
	PromptResponses map[string]string
	Error           error
}

// ExecutedCommand represents a command that was executed
type ExecutedCommand struct {
	Name string
	Args []string
	Dir  string
}

func (m *MockExecutor) Run(name string, args ...string) error {
	m.Commands = append(m.Commands, ExecutedCommand{Name: name, Args: args})
	return m.Error
}

func (m *MockExecutor) RunWithDir(dir string, name string, args ...string) error {
	m.Commands = append(m.Commands, ExecutedCommand{Name: name, Args: args, Dir: dir})
	return m.Error
}

func (m *MockExecutor) Prompt(message string, defaultValue string) (string, error) {
	if m.PromptResponses != nil {
		if resp, ok := m.PromptResponses[message]; ok {
			return resp, nil
		}
	}
	return defaultValue, nil
}

// Verify interface compliance
var _ executor.Executor = &MockExecutor{}

// ExecuteWithMock executes a command against a fixture using a mock executor
func ExecuteWithMock(t *testing.T, fixtureName string, command string) (*MockExecutor, error) {
	t.Helper()

	cfg := LoadFixture(t, fixtureName)
	mock := &MockExecutor{
		Commands:        []ExecutedCommand{},
		PromptResponses: make(map[string]string),
	}
	sess := session.NewSession(cfg, mock)

	strat := strategy.GetStrategyForCommand(command, sess)
	if strat == nil {
		t.Fatalf("no strategy found for command: %s", command)
	}

	err := strat.Execute(sess)
	return mock, err
}

// AssertTaskNames checks that the task names match the expected list
func AssertTaskNames(t *testing.T, tasks []task.Task, expected []string) {
	t.Helper()

	if len(tasks) != len(expected) {
		var got []string
		for _, task := range tasks {
			got = append(got, task.Name())
		}
		t.Errorf("task count mismatch:\n  got:      %v\n  expected: %v", got, expected)
		return
	}

	for i, task := range tasks {
		if task.Name() != expected[i] {
			t.Errorf("task[%d] name mismatch: got %q, expected %q", i, task.Name(), expected[i])
		}
	}
}

// AssertCommandExecuted checks if a specific command was executed
func AssertCommandExecuted(t *testing.T, mock *MockExecutor, commandName string) {
	t.Helper()

	for _, cmd := range mock.Commands {
		if cmd.Name == commandName {
			return
		}
	}

	var executed []string
	for _, cmd := range mock.Commands {
		executed = append(executed, cmd.Name)
	}
	t.Errorf("command %q was not executed. Executed commands: %v", commandName, executed)
}

// AssertServiceExists checks if a service exists in the config
func AssertServiceExists(t *testing.T, cfg *config.Config, serviceName string) {
	t.Helper()

	for _, svc := range cfg.Services {
		if svc.Name == serviceName {
			return
		}
	}

	var services []string
	for _, svc := range cfg.Services {
		services = append(services, svc.Name)
	}
	t.Errorf("service %q not found. Available services: %v", serviceName, services)
}

// AssertModuleExists checks if a module type exists in a service
func AssertModuleExists(t *testing.T, cfg *config.Config, serviceName string, checkFunc func(config.ModuleConfig) bool) {
	t.Helper()

	for _, svc := range cfg.Services {
		if svc.Name == serviceName {
			for _, mod := range svc.Modules {
				if checkFunc(mod) {
					return
				}
			}
			t.Errorf("module not found in service %q", serviceName)
			return
		}
	}

	t.Errorf("service %q not found", serviceName)
}

// CopyFixture copies a fixture to a temporary directory for tests that need to modify files
func CopyFixture(t *testing.T, name string) string {
	t.Helper()

	srcPath := FixturePath(name)
	tmpDir, err := os.MkdirTemp("", "cleat-fixture-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	err = copyDir(srcPath, tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to copy fixture: %v", err)
	}

	return tmpDir
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}
