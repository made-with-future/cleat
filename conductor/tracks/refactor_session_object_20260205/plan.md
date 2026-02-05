# Implementation Plan: Refactor Transient State to Session Object

Replace global transient state with an explicitly passed `Session` object.

## Phase 1: Session Definition and Schema Update [checkpoint: a66c2b4]
- [x] Task: Create \`internal/session\` package (537d6db)
    - [x] Define \`Session\` struct in \`internal/session/session.go\`
    - [x] Add constructor \`NewSession(cfg *schema.Config, exec executor.Executor)\`
- [x] Task: Remove global state from \`internal/config\` (e0a83d9)
    - [x] Delete \`transientInputs\` and \`SetTransientInputs\` from \`internal/config/config.go\`
- [x] Task: Conductor - User Manual Verification 'Phase 1: Session Definition and Schema Update' (Protocol in workflow.md) (a66c2b4)

## Phase 2: Core Interface Refactoring
- [ ] Task: Refactor `internal/task` Interface
    - [ ] Update `Task` interface in `internal/task/task.go` to use `*session.Session`
    - [ ] Update all `Task` implementations (Docker, NPM, Django, GCP, Terraform)
- [ ] Task: Refactor `internal/strategy` Interface
    - [ ] Update `Strategy` interface and `BaseStrategy` in `internal/strategy/strategy.go`
    - [ ] Update all `Strategy` implementations (Build, Run, NPM, Docker, etc.)
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Core Interface Refactoring' (Protocol in workflow.md)

## Phase 3: Integration and Cleanup
- [ ] Task: Update CLI Entry Points
    - [ ] Update commands in `internal/cmd` (run, build, etc.) to initialize and use the `Session`
- [ ] Task: Update and Fix Tests
    - [ ] Update all unit and integration tests to use the `Session` object
    - [ ] Verify that all tests pass and coverage is maintained
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Integration and Cleanup' (Protocol in workflow.md)