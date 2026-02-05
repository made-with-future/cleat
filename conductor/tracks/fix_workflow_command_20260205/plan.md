# Implementation Plan: Restore Workflow Command Dispatching

Restore the missing `WorkflowProvider` to the strategy dispatcher to fix the regression where named workflows from `cleat.yaml` could not be executed.

## Phase 1: Test Setup & Command Resolution (Red Phase) [checkpoint: c7cbf28]
- [x] Task: Write failing tests for workflow command resolution [dc36c9f]
    - [x] Create test cases in `internal/strategy/strategy_test.go` that attempt to resolve `workflow:test-workflow`.
    - [x] Define a sample workflow in the test configuration.
    - [x] Verify that `GetStrategyForCommand` currently returns `nil` for these commands.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Test Setup & Command Resolution (Red Phase)' (Protocol in workflow.md)

## Phase 2: Implementation of WorkflowProvider (Green Phase) [checkpoint: 63e6f63]
- [x] Task: Implement the `WorkflowProvider` [3be00d1]
    - [x] Define `WorkflowProvider` struct in `internal/strategy/strategy.go`.
    - [x] Implement `CanHandle(command string)` to detect the `workflow:` prefix.
    - [x] Implement `GetStrategy(command string, sess *session.Session)` to look up the workflow in `sess.Config.Workflows`.
    - [x] Create a `WorkflowStrategy` that wraps the execution of multiple sub-commands.
- [x] Task: Register the provider and pass tests [3be00d1]
    - [x] Add `&WorkflowProvider{}` to the `GetProviders()` slice.
    - [x] Run tests and ensure the previously failing cases now pass.
    - [x] Verify code coverage for the new provider is >80%.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Implementation of WorkflowProvider (Green Phase)' (Protocol in workflow.md)

## Phase 3: Advanced Execution & TUI Verification
- [x] Task: Support dynamic inputs and error handling [caa5bc3]
    - [x] Ensure `WorkflowStrategy` correctly bubbles up `InputRequirements` from its constituent tasks.
    - [x] Verify that a failure in one step stops the workflow (Fail Fast).
- [x] Task: Final verification and TUI check [d0ce479]
    - [x] Perform manual verification in the TUI to ensure workflows can be triggered without the "unknown command" error.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Advanced Execution & TUI Verification' (Protocol in workflow.md)
