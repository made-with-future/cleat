package task

import (
	"errors"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
)

// mockExecutor records commands for verification
type mockExecutor struct {
	commands []struct {
		name string
		args []string
	}
	err error
}

func (m *mockExecutor) Run(name string, args ...string) error {
	m.commands = append(m.commands, struct {
		name string
		args []string
	}{name: name, args: args})
	return m.err
}

func TestBaseTask(t *testing.T) {
	bt := &BaseTask{
		TaskName:        "test:task",
		TaskDescription: "A test task",
		TaskDeps:        []string{"dep1", "dep2"},
	}

	if bt.Name() != "test:task" {
		t.Errorf("expected name 'test:task', got %q", bt.Name())
	}
	if bt.Description() != "A test task" {
		t.Errorf("expected description 'A test task', got %q", bt.Description())
	}
	if len(bt.Dependencies()) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(bt.Dependencies()))
	}
}

func TestDockerBuild(t *testing.T) {
	task := NewDockerBuild()

	if task.Name() != "docker:build" {
		t.Errorf("expected name 'docker:build', got %q", task.Name())
	}

	t.Run("ShouldRun with Docker enabled", func(t *testing.T) {
		cfg := &config.Config{Docker: true}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true when Docker is enabled")
		}
	})

	t.Run("ShouldRun with Docker disabled", func(t *testing.T) {
		cfg := &config.Config{Docker: false}
		if task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return false when Docker is disabled")
		}
	})

	t.Run("Run executes docker compose build", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{Docker: true}

		err := task.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "docker" {
			t.Errorf("expected command 'docker', got %q", mock.commands[0].name)
		}
		expectedArgs := []string{"compose", "build"}
		for i, arg := range expectedArgs {
			if mock.commands[0].args[i] != arg {
				t.Errorf("expected arg %d to be %q, got %q", i, arg, mock.commands[0].args[i])
			}
		}
	})

	t.Run("Run returns executor error", func(t *testing.T) {
		mock := &mockExecutor{err: errors.New("docker failed")}
		cfg := &config.Config{Docker: true}

		err := task.Run(cfg, mock)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Commands returns correct command", func(t *testing.T) {
		cfg := &config.Config{Docker: true}
		cmds := task.Commands(cfg)
		if len(cmds) != 1 {
			t.Fatalf("expected 1 command, got %d", len(cmds))
		}
		if cmds[0][0] != "docker" || cmds[0][1] != "compose" || cmds[0][2] != "build" {
			t.Errorf("unexpected command: %v", cmds[0])
		}
	})
}

func TestDockerUp(t *testing.T) {
	task := NewDockerUp()

	if task.Name() != "docker:up" {
		t.Errorf("expected name 'docker:up', got %q", task.Name())
	}

	t.Run("ShouldRun with Docker enabled", func(t *testing.T) {
		cfg := &config.Config{Docker: true}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true when Docker is enabled")
		}
	})

	t.Run("Run executes docker compose up", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{Docker: true}

		err := task.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "docker" {
			t.Errorf("expected command 'docker', got %q", mock.commands[0].name)
		}
	})
}

func TestDockerDown(t *testing.T) {
	task := NewDockerDown()

	if task.Name() != "docker:down" {
		t.Errorf("expected name 'docker:down', got %q", task.Name())
	}

	t.Run("ShouldRun with Docker enabled", func(t *testing.T) {
		cfg := &config.Config{Docker: true}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true when Docker is enabled")
		}
	})

	t.Run("Run executes docker compose down with all profiles", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{Docker: true}

		err := task.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "docker" {
			t.Errorf("expected command 'docker', got %q", mock.commands[0].name)
		}

		expectedArgs := []string{"compose", "--profile", "*", "down", "--remove-orphans"}
		if len(mock.commands[0].args) != len(expectedArgs) {
			t.Fatalf("expected %d args, got %d", len(expectedArgs), len(mock.commands[0].args))
		}
		for i, arg := range expectedArgs {
			if mock.commands[0].args[i] != arg {
				t.Errorf("expected arg %d to be %q, got %q", i, arg, mock.commands[0].args[i])
			}
		}
	})
}

func TestDockerRebuild(t *testing.T) {
	task := NewDockerRebuild()

	if task.Name() != "docker:rebuild" {
		t.Errorf("expected name 'docker:rebuild', got %q", task.Name())
	}

	t.Run("ShouldRun with Docker enabled", func(t *testing.T) {
		cfg := &config.Config{Docker: true}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true when Docker is enabled")
		}
	})

	t.Run("Run executes docker compose down and build", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{Docker: true}

		err := task.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 2 {
			t.Fatalf("expected 2 commands, got %d", len(mock.commands))
		}

		// First command should be docker compose down
		if mock.commands[0].name != "docker" {
			t.Errorf("expected first command 'docker', got %q", mock.commands[0].name)
		}
		expectedDownArgs := []string{"compose", "--profile", "*", "down", "--remove-orphans", "--rmi", "all", "--volumes"}
		for i, arg := range expectedDownArgs {
			if mock.commands[0].args[i] != arg {
				t.Errorf("expected down arg %d to be %q, got %q", i, arg, mock.commands[0].args[i])
			}
		}

		// Second command should be docker compose build
		if mock.commands[1].name != "docker" {
			t.Errorf("expected second command 'docker', got %q", mock.commands[1].name)
		}
		expectedBuildArgs := []string{"compose", "--profile", "*", "build", "--no-cache"}
		for i, arg := range expectedBuildArgs {
			if mock.commands[1].args[i] != arg {
				t.Errorf("expected build arg %d to be %q, got %q", i, arg, mock.commands[1].args[i])
			}
		}
	})

	t.Run("Commands returns two commands", func(t *testing.T) {
		cfg := &config.Config{Docker: true}
		cmds := task.Commands(cfg)
		if len(cmds) != 2 {
			t.Fatalf("expected 2 commands, got %d", len(cmds))
		}
		if cmds[0][0] != "docker" || cmds[0][4] != "down" {
			t.Errorf("unexpected first command: %v", cmds[0])
		}
		if cmds[1][0] != "docker" || cmds[1][4] != "build" {
			t.Errorf("unexpected second command: %v", cmds[1])
		}
	})
}

