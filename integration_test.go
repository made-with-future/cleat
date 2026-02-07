package cleat_test

import (
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/madewithfuture/cleat/testdata/testutil"
)

// TestSimpleDjangoFixture tests a basic Django project with Docker
func TestSimpleDjangoFixture(t *testing.T) {
	cfg := testutil.LoadFixture(t, "simple-django")

	// Verify config loaded correctly
	if !cfg.Docker {
		t.Error("expected Docker to be true")
	}

	// Verify services
	testutil.AssertServiceExists(t, cfg, "backend")
	testutil.AssertModuleExists(t, cfg, "backend", func(mod config.ModuleConfig) bool {
		return mod.Python != nil && mod.Python.Django
	})

	// Test build command
	mock := &testutil.MockExecutor{}
	sess := session.NewSession(cfg, mock)
	strat := strategy.GetStrategyForCommand("build", sess)
	if strat == nil {
		t.Fatal("expected build strategy")
	}

	tasks := strat.Tasks()
	// Verify docker:build exists
	foundDockerBuild := false
	foundCollectstatic := false
	for _, t := range tasks {
		if t.Name() == "docker:build" {
			foundDockerBuild = true
		}
		if t.Name() == "django:collectstatic" {
			foundCollectstatic = true
		}
	}
	if !foundDockerBuild {
		t.Error("expected docker:build task")
	}
	if !foundCollectstatic {
		t.Error("expected django:collectstatic task")
	}

	// Test run command
	strat = strategy.GetStrategyForCommand("run", sess)
	if strat == nil {
		t.Fatal("expected run strategy")
	}
	tasks = strat.Tasks()
	if len(tasks) == 0 {
		t.Error("expected run strategy to have tasks")
	}
}

// TestSimpleNpmFixture tests a basic NPM/frontend project without Docker
func TestSimpleNpmFixture(t *testing.T) {
	cfg := testutil.LoadFixture(t, "simple-npm")

	// Verify config
	if cfg.Docker {
		t.Error("expected Docker to be false")
	}

	testutil.AssertServiceExists(t, cfg, "frontend")
	testutil.AssertModuleExists(t, cfg, "frontend", func(mod config.ModuleConfig) bool {
		return mod.Npm != nil && len(mod.Npm.Scripts) > 0
	})

	// Test build command
	mock := &testutil.MockExecutor{}
	sess := session.NewSession(cfg, mock)
	strat := strategy.GetStrategyForCommand("build", sess)
	if strat == nil {
		t.Fatal("expected build strategy")
	}

	tasks := strat.Tasks()
	// Verify npm:run:build exists (docker:build may also be present as fallback)
	foundNpmBuild := false
	for _, t := range tasks {
		if t.Name() == "npm:run:build" {
			foundNpmBuild = true
			break
		}
	}
	if !foundNpmBuild {
		t.Error("expected npm:run:build task")
	}

	// Test npm run commands
	strat = strategy.GetStrategyForCommand("npm run test", sess)
	if strat == nil {
		t.Fatal("expected npm test strategy")
	}
	if strat.Name() != "npm:test" {
		t.Errorf("expected strategy name 'npm:test', got %q", strat.Name())
	}
}

// TestDjangoWithNpmFixture tests a project with both Django and NPM modules
func TestDjangoWithNpmFixture(t *testing.T) {
	cfg := testutil.LoadFixture(t, "django-with-npm")

	// Verify both modules exist
	testutil.AssertModuleExists(t, cfg, "backend", func(mod config.ModuleConfig) bool {
		return mod.Python != nil && mod.Python.Django
	})
	testutil.AssertModuleExists(t, cfg, "backend", func(mod config.ModuleConfig) bool {
		return mod.Npm != nil
	})

	// Test build command includes both
	mock := &testutil.MockExecutor{}
	sess := session.NewSession(cfg, mock)
	strat := strategy.GetStrategyForCommand("build", sess)
	if strat == nil {
		t.Fatal("expected build strategy")
	}

	tasks := strat.Tasks()
	// Verify expected tasks exist (may have duplicates due to multiple services)
	taskNames := make(map[string]bool)
	for _, t := range tasks {
		taskNames[t.Name()] = true
	}
	expectedTasks := []string{"docker:build", "npm:run:build", "django:collectstatic"}
	for _, expected := range expectedTasks {
		if !taskNames[expected] {
			t.Errorf("expected task %q not found", expected)
		}
	}
}

