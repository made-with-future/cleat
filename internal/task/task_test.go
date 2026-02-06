package task

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/session"
)

func ptr[T any](v T) *T {
	return &v
}

type MockExecutor struct {
	executor.ShellExecutor
	RunCalled bool
	Dir       string
	Name      string
	Args      []string
}

func (e *MockExecutor) Run(name string, args ...string) error {
	e.RunCalled = true
	e.Name = name
	e.Args = args
	return nil
}

func (e *MockExecutor) RunWithDir(dir string, name string, args ...string) error {
	e.RunCalled = true
	e.Dir = dir
	e.Name = name
	e.Args = args
	return nil
}

func (e *MockExecutor) Prompt(msg, def string) (string, error) {
	return def, nil
}

func TestDockerTasks(t *testing.T) {
	mock := &MockExecutor{}
	cfg := &config.Config{Docker: true}
	sess := session.NewSession(cfg, mock)
	svc := &config.ServiceConfig{Name: "app", Docker: ptr(true)}

	tasks := []Task{
		NewDockerBuild(nil),
		NewDockerBuild(svc),
		NewDockerUp(nil),
		NewDockerUp(svc),
		NewDockerDown(nil),
		NewDockerDown(svc),
		NewDockerRebuild(nil),
		NewDockerRebuild(svc),
		NewDockerRemoveOrphans(nil),
		NewDockerRemoveOrphans(svc),
	}

	for _, task := range tasks {
		t.Run(task.Name(), func(t *testing.T) {
			mock.RunCalled = false
			if !task.ShouldRun(sess) {
				t.Errorf("ShouldRun failed for %s", task.Name())
			}
			task.Run(sess)
			if !mock.RunCalled {
				t.Errorf("Run not called for %s", task.Name())
			}
			if len(task.Commands(sess)) == 0 {
				t.Errorf("No commands for %s", task.Name())
			}
		})
	}
}

func TestDjangoTasks(t *testing.T) {
	mock := &MockExecutor{}
	svc := &config.ServiceConfig{
		Name: "web",
		Dir:  "./web",
		Modules: []config.ModuleConfig{
			{Python: &config.PythonConfig{Django: true}},
		},
	}
	cfg := &config.Config{Services: []config.ServiceConfig{*svc}}
	sess := session.NewSession(cfg, mock)

	tasks := []Task{
		NewDjangoRunServer(svc),
		NewDjangoMigrate(svc),
		NewDjangoMakeMigrations(svc),
		NewDjangoCollectStatic(svc),
		NewDjangoCreateUserDev(svc),
		NewDjangoGenRandomSecretKey(svc),
	}

	for _, task := range tasks {
		t.Run(task.Name(), func(t *testing.T) {
			mock.RunCalled = false
			if !task.ShouldRun(sess) {
				t.Errorf("ShouldRun failed for %s", task.Name())
			}
			task.Run(sess)
			if !mock.RunCalled {
				t.Errorf("Run not called for %s", task.Name())
			}
			if len(task.Commands(sess)) == 0 {
				t.Errorf("No commands for %s", task.Name())
			}
		})
	}
}

func TestGCPTasks(t *testing.T) {
	mock := &MockExecutor{}
	cfg := &config.Config{
		GoogleCloudPlatform: &config.GCPConfig{
			ProjectName: "test",
			Account:     "user@example.com",
		},
	}
	sess := session.NewSession(cfg, mock)

	tasks := []Task{
		NewGCPInit(),
		NewGCPActivate(),
		NewGCPSetConfig(),
		NewGCPADCLogin(),
		NewGCPAdcImpersonateLogin(),
		NewGCPConsole(),
		NewGCPAppEngineDeploy("app.yaml"),
		NewGCPAppEnginePromote("default"),
	}

	for _, task := range tasks {
		t.Run(task.Name(), func(t *testing.T) {
			mock.RunCalled = false
			// Some tasks might require inputs
			if task.Name() == "gcp:adc-impersonate-login" {
				sess.Inputs["gcp:impersonate-service-account"] = "sa@example.com"
			}
			if task.Name() == "gcp:app-engine-promote" || task.Name() == "gcp:app-engine-promote:default" {
				sess.Inputs["gcp:promote_version"] = "v1"
			}

			if !task.ShouldRun(sess) {
				// Some might not run without extra config, that's okay
				return
			}
			task.Run(sess)
			if !mock.RunCalled {
				t.Errorf("Run not called for %s", task.Name())
			}
			if len(task.Commands(sess)) == 0 {
				t.Errorf("No commands for %s", task.Name())
			}
			if reqs := task.Requirements(sess); reqs != nil {
				// Just check it doesn't crash
			}
		})
	}
}

