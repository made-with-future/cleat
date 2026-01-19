package strategy

import (
	"errors"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/task"
)

// mockExecutor for testing
type mockExecutor struct {
	commands []string
	err      error
}

func (m *mockExecutor) Run(name string, args ...string) error {
	m.commands = append(m.commands, name)
	return m.err
}

// mockTask for testing strategy execution
type mockTask struct {
	task.BaseTask
	shouldRun bool
	runCalled bool
	runErr    error
}

func (t *mockTask) ShouldRun(cfg *config.Config) bool {
	return t.shouldRun
}

func (t *mockTask) Run(cfg *config.Config, exec executor.Executor) error {
	t.runCalled = true
	return t.runErr
}

func (t *mockTask) Commands(cfg *config.Config) [][]string {
	return [][]string{{"mock", "command"}}
}

func TestBaseStrategy(t *testing.T) {
	tasks := []task.Task{
		&mockTask{BaseTask: task.BaseTask{TaskName: "task1"}, shouldRun: true},
		&mockTask{BaseTask: task.BaseTask{TaskName: "task2"}, shouldRun: true},
	}

	s := NewBaseStrategy("test", tasks)

	if s.Name() != "test" {
		t.Errorf("expected name 'test', got %q", s.Name())
	}

	if len(s.Tasks()) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(s.Tasks()))
	}
}

func TestExecuteRunsTasks(t *testing.T) {
	task1 := &mockTask{BaseTask: task.BaseTask{TaskName: "task1"}, shouldRun: true}
	task2 := &mockTask{BaseTask: task.BaseTask{TaskName: "task2"}, shouldRun: true}

	s := NewBaseStrategy("test", []task.Task{task1, task2})
	mock := &mockExecutor{}
	cfg := &config.Config{}

	err := s.Execute(cfg, mock)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !task1.runCalled {
		t.Error("expected task1.Run to be called")
	}
	if !task2.runCalled {
		t.Error("expected task2.Run to be called")
	}
}

func TestExecuteSkipsTasksThatShouldNotRun(t *testing.T) {
	task1 := &mockTask{BaseTask: task.BaseTask{TaskName: "task1"}, shouldRun: true}
	task2 := &mockTask{BaseTask: task.BaseTask{TaskName: "task2"}, shouldRun: false}

	s := NewBaseStrategy("test", []task.Task{task1, task2})
	mock := &mockExecutor{}
	cfg := &config.Config{}

	err := s.Execute(cfg, mock)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !task1.runCalled {
		t.Error("expected task1.Run to be called")
	}
	if task2.runCalled {
		t.Error("expected task2.Run NOT to be called")
	}
}

func TestExecuteStopsOnError(t *testing.T) {
	task1 := &mockTask{
		BaseTask:  task.BaseTask{TaskName: "task1"},
		shouldRun: true,
		runErr:    errors.New("task1 failed"),
	}
	task2 := &mockTask{BaseTask: task.BaseTask{TaskName: "task2"}, shouldRun: true}

	s := NewBaseStrategy("test", []task.Task{task1, task2})
	mock := &mockExecutor{}
	cfg := &config.Config{}

	err := s.Execute(cfg, mock)
	if err == nil {
		t.Error("expected error, got nil")
	}

	if !task1.runCalled {
		t.Error("expected task1.Run to be called")
	}
	if task2.runCalled {
		t.Error("expected task2.Run NOT to be called after task1 error")
	}
}

func TestExecuteWithNoApplicableTasks(t *testing.T) {
	task1 := &mockTask{BaseTask: task.BaseTask{TaskName: "task1"}, shouldRun: false}

	s := NewBaseStrategy("test", []task.Task{task1})
	mock := &mockExecutor{}
	cfg := &config.Config{}

	err := s.Execute(cfg, mock)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if task1.runCalled {
		t.Error("expected task1.Run NOT to be called")
	}
}

// ... existing code ...
func TestDependencyOrder(t *testing.T) {
	cfg := &config.Config{}
	executionOrder := []string{}

	// Create custom tasks that track execution order
	orderedTask1 := &orderTrackingTask{
		mockTask: mockTask{BaseTask: task.BaseTask{TaskName: "task1"}, shouldRun: true},
		order:    &executionOrder,
	}
	orderedTask2 := &orderTrackingTask{
		mockTask: mockTask{BaseTask: task.BaseTask{TaskName: "task2", TaskDeps: []string{"task1"}}, shouldRun: true},
		order:    &executionOrder,
	}

	// Add in reverse order to test dependency resolution
	s := NewBaseStrategy("test", []task.Task{orderedTask2, orderedTask1})
	mock := &mockExecutor{}

	err := s.Execute(cfg, mock)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(executionOrder) != 2 {
		t.Fatalf("expected 2 tasks executed, got %d", len(executionOrder))
	}
	if executionOrder[0] != "task1" {
		t.Errorf("expected task1 to run first, got %q", executionOrder[0])
	}
	if executionOrder[1] != "task2" {
		t.Errorf("expected task2 to run second, got %q", executionOrder[1])
	}
}

