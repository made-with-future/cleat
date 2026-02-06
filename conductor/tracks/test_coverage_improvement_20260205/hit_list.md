# Coverage Hit List (Baseline Audit)

Targets: Module >= 75%, Function >= 50%

## internal/ui (Current: 74.4%)
- `events.go`:
    - `handleInputCollection` (38.9%)
    - `handleShowingConfig` (30.8%)
    - `handleDownKey` (47.6%)
    - `handleEnterKey` (49.0%)
    - `openEditor` (0.0%)
- `model.go`:
    - `Init` (0.0%)
    - `paneWidths` (58.3%)
- `ui.go`:
    - `Start` (0.0%)

## internal/cmd (Current: 75.3%)
- `root.go`:
    - `Execute` (0.0%)
    - `init` (0.0%)
    - `waitForAnyKey` (0.0%)
- `terraform.go`:
    - `newTerraformSubcommand` (47.1%)

## Other Notable Low Coverage Functions
- `internal/detector/gcp.go: Detect` (47.6%)
- `internal/logger/logger.go: SetOutput` (0.0%)
- `internal/strategy/docker.go: NewDockerUpStrategyForService` (0.0%)
- `internal/strategy/gcp.go: NewGCPActivateStrategy` (0.0%)
- `internal/strategy/gcp.go: NewGCPSetConfigStrategy` (0.0%)
- `internal/strategy/gcp.go: NewGCPADCLoginStrategy` (0.0%)
- `internal/task/shell.go: NewShellTask`, `ShouldRun`, `Run`, `Commands` (0.0%)
- `internal/strategy/workflow.go: Name`, `Tasks` (0.0%)
