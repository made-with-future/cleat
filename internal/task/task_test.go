package task

import (
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
)

// mockExecutor records commands for verification
type mockExecutor struct {
	commands []struct {
		dir  string
		name string
		args []string
	}
	prompts []struct {
		message      string
		defaultValue string
	}
	promptResponses map[string]string
	err             error
}

func (m *mockExecutor) Run(name string, args ...string) error {
	return m.RunWithDir("", name, args...)
}

func (m *mockExecutor) RunWithDir(dir string, name string, args ...string) error {
	m.commands = append(m.commands, struct {
		dir  string
		name string
		args []string
	}{dir: dir, name: name, args: args})
	return m.err
}

func (m *mockExecutor) Prompt(message string, defaultValue string) (string, error) {
	m.prompts = append(m.prompts, struct {
		message      string
		defaultValue string
	}{message, defaultValue})
	if m.promptResponses != nil {
		if resp, ok := m.promptResponses[message]; ok {
			return resp, nil
		}
	}
	return defaultValue, nil
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
	task := NewDockerBuild(nil)

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

	t.Run("ShouldRun with Service Docker", func(t *testing.T) {
		svc := &config.ServiceConfig{Name: "svc", Docker: true}
		svcTask := NewDockerBuild(svc)
		cfg := &config.Config{Docker: false}
		if !svcTask.ShouldRun(cfg) {
			t.Error("expected ShouldRun true for service task when service docker is enabled")
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
		expectedArgs := []string{"compose", "--profile", "*", "build"}
		for i, arg := range expectedArgs {
			if mock.commands[0].args[i] != arg {
				t.Errorf("expected arg %d to be %q, got %q", i, arg, mock.commands[0].args[i])
			}
		}
	})

	t.Run("Run with project directory", func(t *testing.T) {
		mock := &mockExecutor{}
		svc := &config.ServiceConfig{Name: "svc", Dir: "./svc"}
		svcTask := NewDockerBuild(svc)
		cfg := &config.Config{}

		err := svcTask.Run(cfg, mock)
		if err != nil {
			t.Fatal(err)
		}

		if mock.commands[0].dir != "./svc" {
			t.Errorf("expected dir './svc', got %q", mock.commands[0].dir)
		}

		expected := []string{"compose", "build"}
		args := mock.commands[0].args
		if len(args) != len(expected) {
			t.Fatalf("expected %d args, got %d", len(expected), len(args))
		}
		for i, v := range expected {
			if args[i] != v {
				t.Errorf("arg %d: expected %q, got %q", i, v, args[i])
			}
		}
	})
}

func TestDockerUp(t *testing.T) {
	task := NewDockerUp(nil)

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
	task := NewDockerDown(nil)

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

func TestDockerRemoveOrphans(t *testing.T) {
	task := NewDockerRemoveOrphans(nil)

	if task.Name() != "docker:remove-orphans" {
		t.Errorf("expected name 'docker:remove-orphans', got %q", task.Name())
	}

	t.Run("ShouldRun with Docker enabled", func(t *testing.T) {
		cfg := &config.Config{Docker: true}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true when Docker is enabled")
		}
	})

	t.Run("Run executes docker compose down --remove-orphans with all profiles", func(t *testing.T) {
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
	task := NewDockerRebuild(nil)

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
	svc := &config.ServiceConfig{Name: "default", Dir: "."}
	npm := &config.NpmConfig{Scripts: []string{"build"}}
	task := NewNpmBuild(svc, npm)

	if task.Name() != "npm:build" {
		t.Errorf("expected name 'npm:build', got %q", task.Name())
	}

	t.Run("ShouldRun with scripts", func(t *testing.T) {
		cfg := &config.Config{}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true when scripts exist")
		}
	})

	t.Run("ShouldRun without scripts", func(t *testing.T) {
		cfg := &config.Config{}
		taskNoScripts := NewNpmBuild(svc, &config.NpmConfig{})
		if taskNoScripts.ShouldRun(cfg) {
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
		}
		npmMod := &config.NpmConfig{
			Service: "node",
			Scripts: []string{"build", "test"},
		}
		taskDocker := NewNpmBuild(svc, npmMod)

		err := taskDocker.Run(cfg, mock)
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
		}
		npmMod := &config.NpmConfig{Scripts: []string{"build"}}
		taskLocal := NewNpmBuild(svc, npmMod)

		err := taskLocal.Run(cfg, mock)
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
	svc := &config.ServiceConfig{Name: "default", Dir: "."}
	npm := &config.NpmConfig{Service: "node"}
	task := NewNpmRun(svc, npm, "lint")

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
	svc := &config.ServiceConfig{Name: "default", Dir: "."}
	npm := &config.NpmConfig{Scripts: []string{"build"}}
	task := NewNpmStart(svc, npm)

	t.Run("ShouldRun with npm scripts and no docker", func(t *testing.T) {
		cfg := &config.Config{
			Docker: false,
		}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true")
		}
	})

	t.Run("ShouldRun false when Docker enabled", func(t *testing.T) {
		cfg := &config.Config{
			Docker: true,
		}
		if task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return false when Docker is enabled")
		}
	})
}

func TestDjangoCollectStatic(t *testing.T) {
	svc := &config.ServiceConfig{Name: "default", Dir: "."}
	python := &config.PythonConfig{Django: true}
	task := NewDjangoCollectStatic(svc, python)

	if task.Name() != "django:collectstatic" {
		t.Errorf("expected name 'django:collectstatic', got %q", task.Name())
	}

	t.Run("ShouldRun with Django enabled", func(t *testing.T) {
		cfg := &config.Config{}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true when Django is enabled")
		}
	})

	t.Run("ShouldRun with Django disabled", func(t *testing.T) {
		cfg := &config.Config{}
		taskDisabled := NewDjangoCollectStatic(svc, &config.PythonConfig{Django: false})
		if taskDisabled.ShouldRun(cfg) {
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
			Docker: true,
		}
		pythonMod := &config.PythonConfig{
			Django:        true,
			DjangoService: "backend",
		}
		taskDocker := NewDjangoCollectStatic(svc, pythonMod)

		err := taskDocker.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "docker" {
			t.Errorf("expected command 'docker', got %q", mock.commands[0].name)
		}

		expectedArgs := []string{"compose", "run", "--rm", "backend", "uv", "run", "python", "manage.py", "collectstatic", "--noinput", "--clear"}
		if len(mock.commands[0].args) != len(expectedArgs) {
			t.Fatalf("expected %d args, got %d", len(expectedArgs), len(mock.commands[0].args))
		}
		for i, arg := range expectedArgs {
			if mock.commands[0].args[i] != arg {
				t.Errorf("expected arg %d to be %q, got %q", i, arg, mock.commands[0].args[i])
			}
		}
	})

	t.Run("Run locally", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{
			Docker: false,
		}
		pythonMod := &config.PythonConfig{Django: true}
		taskLocal := NewDjangoCollectStatic(svc, pythonMod)

		err := taskLocal.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "uv" {
			t.Errorf("expected command 'uv', got %q", mock.commands[0].name)
		}
	})

	t.Run("Run locally with pip", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{
			Docker: false,
		}
		pythonMod := &config.PythonConfig{Django: true, PackageManager: "pip"}
		taskPip := NewDjangoCollectStatic(svc, pythonMod)

		err := taskPip.Run(cfg, mock)
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
	svc := &config.ServiceConfig{Name: "default", Dir: "."}
	python := &config.PythonConfig{Django: true}
	task := NewDjangoRunServer(svc, python)

	t.Run("ShouldRun with Django and no Docker", func(t *testing.T) {
		cfg := &config.Config{Docker: false}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true")
		}
	})

	t.Run("ShouldRun false when Docker enabled", func(t *testing.T) {
		cfg := &config.Config{Docker: true}
		if task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return false when Docker is enabled")
		}
	})
}