// TestMultiServiceFixture tests a multi-service project
func TestMultiServiceFixture(t *testing.T) {
	cfg := testutil.LoadFixture(t, "multi-service")

	// Verify all services exist
	testutil.AssertServiceExists(t, cfg, "backend")
	testutil.AssertServiceExists(t, cfg, "frontend")
	testutil.AssertServiceExists(t, cfg, "worker")

	// Verify Docker services from docker-compose.yaml
	if len(cfg.Services) < 3 {
		t.Errorf("expected at least 3 services, got %d", len(cfg.Services))
	}

	// Test build command
	mock := &testutil.MockExecutor{}
	sess := session.NewSession(cfg, mock)
	strat := strategy.GetStrategyForCommand("build", sess)
	if strat == nil {
		t.Fatal("expected build strategy")
	}

	tasks := strat.Tasks()
	if len(tasks) == 0 {
		t.Error("expected build strategy to have tasks")
	}

	// Verify docker:build task exists
	foundDockerBuild := false
	for _, task := range tasks {
		if task.Name() == "docker:build" {
			foundDockerBuild = true
			break
		}
	}
	if !foundDockerBuild {
		t.Error("expected docker:build task in multi-service build")
	}

	// Test service-specific commands
	strat = strategy.GetStrategyForCommand("docker down:backend", sess)
	if strat == nil {
		t.Fatal("expected docker down:backend strategy")
	}
}

// TestTerraformSimpleFixture tests Terraform with a single environment
func TestTerraformSimpleFixture(t *testing.T) {
	cfg := testutil.LoadFixture(t, "terraform-simple")

	// Verify terraform config
	if cfg.Terraform == nil {
		t.Fatal("expected Terraform config to exist")
	}
	if !cfg.Terraform.UseFolders {
		t.Error("expected UseFolders to be true")
	}

	// Verify envs
	if len(cfg.Envs) != 1 || cfg.Envs[0] != "production" {
		t.Errorf("expected envs [production], got %v", cfg.Envs)
	}

	// Test terraform commands
	mock := &testutil.MockExecutor{}
	sess := session.NewSession(cfg, mock)

	strat := strategy.GetStrategyForCommand("terraform plan:production", sess)
	if strat == nil {
		t.Fatal("expected terraform plan:production strategy")
	}
	if strat.Name() != "terraform:plan:production" {
		t.Errorf("expected strategy name 'terraform:plan:production', got %q", strat.Name())
	}

	strat = strategy.GetStrategyForCommand("terraform apply:production", sess)
	if strat == nil {
		t.Fatal("expected terraform apply:production strategy")
	}
}

