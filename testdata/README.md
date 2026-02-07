# Cleat Integration Test Fixtures

This directory contains test fixtures and integration tests for Cleat. Each fixture represents a different project layout that Cleat should support.

## Directory Structure

```
testdata/
├── fixtures/               # Test project layouts
│   ├── simple-django/      # Basic Django project with Docker
│   ├── simple-npm/         # Basic NPM/frontend project
│   ├── django-with-npm/    # Django + NPM in one service
│   ├── multi-service/      # Multiple services (backend, frontend, worker)
│   ├── terraform-simple/   # Terraform with single environment
│   ├── terraform-multi-env/ # Terraform with multiple environments
│   ├── gcp-app-engine/     # GCP App Engine with app.yaml files
│   ├── docker-compose-only/ # Just Docker Compose, no modules
│   ├── no-config/          # Auto-detection without cleat.yaml
│   └── complex-monorepo/   # All features combined
├── testutil/               # Helper functions for tests
│   └── testutil.go         # Test utilities and mock executor
└── README.md               # This file

integration_test.go         # Integration test suite (in project root)
```

## Fixtures

### simple-django
A basic Django project with Docker Compose. Tests:
- Docker detection
- Django module detection
- Build command with collectstatic
- Run command with Docker

**Files**: `cleat.yaml`, `docker-compose.yaml`, `manage.py`, `requirements.txt`

### simple-npm
A basic NPM/frontend project without Docker. Tests:
- NPM module detection
- Build command with NPM scripts
- npm run commands
- Non-Docker projects

**Files**: `cleat.yaml`, `package.json`, `vite.config.js`

### django-with-npm
A project with both Django backend and NPM frontend assets in one service. Tests:
- Multiple modules in one service
- Build command includes both NPM and Django tasks
- Docker services for both Python and Node

**Files**: `cleat.yaml`, `docker-compose.yaml`, `manage.py`, `package.json`

### multi-service
A monorepo with multiple services (backend, frontend, worker). Tests:
- Multiple services detection
- Service-specific commands (e.g., `docker down:backend`)
- Complex Docker Compose setups
- Different modules per service

**Files**: `cleat.yaml`, `docker-compose.yaml`, `backend/manage.py`, `frontend/package.json`

### terraform-simple
Terraform configuration with a single production environment. Tests:
- Terraform detection
- Single environment setup
- `use_folders` configuration
- Terraform commands (plan, apply, etc.)

**Files**: `cleat.yaml`, `.iac/production/main.tf`

### terraform-multi-env
Terraform configuration with multiple environments (production, staging, dev). Tests:
- Multiple environment detection
- Environment-specific commands
- Terraform folder structure

**Files**: `cleat.yaml`, `.iac/{production,staging,dev}/main.tf`

### gcp-app-engine
Google Cloud Platform App Engine configuration. Tests:
- GCP configuration detection
- app.yaml file detection (root and per-service)
- GCP commands (init, deploy, etc.)
- Multi-service GCP deployments

**Files**: `cleat.yaml`, `app.yaml`, `manage.py`, `backend/app.yaml`, `backend/main.py`

### docker-compose-only
A project with only Docker Compose services, no Python/NPM modules. Tests:
- Docker-only projects
- Service detection from docker-compose.yaml
- Docker commands without build tasks

**Files**: `docker-compose.yaml`

### no-config
A project without `cleat.yaml` that relies on auto-detection. Tests:
- Auto-detection of Docker
- Auto-detection of Django
- Auto-detection of NPM
- Default configuration generation

**Files**: `docker-compose.yaml`, `manage.py`, `package.json`

### complex-monorepo
A complex project with all features: multiple services, Terraform, GCP, workflows. Tests:
- All features combined
- Workflows
- Complex service interactions
- Environment files (.envs/)
- Multi-environment deployments

**Files**: Multiple directories with full stack setup

## Running Tests

From the project root:

```bash
# Run all tests (includes integration tests)
make test

# Run all tests with verbose output
go test ./... -v

# Run only integration tests
go test -v -run TestSimple

# Run a specific test
go test -v -run TestSimpleDjangoFixture

# Run tests with coverage
go test -cover
```

## Test Utilities

The `testutil` package provides helper functions for writing tests:

### Loading Fixtures

```go
// Load a fixture's configuration
cfg := testutil.LoadFixture(t, "simple-django")

// Copy a fixture to a temp directory (for tests that modify files)
tmpDir := testutil.CopyFixture(t, "simple-django")
defer os.RemoveAll(tmpDir)
```

### Mock Executor

```go
// Create a mock executor that records commands
mock := &testutil.MockExecutor{}
sess := session.NewSession(cfg, mock)

// Execute a strategy
strat := strategy.GetStrategyForCommand("build", sess)
err := strat.Execute(sess)

// Check what commands would have been executed
testutil.AssertCommandExecuted(t, mock, "docker")
```

### Assertions

```go
// Assert a service exists
testutil.AssertServiceExists(t, cfg, "backend")

// Assert a module exists
testutil.AssertModuleExists(t, cfg, "backend", func(mod config.ModuleConfig) bool {
    return mod.Python != nil && mod.Python.Django
})

// Assert task names match expected
testutil.AssertTaskNames(t, tasks, []string{"docker:build", "django:migrate"})
```

## Adding New Fixtures

To add a new fixture:

1. Create a new directory in `fixtures/`
2. Add realistic project files (cleat.yaml, docker-compose.yaml, etc.)
3. Add a test in `integration_test.go` that verifies the fixture loads correctly
4. Document the fixture in this README

Example structure:

```
fixtures/my-new-fixture/
├── cleat.yaml
├── docker-compose.yaml (if needed)
├── manage.py (if Django)
├── package.json (if NPM)
└── [other relevant files]
```

## Test Categories

The integration test suite covers:

- **Config Loading**: Verify each fixture loads correctly
- **Auto-detection**: Test detection of Docker, Django, NPM, Terraform, GCP
- **Command Resolution**: Test build, run, terraform, gcp, npm commands
- **Service Isolation**: Test per-service commands (e.g., `docker down:service`)
- **Workflows**: Test custom workflow execution
- **Task Dependencies**: Verify tasks execute in correct order
- **Edge Cases**: Invalid configs, missing files, circular dependencies

## Benefits

- **Fast**: Mock executor means no actual docker/npm/django execution
- **Reproducible**: Fixed project layouts ensure consistent test results
- **Comprehensive**: Covers all major Cleat features
- **Maintainable**: Easy to add new fixtures and test cases
- **Documentation**: Fixtures serve as examples of supported configurations