func TestDjangoCreateUserDev(t *testing.T) {
	svc := &config.ServiceConfig{Name: "default", Dir: "."}
	python := &config.PythonConfig{Django: true, DjangoService: "backend"}
	task := NewDjangoCreateUserDev(svc, python)

	if task.Name() != "django:create-user-dev" {
		t.Errorf("expected name 'django:create-user-dev', got %q", task.Name())
	}

	t.Run("ShouldRun with Django and Docker enabled", func(t *testing.T) {
		cfg := &config.Config{Docker: true}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true")
		}
	})

	t.Run("ShouldRun false when Docker disabled", func(t *testing.T) {
		cfg := &config.Config{Docker: false}
		if task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return false")
		}
	})
}

func TestDjangoMigrate(t *testing.T) {
	svc := &config.ServiceConfig{Name: "default", Dir: "."}
	python := &config.PythonConfig{Django: true, DjangoService: "backend"}
	task := NewDjangoMigrate(svc, python)

	if task.Name() != "django:migrate" {
		t.Errorf("expected name 'django:migrate', got %q", task.Name())
	}

	t.Run("ShouldRun with Django enabled", func(t *testing.T) {
		cfg := &config.Config{}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true when Django is enabled")
		}
	})

	t.Run("Run via Docker", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{
			Docker: true,
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

		expectedArgs := []string{"compose", "run", "--rm", "backend", "uv", "run", "python", "manage.py", "migrate", "--noinput"}
		if len(mock.commands[0].args) != len(expectedArgs) {
			t.Fatalf("expected %d args, got %d", len(expectedArgs), len(mock.commands[0].args))
		}
		for i, v := range expectedArgs {
			if mock.commands[0].args[i] != v {
				t.Errorf("arg %d: expected %q, got %q", i, v, mock.commands[0].args[i])
			}
		}
	})

	t.Run("Run via Docker with pip", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{
			Docker: true,
		}
		pythonPip := &config.PythonConfig{Django: true, DjangoService: "backend", PackageManager: "pip"}
		taskPip := NewDjangoMigrate(svc, pythonPip)

		err := taskPip.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expectedArgs := []string{"compose", "run", "--rm", "backend", "python", "manage.py", "migrate", "--noinput"}
		actualArgs := mock.commands[0].args
		if len(actualArgs) != len(expectedArgs) {
			t.Fatalf("expected %d args, got %d", len(expectedArgs), len(actualArgs))
		}
		for i, v := range expectedArgs {
			if actualArgs[i] != v {
				t.Errorf("arg %d: expected %q, got %q", i, v, actualArgs[i])
			}
		}
	})
}