type orderTrackingTask struct {
	mockTask
	order *[]string
}

func (t *orderTrackingTask) Run(cfg *config.Config, exec executor.Executor) error {
	*t.order = append(*t.order, t.Name())
	return nil
}

func TestCircularDependencyDetection(t *testing.T) {
	// task1 depends on task2, task2 depends on task1
	task1 := &mockTask{
		BaseTask:  task.BaseTask{TaskName: "task1", TaskDeps: []string{"task2"}},
		shouldRun: true,
	}
	task2 := &mockTask{
		BaseTask:  task.BaseTask{TaskName: "task2", TaskDeps: []string{"task1"}},
		shouldRun: true,
	}

	s := NewBaseStrategy("test", []task.Task{task1, task2})
	mock := &mockExecutor{}
	cfg := &config.Config{}

	err := s.Execute(cfg, mock)
	if err == nil {
		t.Error("expected circular dependency error, got nil")
	}
}

func TestRegistry(t *testing.T) {
	// Save registry for restoration
	oldRegistry := make(map[string]func() Strategy)
	for k, v := range Registry {
		oldRegistry[k] = v
	}
	defer func() {
		Registry = oldRegistry
	}()

	// Clear registry for test
	for k := range Registry {
		delete(Registry, k)
	}

	Register("test-strategy", func() Strategy {
		return NewBaseStrategy("test-strategy", nil)
	})

	s, ok := Get("test-strategy")
	if !ok {
		t.Fatal("expected to find registered strategy")
	}
	if s.Name() != "test-strategy" {
		t.Errorf("expected name 'test-strategy', got %q", s.Name())
	}

	_, ok = Get("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent strategy")
	}
}

func TestBuildStrategy(t *testing.T) {
	s := NewBuildStrategy()

	if s.Name() != "build" {
		t.Errorf("expected name 'build', got %q", s.Name())
	}

	tasks := s.Tasks()
	if len(tasks) == 0 {
		t.Error("expected build strategy to have tasks")
	}

	// Verify expected tasks are present
	taskNames := make(map[string]bool)
	for _, task := range tasks {
		taskNames[task.Name()] = true
	}

	expectedTasks := []string{"docker:build", "npm:build", "django:collectstatic"}
	for _, name := range expectedTasks {
		if !taskNames[name] {
			t.Errorf("expected build strategy to contain task %q", name)
		}
	}
}

func TestRunStrategy(t *testing.T) {
	s := NewRunStrategy()

	if s.Name() != "run" {
		t.Errorf("expected name 'run', got %q", s.Name())
	}

	tasks := s.Tasks()
	if len(tasks) == 0 {
		t.Error("expected run strategy to have tasks")
	}
}

func TestNpmScriptStrategy(t *testing.T) {
	s := NewNpmScriptStrategy("lint")

	if s.Name() != "npm:lint" {
		t.Errorf("expected name 'npm:lint', got %q", s.Name())
	}

	tasks := s.Tasks()
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
}

// Verify interface compliance
var _ executor.Executor = &mockExecutor{}
var _ task.Task = &mockTask{}

func TestResolveCommandTasks(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
		Npm: config.NpmConfig{
			Scripts: []string{"build"},
		},
	}

	tests := []struct {
		command string
		want    []string
	}{
		{"build", []string{"docker:build", "npm:build"}},
		{"run", []string{"docker:up"}},
		{"docker down", []string{"docker:down"}},
		{"docker rebuild", []string{"docker:rebuild"}},
		{"npm run build", []string{"npm:run:build"}},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			tasks, err := ResolveCommandTasks(tt.command, cfg)
			if err != nil {
				t.Fatalf("ResolveCommandTasks(%q) error: %v", tt.command, err)
			}

			var got []string
			for _, task := range tasks {
				got = append(got, task.Name())
			}

			if len(got) != len(tt.want) {
				t.Errorf("got %v tasks, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("at index %d: got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
