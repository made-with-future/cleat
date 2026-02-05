package task

import (
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/session"
)

func ptrBool(b bool) *bool {
	return &b
}

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

func newTestSession(cfg *config.Config, exec executor.Executor) *session.Session {
	sess := session.NewSession(cfg, exec)
	if cfg != nil && cfg.Inputs != nil {
		for k, v := range cfg.Inputs {
			sess.Inputs[k] = v
		}
	}
	return sess
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
		sess := newTestSession(&config.Config{Docker: true}, nil)
		if !task.ShouldRun(sess) {
			t.Error("expected ShouldRun to return true when Docker is enabled")
		}
	})

	t.Run("ShouldRun with Docker disabled", func(t *testing.T) {
		sess := newTestSession(&config.Config{Docker: false}, nil)
		if task.ShouldRun(sess) {
			t.Error("expected ShouldRun to return false when Docker is disabled")
		}
	})

	t.Run("ShouldRun with Service Docker", func(t *testing.T) {
		svc := &config.ServiceConfig{Name: "svc", Docker: ptrBool(true)}
		svcTask := NewDockerBuild(svc)
		sess := newTestSession(&config.Config{Docker: false}, nil)
		if !svcTask.ShouldRun(sess) {
			t.Error("expected ShouldRun true for service task when service docker is enabled")
		}
	})

	t.Run("Run executes docker compose build", func(t *testing.T) {
		mock := &mockExecutor{}
		sess := newTestSession(&config.Config{Docker: true}, mock)

		err := task.Run(sess)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "docker" {
			t.Errorf("expected command 'docker', got %q", mock.commands[0].name)
		}
		expectedArgs := []string{"--log-level", "error", "compose", "--profile", "*", "build"}
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
		sess := newTestSession(&config.Config{}, mock)

		err := svcTask.Run(sess)
		if err != nil {
			t.Fatal(err)
		}

		if mock.commands[0].dir != "./svc" {
			t.Errorf("expected dir './svc', got %q", mock.commands[0].dir)
		}

		expected := []string{"--log-level", "error", "compose", "build"}
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
		sess := newTestSession(&config.Config{Docker: true}, nil)
		if !task.ShouldRun(sess) {
			t.Error("expected ShouldRun to return true when Docker is enabled")
		}
	})

	t.Run("Run executes docker compose up", func(t *testing.T) {
		mock := &mockExecutor{}
		sess := newTestSession(&config.Config{Docker: true}, mock)

		err := task.Run(sess)
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
		sess := newTestSession(&config.Config{Docker: true}, nil)
		if !task.ShouldRun(sess) {
			t.Error("expected ShouldRun to return true when Docker is enabled")
		}
	})

	t.Run("Run executes docker compose down with all profiles", func(t *testing.T) {
		mock := &mockExecutor{}
		sess := newTestSession(&config.Config{Docker: true}, mock)

		err := task.Run(sess)
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
		sess := newTestSession(&config.Config{Docker: true}, nil)
		if !task.ShouldRun(sess) {
			t.Error("expected ShouldRun to return true when Docker is enabled")
		}
	})

	t.Run("Run executes docker compose down --remove-orphans with all profiles", func(t *testing.T) {
		mock := &mockExecutor{}
		sess := newTestSession(&config.Config{Docker: true}, mock)

		err := task.Run(sess)
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
		sess := newTestSession(&config.Config{Docker: true}, nil)
		if !task.ShouldRun(sess) {
			t.Error("expected ShouldRun to return true when Docker is enabled")
		}
	})

	t.Run("Run executes docker compose down and build", func(t *testing.T) {
		mock := &mockExecutor{}
		sess := newTestSession(&config.Config{Docker: true}, mock)

		err := task.Run(sess)
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
		expectedBuildArgs := []string{"--log-level", "error", "compose", "--profile", "*", "build", "--no-cache"}
		for i, arg := range expectedBuildArgs {
			if mock.commands[1].args[i] != arg {
				t.Errorf("expected build arg %d to be %q, got %q", i, arg, mock.commands[1].args[i])
			}
		}
	})

	t.Run("Commands returns two commands", func(t *testing.T) {
		sess := newTestSession(&config.Config{Docker: true}, nil)
		cmds := task.Commands(sess)
		if len(cmds) != 2 {
			t.Fatalf("expected 2 commands, got %d", len(cmds))
		}
		if cmds[0][0] != "docker" || cmds[0][4] != "down" {
			t.Errorf("unexpected first command: %v", cmds[0])
		}
		if cmds[1][0] != "docker" || cmds[1][6] != "build" {
			t.Errorf("unexpected second command: %v", cmds[1])
		}
	})
}