func TestDjangoMakeMigrations(t *testing.T) {
	svc := &config.ServiceConfig{Name: "default", Dir: "."}
	python := &config.PythonConfig{Django: true, DjangoService: "backend"}
	task := NewDjangoMakeMigrations(svc, python)

	if task.Name() != "django:makemigrations" {
		t.Errorf("expected name 'django:makemigrations', got %q", task.Name())
	}

	t.Run("ShouldRun with Django enabled", func(t *testing.T) {
		cfg := &config.Config{}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true when Django is enabled")
		}
	})

	t.Run("Run via Docker", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{
			Docker: true,
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

		expectedArgs := []string{"compose", "run", "--rm", "backend", "uv", "run", "python", "manage.py", "makemigrations"}
		if len(mock.commands[0].args) != len(expectedArgs) {
			t.Fatalf("expected %d args, got %d", len(expectedArgs), len(mock.commands[0].args))
		}
		for i, v := range expectedArgs {
			if mock.commands[0].args[i] != v {
				t.Errorf("arg %d: expected %q, got %q", i, v, mock.commands[0].args[i])
			}
		}
	})

	t.Run("Run locally", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{
			Docker: false,
		}

		err := task.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "uv" {
			t.Errorf("expected command 'uv', got %q", mock.commands[0].name)
		}
	})
}