func TestNpmBuild(t *testing.T) {
	task := NewNpmBuild()

	if task.Name() != "npm:build" {
		t.Errorf("expected name 'npm:build', got %q", task.Name())
	}

	t.Run("ShouldRun with scripts", func(t *testing.T) {
		cfg := &config.Config{Npm: config.NpmConfig{Scripts: []string{"build"}}}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true when scripts exist")
		}
	})

	t.Run("ShouldRun without scripts", func(t *testing.T) {
		cfg := &config.Config{}
		if task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return false when no scripts")
		}
	})

	t.Run("Dependencies includes docker:build", func(t *testing.T) {
		deps := task.Dependencies()
		found := false
		for _, d := range deps {
			if d == "docker:build" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected npm:build to depend on docker:build")
		}
	})

	t.Run("Run via Docker", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{
			Docker: true,
			Npm: config.NpmConfig{
				Service: "node",
				Scripts: []string{"build", "test"},
			},
		}

		err := task.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 2 {
			t.Fatalf("expected 2 commands, got %d", len(mock.commands))
		}
		// Both should use docker compose run
		for _, cmd := range mock.commands {
			if cmd.name != "docker" {
				t.Errorf("expected command 'docker', got %q", cmd.name)
			}
		}
	})

	t.Run("Run locally", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{
			Docker: false,
			Npm:    config.NpmConfig{Scripts: []string{"build"}},
		}

		err := task.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "npm" {
			t.Errorf("expected command 'npm', got %q", mock.commands[0].name)
		}
	})
}

func TestNpmRun(t *testing.T) {
	task := NewNpmRun("lint")

	if task.Name() != "npm:run:lint" {
		t.Errorf("expected name 'npm:run:lint', got %q", task.Name())
	}

	t.Run("ShouldRun always true", func(t *testing.T) {
		cfg := &config.Config{}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true")
		}
	})

	t.Run("Run executes script via Docker", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{
			Docker: true,
			Npm:    config.NpmConfig{Service: "node"},
		}

		err := task.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "docker" {
			t.Errorf("expected command 'docker', got %q", mock.commands[0].name)
		}
	})
}

func TestNpmStart(t *testing.T) {
	task := NewNpmStart()

	t.Run("ShouldRun with npm scripts and no docker/django", func(t *testing.T) {
		cfg := &config.Config{
			Docker: false,
			Django: false,
			Npm:    config.NpmConfig{Scripts: []string{"build"}},
		}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true")
		}
	})

	t.Run("ShouldRun false when Docker enabled", func(t *testing.T) {
		cfg := &config.Config{
			Docker: true,
			Npm:    config.NpmConfig{Scripts: []string{"build"}},
		}
		if task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return false when Docker is enabled")
		}
	})
}

func TestDjangoCollectStatic(t *testing.T) {
	task := NewDjangoCollectStatic()

	if task.Name() != "django:collectstatic" {
		t.Errorf("expected name 'django:collectstatic', got %q", task.Name())
	}

	t.Run("ShouldRun with Django enabled", func(t *testing.T) {
		cfg := &config.Config{Django: true}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true when Django is enabled")
		}
	})

	t.Run("ShouldRun with Django disabled", func(t *testing.T) {
		cfg := &config.Config{Django: false}
		if task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return false when Django is disabled")
		}
	})

	t.Run("Dependencies includes docker:build and npm:build", func(t *testing.T) {
		deps := task.Dependencies()
		hasDocker := false
		hasNpm := false
		for _, d := range deps {
			if d == "docker:build" {
				hasDocker = true
			}
			if d == "npm:build" {
				hasNpm = true
			}
		}
		if !hasDocker || !hasNpm {
			t.Error("expected django:collectstatic to depend on docker:build and npm:build")
		}
	})

	t.Run("Run via Docker", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{
			Docker:        true,
			Django:        true,
			DjangoService: "backend",
		}

		err := task.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "docker" {
			t.Errorf("expected command 'docker', got %q", mock.commands[0].name)
		}
	})

	t.Run("Run locally", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{
			Docker: false,
			Django: true,
		}

		err := task.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "python" {
			t.Errorf("expected command 'python', got %q", mock.commands[0].name)
		}
	})
}

func TestDjangoRunServer(t *testing.T) {
	task := NewDjangoRunServer()

	t.Run("ShouldRun with Django and no Docker", func(t *testing.T) {
		cfg := &config.Config{Django: true, Docker: false}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true")
		}
	})

	t.Run("ShouldRun false when Docker enabled", func(t *testing.T) {
		cfg := &config.Config{Django: true, Docker: true}
		if task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return false when Docker is enabled")
		}
	})
}

// Verify all tasks implement the Task interface
func TestTaskInterface(t *testing.T) {
	var _ Task = NewDockerBuild()
	var _ Task = NewDockerUp()
	var _ Task = NewDockerDown()
	var _ Task = NewDockerRebuild()
	var _ Task = NewNpmBuild()
	var _ Task = NewNpmRun("test")
	var _ Task = NewNpmStart()
	var _ Task = NewDjangoCollectStatic()
	var _ Task = NewDjangoRunServer()
}

// Helper to verify interface at compile time
var _ executor.Executor = &mockExecutor{}
