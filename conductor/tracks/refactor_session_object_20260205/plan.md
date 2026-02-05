# Implementation Plan: Refactor Transient State to Session Object

Replace global transient state with an explicitly passed `Session` object.

## Phase 1: Session Definition and Schema Update [checkpoint: a66c2b4]
- [x] Task: Create `internal/session` package (537d6db)
    - [x] Define `Session` struct in `internal/session/session.go`
    - [x] Add constructor `NewSession(cfg *schema.Config, exec executor.Executor)`
- [x] Task: Remove global state from `internal/config` (e0a83d9)
    - [x] Delete `transientInputs` and `SetTransientInputs` from `internal/config/config.go`
- [x] Task: Conductor - User Manual Verification 'Phase 1: Session Definition and Schema Update' (Protocol in workflow.md) (a66c2b4)

## Phase 2: Core Interface Refactoring
- [x] Task: Refactor `internal/task` Interface (1e6bc6f)
    - [x] Update `Task` interface in `internal/task/task.go` to use `*session.Session`
    - [x] Update all `Task` implementations (Docker, NPM, Django, GCP, Terraform)
- [x] Task: Refactor `internal/strategy` Interface (1e6bc6f)
    - [x] Update `Strategy` interface and `BaseStrategy` in `internal/strategy/strategy.go`
    - [x] Update all `Strategy` implementations (Build, Run, NPM, Docker, etc.)
- [x] Task: Conductor - User Manual Verification 'Phase 2: Core Interface Refactoring' (Protocol in workflow.md) (1e6bc6f)

## Phase 3: Integration and Cleanup
- [x] Task: Update CLI Entry Points (1e6bc6f)
    - [x] Update commands in `internal/cmd` (run, build, etc.) to initialize and use the `Session`
- [x] Task: Update and Fix Tests (1e6bc6f)
    - [x] Update all unit and integration tests to use the `Session` object
    - [x] Verify that all tests pass and coverage is maintained
- [x] Task: Conductor - User Manual Verification 'Phase 3: Integration and Cleanup' (Protocol in workflow.md) (1e6bc6f)