func TestDjangoGenRandomSecretKey(t *testing.T) {
	svc := &config.ServiceConfig{Name: "default", Dir: "."}
	python := &config.PythonConfig{Django: true, DjangoService: "backend"}
	task := NewDjangoGenRandomSecretKey(svc, python)

	if task.Name() != "django:gen-random-secret-key" {
		t.Errorf("expected name 'django:gen-random-secret-key', got %q", task.Name())
	}

	t.Run("ShouldRun with Django enabled", func(t *testing.T) {
		cfg := &config.Config{}
		if !task.ShouldRun(cfg) {
			t.Error("expected ShouldRun to return true when Django is enabled")
		}
	})

	t.Run("Run via Docker", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{
			Docker: true,
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

		pyCmd := "from django.core.management.utils import get_random_secret_key; print(get_random_secret_key())"
		expectedArgs := []string{"compose", "run", "--rm", "backend", "uv", "run", "python", "-c", pyCmd}
		if len(mock.commands[0].args) != len(expectedArgs) {
			t.Fatalf("expected %d args, got %d", len(expectedArgs), len(mock.commands[0].args))
		}
		for i, v := range expectedArgs {
			if mock.commands[0].args[i] != v {
				t.Errorf("arg %d: expected %q, got %q", i, v, mock.commands[0].args[i])
			}
		}
	})

	t.Run("Run locally with uv", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{
			Docker: false,
		}

		err := task.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "uv" {
			t.Errorf("expected command 'uv', got %q", mock.commands[0].name)
		}

		pyCmd := "from django.core.management.utils import get_random_secret_key; print(get_random_secret_key())"
		expectedArgs := []string{"run", "python", "-c", pyCmd}
		if len(mock.commands[0].args) != len(expectedArgs) {
			t.Fatalf("expected %d args, got %d", len(expectedArgs), len(mock.commands[0].args))
		}
		for i, v := range expectedArgs {
			if mock.commands[0].args[i] != v {
				t.Errorf("arg %d: expected %q, got %q", i, v, mock.commands[0].args[i])
			}
		}
	})

	t.Run("Run locally with pip", func(t *testing.T) {
		mock := &mockExecutor{}
		cfg := &config.Config{
			Docker: false,
		}
		pythonPip := &config.PythonConfig{Django: true, DjangoService: "backend", PackageManager: "pip"}
		taskPip := NewDjangoGenRandomSecretKey(svc, pythonPip)

		err := taskPip.Run(cfg, mock)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "python" {
			t.Errorf("expected command 'python', got %q", mock.commands[0].name)
		}

		pyCmd := "from django.core.management.utils import get_random_secret_key; print(get_random_secret_key())"
		expectedArgs := []string{"-c", pyCmd}
		if len(mock.commands[0].args) != len(expectedArgs) {
			t.Fatalf("expected %d args, got %d", len(expectedArgs), len(mock.commands[0].args))
		}
		for i, v := range expectedArgs {
			if mock.commands[0].args[i] != v {
				t.Errorf("arg %d: expected %q, got %q", i, v, mock.commands[0].args[i])
			}
		}
	})
}

func TestGCPActivate(t *testing.T) {
	cfg := &config.Config{
		GoogleCloudPlatform: &config.GCPConfig{
			ProjectName: "test-project",
		},
	}
	task := NewGCPActivate()

	if !task.ShouldRun(cfg) {
		t.Error("Expected ShouldRun to return true")
	}

	mock := &mockExecutor{}
	if err := task.Run(cfg, mock); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(mock.commands) != 2 {
		t.Errorf("Expected 2 calls, got %d", len(mock.commands))
	}
	expected1 := "gcloud config configurations activate test-project"
	actual1 := mock.commands[0].name
	for _, arg := range mock.commands[0].args {
		actual1 += " " + arg
	}
	if actual1 != expected1 {
		t.Errorf("Expected call 1 '%s', got '%s'", expected1, actual1)
	}

	expected2 := "gcloud auth application-default set-quota-project test-project"
	actual2 := mock.commands[1].name
	for _, arg := range mock.commands[1].args {
		actual2 += " " + arg
	}
	if actual2 != expected2 {
		t.Errorf("Expected call 2 '%s', got '%s'", expected2, actual2)
	}
}

func TestGCPInit(t *testing.T) {
	cfg := &config.Config{
		GoogleCloudPlatform: &config.GCPConfig{
			ProjectName: "test-project",
		},
	}
	task := NewGCPInit()

	if !task.ShouldRun(cfg) {
		t.Error("Expected ShouldRun to return true")
	}

	mock := &mockExecutor{}
	if err := task.Run(cfg, mock); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(mock.commands) != 1 {
		t.Errorf("Expected 1 call, got %d", len(mock.commands))
	}
	expected1 := "gcloud config configurations create test-project"
	actual1 := mock.commands[0].name
	for _, arg := range mock.commands[0].args {
		actual1 += " " + arg
	}
	if actual1 != expected1 {
		t.Errorf("Expected call 1 '%s', got '%s'", expected1, actual1)
	}
}

