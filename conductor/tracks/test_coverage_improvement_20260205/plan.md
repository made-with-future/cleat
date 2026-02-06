# Implementation Plan: Project-wide Test Coverage Audit and Improvement

Increase project-wide code coverage to >= 75%, ensuring individual modules hit >= 75% and individual functions hit >= 50%.

## Phase 1: Audit and Baseline [checkpoint: 8317f09]
- [x] Task: Generate baseline coverage report
    - [x] Run `go test -coverprofile=coverage.out ./...`
    - [x] Identify all modules and functions below the targets (75% module, 50% function)
    - [x] Create a "hit list" of low-coverage areas in `internal/ui` and `internal/cmd`
- [x] Task: Conductor - User Manual Verification 'Phase 1: Audit and Baseline' (Protocol in workflow.md)

## Phase 2: Core Logic and Utility Strengthening [checkpoint: d3d46e0]
- [x] Task: Improve coverage for `internal/detector`
    - [x] Write failing tests for auto-detection edge cases (Red Phase)
    - [x] Implement/Refactor to pass tests (Green Phase)
    - [x] Target: >= 75% coverage for `internal/detector`
- [x] Task: Improve coverage for `internal/executor` and `internal/logger`
    - [x] Write failing tests for error paths and side effects (Red Phase)
    - [x] Implement/Refactor to pass tests (Green Phase)
    - [x] Target: >= 75% coverage for these modules
- [x] Task: Conductor - User Manual Verification 'Phase 2: Core Logic and Utility Strengthening' (Protocol in workflow.md)

## Phase 3: Command and Strategy Refactoring
- [x] Task: Refactor and Test `internal/cmd`
    - [x] Identify complex `RunE` blocks in `internal/cmd/*.go`
    - [x] Extract logic into testable helper functions or methods
    - [x] Write failing tests for command routing and argument handling (Red Phase)
    - [x] Implement/Refactor to pass tests (Green Phase)
    - [x] Target: >= 75% coverage for `internal/cmd`
- [x] Task: Conductor - User Manual Verification 'Phase 3: Command and Strategy Refactoring' (Protocol in workflow.md)

## Phase 4: TUI Robustness
- [ ] Task: Improve coverage for `internal/ui`
    - [ ] Identify complex event handlers in `internal/ui/events.go` and `internal/ui/model.go`
    - [ ] Extract state transition logic from TUI rendering loops
    - [ ] Write failing tests for keyboard shortcuts and state transitions (Red Phase)
    - [ ] Implement/Refactor to pass tests (Green Phase)
    - [ ] Target: >= 75% coverage for `internal/ui`
- [ ] Task: Conductor - User Manual Verification 'Phase 4: TUI Robustness' (Protocol in workflow.md)

## Phase 5: Final Validation
- [ ] Task: Final Coverage Sweep
    - [ ] Run final project-wide coverage report
    - [ ] Verify no function is below 50% coverage
    - [ ] Verify overall coverage is >= 75%
- [ ] Task: Conductor - User Manual Verification 'Phase 5: Final Validation' (Protocol in workflow.md)