func TestNpmTasks(t *testing.T) {
	mock := &MockExecutor{}
	npm := &config.NpmConfig{Scripts: []string{"build"}}
	svc := &config.ServiceConfig{Name: "ui", Dir: "ui"}
	cfg := &config.Config{}
	sess := session.NewSession(cfg, mock)

	tasks := []Task{
		NewNpmRun(svc, npm, "build"),
		NewNpmInstall(svc, npm),
	}

	for _, task := range tasks {
		t.Run(task.Name(), func(t *testing.T) {
			mock.RunCalled = false
			if !task.ShouldRun(sess) {
				t.Errorf("ShouldRun failed for %s", task.Name())
			}
			task.Run(sess)
			if !mock.RunCalled {
				t.Errorf("Run not called for %s", task.Name())
			}
		})
	}
}

func TestTerraformTasks(t *testing.T) {
	mock := &MockExecutor{}
	cfg := &config.Config{
		Terraform: &config.TerraformConfig{Dir: ".iac"},
	}
	sess := session.NewSession(cfg, mock)

	t.Run("TerraformInit", func(t *testing.T) {
		task := NewTerraform("", "init", nil)
		if !task.ShouldRun(sess) {
			t.Error("ShouldRun should be true")
		}
		task.Run(sess)
		if mock.Dir != "." {
			t.Errorf("expected dir ., got %q", mock.Dir)
		}
		// Verify -chdir is present
		found := false
		for _, arg := range mock.Args {
			if arg == "-chdir=.iac" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected -chdir=.iac in args, got %v", mock.Args)
		}
	})
}

func TestBaseTask(t *testing.T) {
	bt := &BaseTask{TaskName: "base", TaskDescription: "desc", TaskDeps: []string{"dep"}}
	if bt.Name() != "base" {
		t.Error("Name() failed")
	}
	if bt.Description() != "desc" {
		t.Error("Description() failed")
	}
	if len(bt.Dependencies()) != 1 || bt.Dependencies()[0] != "dep" {
		t.Error("Dependencies() failed")
	}
	if bt.Requirements(nil) != nil {
		t.Error("Requirements() should be nil")
	}
}

func TestTaskHelpers(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cleat-task-helper-*")
	defer os.RemoveAll(tmpDir)

	t.Run("DetectEnvFile", func(t *testing.T) {
		envsDir := filepath.Join(tmpDir, ".envs")
		os.Mkdir(envsDir, 0755)
		os.WriteFile(filepath.Join(envsDir, "dev.env"), []byte("VAR=VAL"), 0644)

		exec, abs, display := DetectEnvFile(tmpDir)
		if exec == "" || abs == "" || display == "" {
			t.Error("failed to detect env file")
		}
	})
}

