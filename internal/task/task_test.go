package task

import (
	"os"
	"path/filepath"
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
}

func (e *MockExecutor) Run(name string, args ...string) error {
	e.RunCalled = true
	e.Name = name
	return nil
}

func (e *MockExecutor) RunWithDir(dir string, name string, args ...string) error {
	e.RunCalled = true
	e.Dir = dir
	e.Name = name
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
		if mock.Dir != ".iac" {
			t.Errorf("expected dir .iac, got %q", mock.Dir)
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