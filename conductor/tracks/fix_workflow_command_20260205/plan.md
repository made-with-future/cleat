# Implementation Plan: Restore Workflow Command Dispatching

Restore the missing `WorkflowProvider` to the strategy dispatcher to fix the regression where named workflows from `cleat.yaml` could not be executed.

## Phase 1: Test Setup & Command Resolution (Red Phase)
- [ ] Task: Write failing tests for workflow command resolution
    - [ ] Create test cases in `internal/strategy/strategy_test.go` that attempt to resolve `workflow:test-workflow`.
    - [ ] Define a sample workflow in the test configuration.
    - [ ] Verify that `GetStrategyForCommand` currently returns `nil` for these commands.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Test Setup & Command Resolution (Red Phase)' (Protocol in workflow.md)

## Phase 2: Implementation of WorkflowProvider (Green Phase)
- [ ] Task: Implement the `WorkflowProvider`
    - [ ] Define `WorkflowProvider` struct in `internal/strategy/strategy.go`.
    - [ ] Implement `CanHandle(command string)` to detect the `workflow:` prefix.
    - [ ] Implement `GetStrategy(command string, sess *session.Session)` to look up the workflow in `sess.Config.Workflows`.
    - [ ] Create a `WorkflowStrategy` that wraps the execution of multiple sub-commands.
- [ ] Task: Register the provider and pass tests
    - [ ] Add `&WorkflowProvider{}` to the `GetProviders()` slice.
    - [ ] Run tests and ensure the previously failing cases now pass.
    - [ ] Verify code coverage for the new provider is >80%.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Implementation of WorkflowProvider (Green Phase)' (Protocol in workflow.md)

## Phase 3: Advanced Execution & TUI Verification
- [ ] Task: Support dynamic inputs and error handling
    - [ ] Ensure `WorkflowStrategy` correctly bubbles up `InputRequirements` from its constituent tasks.
    - [ ] Verify that a failure in one step stops the workflow (Fail Fast).
- [ ] Task: Final verification and TUI check
    - [ ] Perform manual verification in the TUI to ensure workflows can be triggered without the "unknown command" error.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Advanced Execution & TUI Verification' (Protocol in workflow.md)