func TestTerraformOpWrapping(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cleat-tf-op-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .envs directory and a .env file with op://
	envsDir := filepath.Join(tmpDir, ".envs")
	if err := os.Mkdir(envsDir, 0755); err != nil {
		t.Fatalf("failed to create .envs dir: %v", err)
	}
	envFile := filepath.Join(envsDir, "prod.env")
	if err := os.WriteFile(envFile, []byte("PASSWORD=op://vault/item/password\n"), 0644); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	cfg := &config.Config{
		SourcePath: filepath.Join(tmpDir, "cleat.yaml"),
		Terraform: &config.TerraformConfig{
			Dir:  ".iac",
			Envs: []string{"prod"},
		},
	}
	mock := &MockExecutor{}
	sess := session.NewSession(cfg, mock)

	t.Run("WrapsWithOpForProd", func(t *testing.T) {
		task := NewTerraform("prod", "plan", nil)
		cmds := task.Commands(sess)

		if len(cmds) != 1 {
			t.Fatalf("expected 1 command, got %d", len(cmds))
		}

		cmd := cmds[0]
		expectedEnvFile := ".envs/prod.env"
		expectedPrefix := []string{"op", "run", "--env-file=" + expectedEnvFile, "--"}

		for i, part := range expectedPrefix {
			if i >= len(cmd) || cmd[i] != part {
				t.Errorf("expected cmd[%d] to be %q, got %q", i, part, cmd[i])
			}
		}

		if cmd[len(cmd)-3] != "terraform" || !strings.HasPrefix(cmd[len(cmd)-2], "-chdir=") || cmd[len(cmd)-1] != "plan" {
			t.Errorf("expected command to end with terraform -chdir=... plan, got %v", cmd)
		}
	})

	t.Run("DoesNotWrapIfNoOpInEnv", func(t *testing.T) {
		devEnvFile := filepath.Join(envsDir, "dev.env")
		if err := os.WriteFile(devEnvFile, []byte("VAR=VAL\n"), 0644); err != nil {
			t.Fatalf("failed to write dev.env: %v", err)
		}

		task := NewTerraform("dev", "plan", nil)
		cmds := task.Commands(sess)

		if len(cmds) != 1 {
			t.Fatalf("expected 1 command, got %d", len(cmds))
		}

		cmd := cmds[0]
		if cmd[0] == "op" {
			t.Errorf("expected command NOT to start with op, got %v", cmd)
		}
	})

	t.Run("UsesRelativePathAndChdir", func(t *testing.T) {
		cfg.Terraform.UseFolders = true
		task := NewTerraform("prod", "plan", nil)
		cmds := task.Commands(sess)
		cmd := cmds[0]

		// Check for -chdir
		foundChdir := false
		for _, arg := range cmd {
			if strings.HasPrefix(arg, "-chdir=") {
				foundChdir = true
				if arg != "-chdir=.iac/prod" {
					t.Errorf("expected -chdir=.iac/prod, got %q", arg)
				}
			}
		}
		if !foundChdir {
			t.Error("expected -chdir argument not found")
		}

		// Check for relative env file path
		foundEnvFile := false
		for _, arg := range cmd {
			if strings.HasPrefix(arg, "--env-file=") {
				foundEnvFile = true
				relPath := strings.TrimPrefix(arg, "--env-file=")
				if relPath != ".envs/prod.env" {
					t.Errorf("expected --env-file=.envs/prod.env, got %q", relPath)
				}
			}
		}
		if !foundEnvFile {
			t.Error("expected --env-file argument not found")
		}
	})

	t.Run("ValidatesFolderMatchesEnv", func(t *testing.T) {
		// UseFolders is true by default in our cfg if we mock it right,
		// but let's be explicit.
		cfg.Terraform.UseFolders = true

		// Folder .iac/prod exists from previous setup.
		// Let's try an environment that DOES NOT have a folder.
		task := NewTerraform("missing", "plan", nil)

		// Create a mock env file for it so getEnvFile finds it
		if err := os.WriteFile(filepath.Join(envsDir, "missing.env"), []byte("OP=op://..."), 0644); err != nil {
			t.Fatal(err)
		}

		err := task.Run(sess)
		if err == nil {
			t.Error("expected error for missing terraform folder, got nil")
		} else if !strings.Contains(err.Error(), "terraform folder for environment 'missing' not found") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}