func TestTerraformTask(t *testing.T) {
	cfg := &config.Config{
		Terraform: &config.TerraformConfig{UseFolders: true},
		Envs:      []string{"production"},
	}

	t.Run("init", func(t *testing.T) {
		task := NewTerraformTask("production", "init", nil)
		if !task.ShouldRun(cfg) {
			t.Error("Expected ShouldRun to be true")
		}

		mock := &mockExecutor{}
		if err := task.Run(cfg, mock); err != nil {
			t.Fatalf("Run failed: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("Expected 1 call, got %d", len(mock.commands))
		}

		// Expectations depend on whether 'op' is in PATH and .envs/production.env exists
		// In tests, they likely don't.
		expected := "terraform -chdir=.iac/production init"
		actual := mock.commands[0].name
		for _, arg := range mock.commands[0].args {
			actual += " " + arg
		}
		if actual != expected {
			t.Errorf("Expected '%s', got '%s'", expected, actual)
		}
	})

	t.Run("single-env", func(t *testing.T) {
		singleCfg := &config.Config{
			Terraform: &config.TerraformConfig{UseFolders: false},
		}
		task := NewTerraformTask("", "plan", nil)
		mock := &mockExecutor{}
		task.Run(singleCfg, mock)

		expected := "terraform -chdir=.iac plan"
		actual := mock.commands[0].name
		for _, arg := range mock.commands[0].args {
			actual += " " + arg
		}
		if actual != expected {
			t.Errorf("Expected '%s', got '%s'", expected, actual)
		}
	})

	t.Run("init-upgrade", func(t *testing.T) {
		task := NewTerraformTask("production", "init", []string{"-upgrade"})
		mock := &mockExecutor{}
		task.Run(cfg, mock)

		expected := "terraform -chdir=.iac/production init -upgrade"
		actual := mock.commands[0].name
		for _, arg := range mock.commands[0].args {
			actual += " " + arg
		}
		if actual != expected {
			t.Errorf("Expected '%s', got '%s'", expected, actual)
		}
	})

	t.Run("apply-refresh", func(t *testing.T) {
		task := NewTerraformTask("production", "apply", []string{"-refresh-only"})
		mock := &mockExecutor{}
		task.Run(cfg, mock)

		expected := "terraform -chdir=.iac/production apply -refresh-only"
		actual := mock.commands[0].name
		for _, arg := range mock.commands[0].args {
			actual += " " + arg
		}
		if actual != expected {
			t.Errorf("Expected '%s', got '%s'", expected, actual)
		}
	})
}

func TestGCPADCLogin(t *testing.T) {
	cfg := &config.Config{
		GoogleCloudPlatform: &config.GCPConfig{
			ProjectName: "test-project",
		},
	}
	task := NewGCPADCLogin()

	if !task.ShouldRun(cfg) {
		t.Error("Expected ShouldRun to return true")
	}

	mock := &mockExecutor{}
	if err := task.Run(cfg, mock); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(mock.commands) != 4 {
		t.Errorf("Expected 4 calls, got %d", len(mock.commands))
	}
	expectedCommands := []string{
		"gcloud config configurations activate test-project",
		"gcloud auth application-default login --project test-project",
		"gcloud auth login --project test-project",
		"gcloud auth application-default set-quota-project test-project",
	}

	for i, expected := range expectedCommands {
		actual := mock.commands[i].name
		for _, arg := range mock.commands[i].args {
			actual += " " + arg
		}
		if actual != expected {
			t.Errorf("Expected call %d '%s', got '%s'", i+1, expected, actual)
		}
	}
}

func TestGCPSetConfig(t *testing.T) {
	t.Run("Without account", func(t *testing.T) {
		cfg := &config.Config{
			GoogleCloudPlatform: &config.GCPConfig{
				ProjectName: "test-project",
			},
		}
		task := NewGCPSetConfig()

		mock := &mockExecutor{}
		if err := task.Run(cfg, mock); err != nil {
			t.Fatalf("Run failed: %v", err)
		}

		if len(mock.commands) != 3 {
			t.Errorf("Expected 3 calls, got %d", len(mock.commands))
		}
		expectedCommands := []string{
			"gcloud config set project test-project",
			"gcloud config set app/promote_by_default false",
			"gcloud config set billing/quota_project test-project",
		}

		for i, expected := range expectedCommands {
			actual := mock.commands[i].name
			for _, arg := range mock.commands[i].args {
				actual += " " + arg
			}
			if actual != expected {
				t.Errorf("Expected call %d '%s', got '%s'", i+1, expected, actual)
			}
		}
	})

	t.Run("With account", func(t *testing.T) {
		cfg := &config.Config{
			GoogleCloudPlatform: &config.GCPConfig{
				ProjectName: "test-project",
				Account:     "test@example.com",
			},
		}
		task := NewGCPSetConfig()

		mock := &mockExecutor{}
		if err := task.Run(cfg, mock); err != nil {
			t.Fatalf("Run failed: %v", err)
		}

		if len(mock.commands) != 4 {
			t.Errorf("Expected 4 calls, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "gcloud" || mock.commands[0].args[3] != "test@example.com" {
			t.Errorf("Expected account set to test@example.com, got %v", mock.commands[0])
		}
	})

	t.Run("With account from inputs", func(t *testing.T) {
		cfg := &config.Config{
			GoogleCloudPlatform: &config.GCPConfig{
				ProjectName: "test-project",
			},
			Inputs: map[string]string{
				"gcp:account": "input@example.com",
			},
		}
		task := NewGCPSetConfig()

		mock := &mockExecutor{}
		if err := task.Run(cfg, mock); err != nil {
			t.Fatalf("Run failed: %v", err)
		}

		if len(mock.commands) != 4 {
			t.Errorf("Expected 4 calls, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "gcloud" || mock.commands[0].args[3] != "input@example.com" {
			t.Errorf("Expected account set to input@example.com, got %v", mock.commands[0])
		}
	})

	t.Run("Requirements", func(t *testing.T) {
		task := NewGCPSetConfig()

		t.Run("Needs account", func(t *testing.T) {
			cfg := &config.Config{
				GoogleCloudPlatform: &config.GCPConfig{
					ProjectName: "test-project",
				},
			}
			reqs := task.Requirements(cfg)
			if len(reqs) != 1 || reqs[0].Key != "gcp:account" {
				t.Errorf("Expected 1 requirement for gcp:account, got %v", reqs)
			}
		})

		t.Run("Has account in config", func(t *testing.T) {
			cfg := &config.Config{
				GoogleCloudPlatform: &config.GCPConfig{
					ProjectName: "test-project",
					Account:     "test@example.com",
				},
			}
			reqs := task.Requirements(cfg)
			if len(reqs) != 0 {
				t.Errorf("Expected 0 requirements, got %v", reqs)
			}
		})
	})
}

// Verify all tasks implement the Task interface
func TestTaskInterface(t *testing.T) {
	svc := &config.ServiceConfig{Name: "default", Dir: "."}
	python := &config.PythonConfig{Django: true}
	npm := &config.NpmConfig{Scripts: []string{"build"}}

	var _ Task = NewDockerBuild(nil)
	var _ Task = NewDockerUp(nil)
	var _ Task = NewDockerDown(nil)
	var _ Task = NewDockerRebuild(nil)
	var _ Task = NewNpmBuild(svc, npm)
	var _ Task = NewNpmRun(svc, npm, "test")
	var _ Task = NewNpmStart(svc, npm)
	var _ Task = NewDjangoCollectStatic(svc, python)
	var _ Task = NewDjangoRunServer(svc, python)
	var _ Task = NewDjangoCreateUserDev(svc, python)
	var _ Task = NewDjangoMigrate(svc, python)
	var _ Task = NewGCPActivate()
	var _ Task = NewGCPInit()
	var _ Task = NewGCPSetConfig()
}

// Helper to verify interface at compile time
var _ executor.Executor = &mockExecutor{}