func TestNpmInstall(t *testing.T) {
	svc := &config.ServiceConfig{Name: "default", Dir: "."}
	npm := &config.NpmConfig{}
	task := NewNpmInstall(svc, npm)

	if task.Name() != "npm:install" {
		t.Errorf("expected name 'npm:install', got %q", task.Name())
	}

	t.Run("ShouldRun with npm config", func(t *testing.T) {
		sess := newTestSession(&config.Config{}, nil)
		if !task.ShouldRun(sess) {
			t.Error("expected ShouldRun to return true when npm config exists")
		}
	})

	t.Run("Run locally", func(t *testing.T) {
		mock := &mockExecutor{}
		sess := newTestSession(&config.Config{
			Docker: false,
		}, mock)
		taskLocal := NewNpmInstall(svc, npm)

		err := taskLocal.Run(sess)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mock.commands) != 1 {
			t.Fatalf("expected 1 command, got %d", len(mock.commands))
		}
		if mock.commands[0].name != "npm" {
			t.Errorf("expected command 'npm', got %q", mock.commands[0].name)
		}
		if mock.commands[0].args[0] != "install" {
			t.Errorf("expected arg 'install', got %q", mock.commands[0].args[0])
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
		sess := newTestSession(&config.Config{}, nil)
		if !task.ShouldRun(sess) {
			t.Error("expected ShouldRun to return true")
		}
	})

	t.Run("Run executes script", func(t *testing.T) {
		mock := &mockExecutor{}
		sess := newTestSession(&config.Config{
			Docker: false,
		}, mock)
		taskDocker := NewNpmRun(svc, npm, "lint")

		err := taskDocker.Run(sess)
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

func TestDjangoCollectStatic(t *testing.T) {
	svc := &config.ServiceConfig{Name: "default", Dir: "."}
	task := NewDjangoCollectStatic(svc)

	if task.Name() != "django:collectstatic" {
		t.Errorf("expected name 'django:collectstatic', got %q", task.Name())
	}

	t.Run("ShouldRun with Django enabled", func(t *testing.T) {
		svcWithDjango := &config.ServiceConfig{
			Name: "default",
			Modules: []config.ModuleConfig{
				{Python: &config.PythonConfig{Django: true}},
			},
		}
		taskWithDjango := NewDjangoCollectStatic(svcWithDjango)
		sess := newTestSession(&config.Config{}, nil)
		if !taskWithDjango.ShouldRun(sess) {
			t.Error("expected ShouldRun to return true when Django is enabled")
		}
	})

	t.Run("ShouldRun with Django disabled", func(t *testing.T) {
		sess := newTestSession(&config.Config{}, nil)
		taskDisabled := NewDjangoCollectStatic(svc)
		if taskDisabled.ShouldRun(sess) {
			t.Error("expected ShouldRun to return false when Django is disabled")
		}
	})

	t.Run("Run locally", func(t *testing.T) {
		mock := &mockExecutor{}
		sess := newTestSession(&config.Config{
			Docker: false,
		}, mock)
		svcWithDjango := &config.ServiceConfig{
			Name: "default",
			Modules: []config.ModuleConfig{
				{Python: &config.PythonConfig{Django: true}},
			},
		}
		taskLocal := NewDjangoCollectStatic(svcWithDjango)

		err := taskLocal.Run(sess)
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

func TestDjangoRunServer(t *testing.T) {
	svc := &config.ServiceConfig{
		Name: "default",
		Dir:  ".",
		Modules: []config.ModuleConfig{
			{Python: &config.PythonConfig{Django: true}},
		},
	}
	task := NewDjangoRunServer(svc)

	t.Run("ShouldRun with Django enabled", func(t *testing.T) {
		sess := newTestSession(&config.Config{Docker: false}, nil)
		if !task.ShouldRun(sess) {
			t.Error("expected ShouldRun to return true when Django is enabled")
		}
	})
}

func TestDjangoCreateUserDev(t *testing.T) {
	svc := &config.ServiceConfig{
		Name: "default",
		Dir:  ".",
		Modules: []config.ModuleConfig{
			{Python: &config.PythonConfig{Django: true}},
		},
	}
	task := NewDjangoCreateUserDev(svc)

	if task.Name() != "django:create-user-dev" {
		t.Errorf("expected name 'django:create-user-dev', got %q", task.Name())
	}

	t.Run("ShouldRun with Django enabled", func(t *testing.T) {
		sess := newTestSession(&config.Config{Docker: true}, nil)
		if !task.ShouldRun(sess) {
			t.Error("expected ShouldRun to return true")
		}
	})
}

func TestDjangoMigrate(t *testing.T) {
	svc := &config.ServiceConfig{
		Name: "default",
		Dir:  ".",
		Modules: []config.ModuleConfig{
			{Python: &config.PythonConfig{Django: true}},
		},
	}
	task := NewDjangoMigrate(svc)

	if task.Name() != "django:migrate" {
		t.Errorf("expected name 'django:migrate', got %q", task.Name())
	}

	t.Run("ShouldRun with Django enabled", func(t *testing.T) {
		sess := newTestSession(&config.Config{}, nil)
		if !task.ShouldRun(sess) {
			t.Error("expected ShouldRun to return true when Django is enabled")
		}
	})

	t.Run("Run locally", func(t *testing.T) {
		mock := &mockExecutor{}
		sess := newTestSession(&config.Config{
			Docker: false,
		}, mock)

		err := task.Run(sess)
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

func TestDjangoMakeMigrations(t *testing.T) {
	svc := &config.ServiceConfig{
		Name: "default",
		Dir:  ".",
		Modules: []config.ModuleConfig{
			{Python: &config.PythonConfig{Django: true}},
		},
	}
	task := NewDjangoMakeMigrations(svc)

	if task.Name() != "django:makemigrations" {
		t.Errorf("expected name 'django:makemigrations', got %q", task.Name())
	}

	t.Run("ShouldRun with Django enabled", func(t *testing.T) {
		sess := newTestSession(&config.Config{}, nil)
		if !task.ShouldRun(sess) {
			t.Error("expected ShouldRun to return true when Django is enabled")
		}
	})

	t.Run("Run locally", func(t *testing.T) {
		mock := &mockExecutor{}
		sess := newTestSession(&config.Config{
			Docker: false,
		}, mock)

		err := task.Run(sess)
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
	svc := &config.ServiceConfig{
		Name: "default",
		Dir:  ".",
		Modules: []config.ModuleConfig{
			{Python: &config.PythonConfig{Django: true}},
		},
	}
	task := NewDjangoGenRandomSecretKey(svc)

	if task.Name() != "django:gen-random-secret-key" {
		t.Errorf("expected name 'django:gen-random-secret-key', got %q", task.Name())
	}

	t.Run("ShouldRun with Django enabled", func(t *testing.T) {
		sess := newTestSession(&config.Config{}, nil)
		if !task.ShouldRun(sess) {
			t.Error("expected ShouldRun to return true when Django is enabled")
		}
	})

	t.Run("Run locally with uv", func(t *testing.T) {
		mock := &mockExecutor{}
		sess := newTestSession(&config.Config{
			Docker: false,
		}, mock)

		err := task.Run(sess)
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

func TestGCPActivate(t *testing.T) {
	sess := newTestSession(&config.Config{
		GoogleCloudPlatform: &config.GCPConfig{
			ProjectName: "test-project",
		},
	}, &mockExecutor{})
	task := NewGCPActivate()

	if !task.ShouldRun(sess) {
		t.Error("Expected ShouldRun to return true")
	}

	if err := task.Run(sess); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	mock := sess.Exec.(*mockExecutor)
	if len(mock.commands) != 1 {
		t.Errorf("Expected 1 call, got %d", len(mock.commands))
	}
	expected1 := "gcloud config set project test-project"
	actual1 := mock.commands[0].name
	for _, arg := range mock.commands[0].args {
		actual1 += " " + arg
	}
	if actual1 != expected1 {
		t.Errorf("Expected call 1 '%s', got '%s'", expected1, actual1)
	}
}

func TestGCPInit(t *testing.T) {
	sess := newTestSession(&config.Config{
		GoogleCloudPlatform: &config.GCPConfig{
			ProjectName: "test-project",
		},
	}, &mockExecutor{})
	task := NewGCPInit()

	if !task.ShouldRun(sess) {
		t.Error("Expected ShouldRun to return true")
	}

	if err := task.Run(sess); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	mock := sess.Exec.(*mockExecutor)
	if len(mock.commands) != 1 {
		t.Errorf("Expected 1 call, got %d", len(mock.commands))
	}
	expected1 := "gcloud config set project test-project"
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
		task := NewTerraform("production", "init", nil)
		sess := newTestSession(cfg, &mockExecutor{})
		if !task.ShouldRun(sess) {
			t.Error("Expected ShouldRun to be true")
		}

		if err := task.Run(sess); err != nil {
			t.Fatalf("Run failed: %v", err)
		}

		mock := sess.Exec.(*mockExecutor)
		if len(mock.commands) != 1 {
			t.Fatalf("Expected 1 call, got %d", len(mock.commands))
		}

		if mock.commands[0].dir != ".iac/production" {
			t.Errorf("Expected dir '.iac/production', got %q", mock.commands[0].dir)
		}
		if mock.commands[0].name != "terraform" || mock.commands[0].args[0] != "init" {
			t.Errorf("Expected terraform init, got %v %v", mock.commands[0].name, mock.commands[0].args)
		}
	})

	t.Run("single-env", func(t *testing.T) {
		singleCfg := &config.Config{
			Terraform: &config.TerraformConfig{UseFolders: false},
		}
		task := NewTerraform("", "plan", nil)
		sess := newTestSession(singleCfg, &mockExecutor{})
		task.Run(sess)

		mock := sess.Exec.(*mockExecutor)
		if mock.commands[0].dir != ".iac" {
			t.Errorf("Expected dir '.iac', got %q", mock.commands[0].dir)
		}
	})
}

// Verify all tasks implement the Task interface
func TestTaskInterface(t *testing.T) {
	svc := &config.ServiceConfig{Name: "default", Dir: "."}

	var _ Task = NewDockerBuild(nil)
	var _ Task = NewDockerUp(nil)
	var _ Task = NewDockerDown(nil)
	var _ Task = NewDockerRebuild(nil)
	var _ Task = NewNpmRun(svc, nil, "test")
	var _ Task = NewDjangoCollectStatic(svc)
	var _ Task = NewDjangoRunServer(svc)
	var _ Task = NewDjangoCreateUserDev(svc)
	var _ Task = NewDjangoMigrate(svc)
	var _ Task = NewGCPActivate()
	var _ Task = NewGCPInit()
	var _ Task = NewGCPAppEnginePromote("")
}

// Helper to verify interface at compile time
var _ executor.Executor = &mockExecutor{}