// TestTerraformMultiEnvFixture tests Terraform with multiple environments
func TestTerraformMultiEnvFixture(t *testing.T) {
	cfg := testutil.LoadFixture(t, "terraform-multi-env")

	// Verify all envs
	expectedEnvs := []string{"production", "staging", "dev"}
	if len(cfg.Envs) != len(expectedEnvs) {
		t.Errorf("expected %d envs, got %d", len(expectedEnvs), len(cfg.Envs))
	}

	for _, env := range expectedEnvs {
		found := false
		for _, cfgEnv := range cfg.Envs {
			if cfgEnv == env {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected env %q not found in config", env)
		}
	}

	// Test terraform commands for each env
	mock := &testutil.MockExecutor{}
	sess := session.NewSession(cfg, mock)

	for _, env := range expectedEnvs {
		strat := strategy.GetStrategyForCommand("terraform plan:"+env, sess)
		if strat == nil {
			t.Errorf("expected terraform plan:%s strategy", env)
		}
	}
}

// TestGCPAppEngineFixture tests GCP App Engine configuration
func TestGCPAppEngineFixture(t *testing.T) {
	cfg := testutil.LoadFixture(t, "gcp-app-engine")

	// Verify GCP config
	if cfg.GoogleCloudPlatform == nil {
		t.Fatal("expected GCP config to exist")
	}
	if cfg.GoogleCloudPlatform.ProjectName != "my-gcp-project" {
		t.Errorf("expected project_name 'my-gcp-project', got %q", cfg.GoogleCloudPlatform.ProjectName)
	}

	// Verify app.yaml files detected
	if cfg.AppYaml == "" {
		t.Error("expected root app.yaml to be detected")
	}

	// Verify services with app.yaml
	foundBackend := false
	for _, svc := range cfg.Services {
		if svc.Name == "backend" && svc.AppYaml != "" {
			foundBackend = true
			break
		}
	}
	if !foundBackend {
		t.Error("expected backend service with app.yaml")
	}

	// Test GCP commands
	mock := &testutil.MockExecutor{}
	sess := session.NewSession(cfg, mock)

	strat := strategy.GetStrategyForCommand("gcp init", sess)
	if strat == nil {
		t.Fatal("expected gcp init strategy")
	}

	strat = strategy.GetStrategyForCommand("gcp app-engine deploy", sess)
	if strat == nil {
		t.Fatal("expected gcp app-engine deploy strategy")
	}
}

// TestDockerComposeOnlyFixture tests a project with only docker-compose, no modules
func TestDockerComposeOnlyFixture(t *testing.T) {
	cfg := testutil.LoadFixture(t, "docker-compose-only")

	// Verify Docker is enabled
	if !cfg.Docker {
		t.Error("expected Docker to be true")
	}

	// Verify services from docker-compose
	if len(cfg.Services) == 0 {
		t.Error("expected services to be detected from docker-compose.yaml")
	}

	// Test docker commands work
	mock := &testutil.MockExecutor{}
	sess := session.NewSession(cfg, mock)

	strat := strategy.GetStrategyForCommand("docker up", sess)
	if strat == nil {
		t.Fatal("expected docker up strategy")
	}

	strat = strategy.GetStrategyForCommand("docker down", sess)
	if strat == nil {
		t.Fatal("expected docker down strategy")
	}
}

// TestNoConfigAutoDetection tests auto-detection without cleat.yaml
func TestNoConfigAutoDetection(t *testing.T) {
	cfg := testutil.LoadFixture(t, "no-config")

	// Verify auto-detection worked
	if !cfg.Docker {
		t.Error("expected Docker to be auto-detected")
	}

	// Verify Django detected
	djangoFound := false
	for _, svc := range cfg.Services {
		for _, mod := range svc.Modules {
			if mod.Python != nil && mod.Python.Django {
				djangoFound = true
				break
			}
		}
	}
	if !djangoFound {
		t.Error("expected Django to be auto-detected")
	}

	// Verify NPM detected
	npmFound := false
	for _, svc := range cfg.Services {
		for _, mod := range svc.Modules {
			if mod.Npm != nil {
				npmFound = true
				break
			}
		}
	}
	if !npmFound {
		t.Error("expected NPM to be auto-detected")
	}
}

// TestComplexMonorepoFixture tests a complex project with all features
func TestComplexMonorepoFixture(t *testing.T) {
	cfg := testutil.LoadFixture(t, "complex-monorepo")

	// Verify all major features are configured
	if !cfg.Docker {
		t.Error("expected Docker to be true")
	}
	if cfg.Terraform == nil {
		t.Error("expected Terraform config")
	}
	if cfg.GoogleCloudPlatform == nil {
		t.Error("expected GCP config")
	}
	if len(cfg.Workflows) == 0 {
		t.Error("expected workflows to be configured")
	}

	// Verify envs
	if len(cfg.Envs) != 2 {
		t.Errorf("expected 2 envs, got %d", len(cfg.Envs))
	}

	// Verify services
	testutil.AssertServiceExists(t, cfg, "backend")
	testutil.AssertServiceExists(t, cfg, "frontend")
	testutil.AssertServiceExists(t, cfg, "worker")

	// Test workflow commands
	mock := &testutil.MockExecutor{}
	sess := session.NewSession(cfg, mock)

	strat := strategy.GetStrategyForCommand("workflow:full-deploy", sess)
	if strat == nil {
		t.Fatal("expected workflow:full-deploy strategy")
	}

	strat = strategy.GetStrategyForCommand("workflow:local-dev", sess)
	if strat == nil {
		t.Fatal("expected workflow:local-dev strategy")
	}

	// Test that regular commands still work
	strat = strategy.GetStrategyForCommand("build", sess)
	if strat == nil {
		t.Fatal("expected build strategy")
	}

	strat = strategy.GetStrategyForCommand("terraform plan:production", sess)
	if strat == nil {
		t.Fatal("expected terraform plan:production strategy")
	}

	strat = strategy.GetStrategyForCommand("gcp app-engine deploy", sess)
	if strat == nil {
		t.Fatal("expected gcp app-engine deploy strategy")
	}
}

// TestCommandResolution tests command resolution across different fixtures
func TestCommandResolution(t *testing.T) {
	tests := []struct {
		fixture     string
		command     string
		expectError bool
	}{
		{"simple-django", "build", false},
		{"simple-django", "run", false},
		{"simple-django", "django migrate", false},
		{"simple-django", "docker up", false},
		{"simple-npm", "build", false},
		{"simple-npm", "npm run build", false},
		{"simple-npm", "npm run test", false},
		{"multi-service", "build", false},
		{"multi-service", "docker down:backend", false},
		{"multi-service", "django migrate:backend", false},
		{"terraform-simple", "terraform plan:production", false},
		{"terraform-multi-env", "terraform apply:staging", false},
		{"gcp-app-engine", "gcp init", false},
		{"gcp-app-engine", "gcp app-engine deploy", false},
	}

	for _, tt := range tests {
		t.Run(tt.fixture+"/"+tt.command, func(t *testing.T) {
			cfg := testutil.LoadFixture(t, tt.fixture)
			mock := &testutil.MockExecutor{}
			sess := session.NewSession(cfg, mock)

			strat := strategy.GetStrategyForCommand(tt.command, sess)
			if tt.expectError {
				if strat != nil {
					t.Errorf("expected no strategy for %q, but got one", tt.command)
				}
			} else {
				if strat == nil {
					t.Errorf("expected strategy for %q, but got nil", tt.command)
				}
			}
		})
	}
}

// TestTaskDependencies tests that task dependencies are properly resolved
func TestTaskDependencies(t *testing.T) {
	cfg := testutil.LoadFixture(t, "simple-django")
	mock := &testutil.MockExecutor{}
	sess := session.NewSession(cfg, mock)

	strat := strategy.GetStrategyForCommand("build", sess)
	if strat == nil {
		t.Fatal("expected build strategy")
	}

	// Execute the strategy
	err := strat.Execute(sess)
	if err != nil {
		t.Fatalf("expected no error executing strategy, got: %v", err)
	}

	// Verify commands were executed (mock doesn't actually run them)
	// This tests that the execution flow works without errors
}

// TestServiceIsolation tests that service-specific commands target the right service
func TestServiceIsolation(t *testing.T) {
	cfg := testutil.LoadFixture(t, "multi-service")
	mock := &testutil.MockExecutor{}
	sess := session.NewSession(cfg, mock)

	// Test service-specific docker command
	strat := strategy.GetStrategyForCommand("docker down:backend", sess)
	if strat == nil {
		t.Fatal("expected docker down:backend strategy")
	}

	tasks := strat.Tasks()
	if len(tasks) != 1 {
		t.Errorf("expected 1 task for service-specific command, got %d", len(tasks))
	}

	// Verify task name includes service
	if len(tasks) > 0 && tasks[0].Name() != "docker:down:backend" {
		t.Errorf("expected task name 'docker:down:backend', got %q", tasks[0].Name())
	}
}

// TestDockerUpCommand tests the docker up command across different fixtures
func TestDockerUpCommand(t *testing.T) {
	tests := []struct {
		name          string
		fixture       string
		command       string
		expectSuccess bool
		expectedTasks []string
	}{
		{
			name:          "docker up on simple-django",
			fixture:       "simple-django",
			command:       "docker up",
			expectSuccess: true,
			expectedTasks: []string{"docker:up"},
		},
		{
			name:          "docker up on multi-service",
			fixture:       "multi-service",
			command:       "docker up",
			expectSuccess: true,
			expectedTasks: []string{"docker:up"},
		},
		{
			name:          "docker up on docker-compose-only",
			fixture:       "docker-compose-only",
			command:       "docker up",
			expectSuccess: true,
			expectedTasks: []string{"docker:up"},
		},
		{
			name:          "docker up on no-config (auto-detected)",
			fixture:       "no-config",
			command:       "docker up",
			expectSuccess: true,
			expectedTasks: []string{"docker:up"},
		},
		{
			name:          "docker up on complex-monorepo",
			fixture:       "complex-monorepo",
			command:       "docker up",
			expectSuccess: true,
			expectedTasks: []string{"docker:up"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testutil.LoadFixture(t, tt.fixture)
			mock := &testutil.MockExecutor{}
			sess := session.NewSession(cfg, mock)

			strat := strategy.GetStrategyForCommand(tt.command, sess)
			if tt.expectSuccess {
				if strat == nil {
					t.Fatalf("expected strategy for %q, got nil", tt.command)
				}

				// Verify task names
				tasks := strat.Tasks()
				taskNames := make([]string, len(tasks))
				for i, task := range tasks {
					taskNames[i] = task.Name()
				}

				if len(taskNames) != len(tt.expectedTasks) {
					t.Errorf("expected %d tasks, got %d: %v", len(tt.expectedTasks), len(taskNames), taskNames)
					return
				}

				for i, expected := range tt.expectedTasks {
					if taskNames[i] != expected {
						t.Errorf("task[%d]: expected %q, got %q", i, expected, taskNames[i])
					}
				}

				// Test execution
				err := strat.Execute(sess)
				if err != nil {
					t.Errorf("unexpected error executing strategy: %v", err)
				}

				// Verify docker command was executed
				if len(mock.Commands) == 0 {
					t.Error("expected commands to be executed")
				} else {
					foundDocker := false
					for _, cmd := range mock.Commands {
						if cmd.Name == "docker" || cmd.Name == "op" {
							foundDocker = true
							break
						}
					}
					if !foundDocker {
						t.Error("expected docker command to be executed")
					}
				}
			} else {
				if strat != nil {
					t.Errorf("expected no strategy for %q, but got one", tt.command)
				}
			}
		})
	}
}

// TestDockerUpWithServiceArgument tests docker up with specific service
func TestDockerUpWithServiceArgument(t *testing.T) {
	cfg := testutil.LoadFixture(t, "multi-service")
	mock := &testutil.MockExecutor{}
	sess := session.NewSession(cfg, mock)

	// Test service-specific docker up
	strat := strategy.GetStrategyForCommand("docker up:backend", sess)
	if strat == nil {
		t.Fatal("expected docker up:backend strategy")
	}

	tasks := strat.Tasks()
	if len(tasks) != 1 {
		t.Errorf("expected 1 task for service-specific command, got %d", len(tasks))
	}

	// Verify task name includes service
	if len(tasks) > 0 && tasks[0].Name() != "docker:up:backend" {
		t.Errorf("expected task name 'docker:up:backend', got %q", tasks[0].Name())
	}

	// Test execution
	err := strat.Execute(sess)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify command was executed
	if len(mock.Commands) == 0 {
		t.Error("expected commands to be executed")
	}
}

// TestDockerUpCommandOutput tests that docker up executes with correct arguments
func TestDockerUpCommandOutput(t *testing.T) {
	cfg := testutil.LoadFixture(t, "simple-django")
	mock := &testutil.MockExecutor{}
	sess := session.NewSession(cfg, mock)

	strat := strategy.GetStrategyForCommand("docker up", sess)
	if strat == nil {
		t.Fatal("expected docker up strategy")
	}

	err := strat.Execute(sess)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify docker compose up was called with correct args
	if len(mock.Commands) == 0 {
		t.Fatal("expected at least one command to be executed")
	}

	cmd := mock.Commands[0]
	if cmd.Name != "docker" && cmd.Name != "op" {
		t.Errorf("expected command to be 'docker' or 'op', got %q", cmd.Name)
	}

	// Check for compose up in arguments
	foundCompose := false
	foundUp := false
	foundRemoveOrphans := false

	for _, arg := range cmd.Args {
		if arg == "compose" {
			foundCompose = true
		}
		if arg == "up" {
			foundUp = true
		}
		if arg == "--remove-orphans" {
			foundRemoveOrphans = true
		}
	}

	if !foundCompose {
		t.Error("expected 'compose' in docker arguments")
	}
	if !foundUp {
		t.Error("expected 'up' in docker arguments")
	}
	if !foundRemoveOrphans {
		t.Error("expected '--remove-orphans' in docker arguments")
	}
}
